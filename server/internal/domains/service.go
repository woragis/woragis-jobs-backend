package jobs

import (
	"time"

	"woragis-jobs-service/pkg/auth"
	"woragis-jobs-service/pkg/crypto"
	apperrors "woragis-jobs-service/pkg/errors"
	appmetrics "woragis-jobs-service/pkg/metrics"

	"github.com/google/uuid"
)

// Removed old error variables - now using structured error codes from pkg/errors

// Service defines the interface for auth service operations
type Service interface {
	register(req *RegisterRequest) (*AuthResponse, error)
	login(req *LoginRequest, userAgent, ipAddress string) (*AuthResponse, error)
	refreshToken(refreshToken string) (*AuthResponse, error)
	logout(refreshToken string) error
	logoutAll(userID uuid.UUID) error
	getUserProfile(userID uuid.UUID) (*Profile, error)
	updateUserProfile(userID uuid.UUID, req *ProfileUpdateRequest) (*Profile, error)
	changePassword(userID uuid.UUID, req *PasswordChangeRequest) error
	createVerificationToken(userID uuid.UUID, tokenType string) (*VerificationToken, error)
	verifyEmail(token string) error
	getUserByID(userID uuid.UUID) (*User, error)
	listUsers(page, limit int, search string) ([]User, int64, error)
	cleanupExpiredSessions() error
}

// serviceImpl implements the Service interface
type serviceImpl struct {
	repo       Repository
	jwtManager *auth.JWTManager
	bcryptCost int
}

// NewService creates a new auth service
func NewService(repo Repository, jwtManager *auth.JWTManager, bcryptCost int) Service {
	return &serviceImpl{
		repo:       repo,
		jwtManager: jwtManager,
		bcryptCost: bcryptCost,
	}
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=30"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=30"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// ProfileUpdateRequest represents profile update request
type ProfileUpdateRequest struct {
	Avatar      string     `json:"avatar"`
	Bio         string     `json:"bio"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      string     `json:"gender"`
	Phone       string     `json:"phone"`
	Location    string     `json:"location"`
	Website     string     `json:"website"`
	SocialLinks string     `json:"social_links"`
	Preferences string     `json:"preferences"`
}

// PasswordChangeRequest represents password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// register registers a new user
func (s *serviceImpl) register(req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.repo.getUserByEmail(req.Email)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); !ok || appErr.Code != apperrors.AUTH_USER_NOT_FOUND {
			return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
		}
	}
	if existingUser != nil {
		return nil, apperrors.New(apperrors.AUTH_EMAIL_ALREADY_EXISTS)
	}

	// Validate password strength
	if err := auth.CheckPasswordStrength(req.Password); err != nil {
		return nil, apperrors.New(apperrors.AUTH_PASSWORD_TOO_SHORT)
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Create user
	user := &User{
		ID:        uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user",
		IsActive:  true,
		IsVerified: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.createUser(user); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Record user registration metric
	appmetrics.RecordUserRegistration()

	// Create default profile
	profile := &Profile{
		ID:           uuid.New(),
		UserID:       user.ID,
		SocialLinks:  "{}", // Empty JSON object
		Preferences:  "{}", // Empty JSON object
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.createProfile(profile); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.Generate(user.ID, user.Email, user.Role, user.Username)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Create session
	session := &Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		IsActive:     true,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.createSession(session); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Update last login
	if err := s.repo.updateLastLogin(user.ID); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Load user with profile
	user, err = s.repo.getUserByID(user.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(), // 24 hours
	}, nil
}

// login authenticates a user
func (s *serviceImpl) login(req *LoginRequest, userAgent, ipAddress string) (*AuthResponse, error) {
	var user *User
	var err error

	// Validate that either email or username is provided
	if req.Email == "" && req.Username == "" {
		return nil, apperrors.New(apperrors.AUTH_INVALID_CREDENTIALS)
	}

	// Get user by email or username
	if req.Email != "" {
		user, err = s.repo.getUserByEmail(req.Email)
	} else {
		user, err = s.repo.getUserByUsername(req.Username)
	}

	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Code == apperrors.AUTH_USER_NOT_FOUND {
			return nil, apperrors.New(apperrors.AUTH_INVALID_CREDENTIALS)
		}
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperrors.New(apperrors.AUTH_UNAUTHORIZED).WithDetails("Account is inactive")
	}

	// Verify password
	if err := auth.VerifyPassword(req.Password, user.Password); err != nil {
		appmetrics.RecordUserLogin(false)
		return nil, apperrors.New(apperrors.AUTH_INVALID_CREDENTIALS)
	}

	// Generate tokens
	accessToken, refreshToken, err := s.jwtManager.Generate(user.ID, user.Email, user.Role, user.Username)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Create session
	session := &Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		IsActive:     true,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.createSession(session); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Update last login
	if err := s.repo.updateLastLogin(user.ID); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Record successful login
	appmetrics.RecordUserLogin(true)

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(), // 24 hours
	}, nil
}

// refreshToken refreshes an access token
func (s *serviceImpl) refreshToken(refreshToken string) (*AuthResponse, error) {
	// Get session by refresh token
	session, err := s.repo.getSessionByRefreshToken(refreshToken)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Code == apperrors.DB_RECORD_NOT_FOUND {
			return nil, apperrors.New(apperrors.CSRF_TOKEN_EXPIRED).WithDetails("Session expired")
		}
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Check if session is valid
	if !session.IsSessionValid() {
		return nil, apperrors.New(apperrors.CSRF_TOKEN_EXPIRED).WithDetails("Session expired")
	}

	// Get user
	user, err := s.repo.getUserByID(session.UserID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperrors.New(apperrors.AUTH_UNAUTHORIZED).WithDetails("Account is inactive")
	}

	// Generate new access token
	newAccessToken, err := s.jwtManager.Refresh(refreshToken)
	if err != nil {
		appmetrics.RecordTokenRefresh(false)
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Record successful token refresh
	appmetrics.RecordTokenRefresh(true)

	return &AuthResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(), // 24 hours
	}, nil
}

// logout logs out a user by deactivating their session
func (s *serviceImpl) logout(refreshToken string) error {
	session, err := s.repo.getSessionByRefreshToken(refreshToken)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Code == apperrors.DB_RECORD_NOT_FOUND {
			return nil // Already logged out
		}
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Revoke refresh token (add to blacklist)
	// Use refresh token expiry duration (typically 7 days)
	if err := s.jwtManager.RevokeToken(refreshToken, 7*24*time.Hour); err != nil {
		// Log error but continue with session deactivation
		// Token revocation is best-effort
	} else {
		// Record token revocation
		appmetrics.RecordTokenRevocation()
	}

	return s.repo.deactivateSession(session.ID)
}

// logoutAll logs out a user from all devices
func (s *serviceImpl) logoutAll(userID uuid.UUID) error {
	// Revoke all tokens for the user (add to blacklist)
	// Use refresh token expiry duration (typically 7 days)
	if err := s.jwtManager.RevokeUserTokens(userID, 7*24*time.Hour); err != nil {
		// Log error but continue with session deactivation
		// Token revocation is best-effort
	} else {
		// Record token revocation
		appmetrics.RecordTokenRevocation()
	}

	return s.repo.deactivateAllUserSessions(userID)
}

// getUserProfile retrieves user profile
func (s *serviceImpl) getUserProfile(userID uuid.UUID) (*Profile, error) {
	return s.repo.getProfileByUserID(userID)
}

// updateUserProfile updates user profile
func (s *serviceImpl) updateUserProfile(userID uuid.UUID, req *ProfileUpdateRequest) (*Profile, error) {
	profile, err := s.repo.getProfileByUserID(userID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Update profile fields
	if req.Avatar != "" {
		profile.Avatar = req.Avatar
	}
	if req.Bio != "" {
		profile.Bio = req.Bio
	}
	if req.DateOfBirth != nil {
		profile.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != "" {
		profile.Gender = req.Gender
	}
	if req.Phone != "" {
		profile.Phone = req.Phone
	}
	if req.Location != "" {
		profile.Location = req.Location
	}
	if req.Website != "" {
		profile.Website = req.Website
	}
	if req.SocialLinks != "" {
		profile.SocialLinks = req.SocialLinks
	}
	if req.Preferences != "" {
		profile.Preferences = req.Preferences
	}

	profile.UpdatedAt = time.Now()

	if err := s.repo.updateProfile(profile); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	return profile, nil
}

// changePassword changes user password
func (s *serviceImpl) changePassword(userID uuid.UUID, req *PasswordChangeRequest) error {
	// Get user
	user, err := s.repo.getUserByID(userID)
	if err != nil {
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Verify current password
	if err := auth.VerifyPassword(req.CurrentPassword, user.Password); err != nil {
		return apperrors.New(apperrors.AUTH_INVALID_CREDENTIALS).WithDetails("Current password is incorrect")
	}

	// Validate new password strength
	if err := auth.CheckPasswordStrength(req.NewPassword); err != nil {
		return apperrors.New(apperrors.AUTH_WEAK_PASSWORD)
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword, s.bcryptCost)
	if err != nil {
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Update password
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.repo.updateUser(user); err != nil {
		appmetrics.RecordPasswordChange(false)
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Record successful password change
	appmetrics.RecordPasswordChange(true)

	// Logout from all devices for security
	return s.logoutAll(userID)
}

// createVerificationToken creates a verification token for email verification
func (s *serviceImpl) createVerificationToken(userID uuid.UUID, tokenType string) (*VerificationToken, error) {
	// Generate random token
	token, err := crypto.GenerateRandomString(32)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	verificationToken := &VerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
		IsUsed:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.createVerificationToken(verificationToken); err != nil {
		return nil, apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	return verificationToken, nil
}

// verifyEmail verifies user email using verification token
func (s *serviceImpl) verifyEmail(token string) error {
	verificationToken, err := s.repo.getVerificationToken(token)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok && appErr.Code == apperrors.DB_RECORD_NOT_FOUND {
			return apperrors.New(apperrors.AUTH_TOKEN_EXPIRED)
		}
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Check if token is valid
	if !verificationToken.IsTokenValid() {
		if verificationToken.IsUsed {
			return apperrors.New(apperrors.AUTH_TOKEN_ALREADY_USED)
		}
		return apperrors.New(apperrors.AUTH_TOKEN_EXPIRED)
	}

	// Mark token as used
	if err := s.repo.markTokenAsUsed(verificationToken.ID); err != nil {
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Verify user email
	if err := s.repo.verifyUserEmail(verificationToken.UserID); err != nil {
		appmetrics.RecordEmailVerification(false)
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	// Record successful email verification
	appmetrics.RecordEmailVerification(true)

	return nil
}

// getUserByID retrieves a user by ID
func (s *serviceImpl) getUserByID(userID uuid.UUID) (*User, error) {
	return s.repo.getUserByID(userID)
}

// listUsers retrieves users with pagination
func (s *serviceImpl) listUsers(page, limit int, search string) ([]User, int64, error) {
	offset := (page - 1) * limit
	return s.repo.listUsers(offset, limit, search)
}

// cleanupExpiredSessions removes expired sessions and tokens
func (s *serviceImpl) cleanupExpiredSessions() error {
	if err := s.repo.deleteExpiredSessions(); err != nil {
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	if err := s.repo.deleteExpiredTokens(); err != nil {
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	if err := s.repo.deleteUsedTokens(); err != nil {
		return apperrors.Wrap(apperrors.DB_QUERY_FAILED, err)
	}

	return nil
}
