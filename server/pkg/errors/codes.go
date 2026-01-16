package errors

// Error code format: SERVICE_CATEGORY_NUMBER
// Each code should be used in exactly one place for easy tracking

const (
	// AUTH - Authentication errors (1000-1099)
	AUTH_JWT_INVALID_SIGNATURE    = "JOBS_1001"
	AUTH_JWT_EXPIRED              = "JOBS_1002"
	AUTH_JWT_MISSING_CLAIMS       = "JOBS_1003"
	AUTH_JWT_MALFORMED            = "JOBS_1004"
	AUTH_TOKEN_MISSING            = "JOBS_1020"
	AUTH_TOKEN_INVALID_FORMAT     = "JOBS_1021"
	AUTH_UNAUTHORIZED             = "JOBS_1022"
	AUTH_SERVICE_UNAVAILABLE      = "JOBS_1030"
	AUTH_SERVICE_ERROR            = "JOBS_1031"
	AUTH_USER_NOT_FOUND           = "JOBS_1040"
	AUTH_EMAIL_ALREADY_EXISTS     = "JOBS_1041"
	AUTH_USERNAME_ALREADY_EXISTS  = "JOBS_1042"
	AUTH_INVALID_CREDENTIALS      = "JOBS_1043"
	AUTH_PASSWORD_TOO_SHORT       = "JOBS_1044"
	AUTH_WEAK_PASSWORD            = "JOBS_1045"
	AUTH_TOKEN_EXPIRED            = "JOBS_1046"
	AUTH_TOKEN_ALREADY_USED       = "JOBS_1047"

	// CSRF - CSRF token errors (2000-2099)
	CSRF_TOKEN_EXPIRED            = "JOBS_2001"
	CSRF_TOKEN_INVALID            = "JOBS_2002"
	CSRF_TOKEN_MISSING            = "JOBS_2003"
	CSRF_TOKEN_GENERATION_FAILED  = "JOBS_2004"
	CSRF_TOKEN_MISMATCH           = "JOBS_2005"

	// DB - Database errors (3000-3099)
	DB_CONNECTION_FAILED          = "JOBS_3001"
	DB_QUERY_FAILED               = "JOBS_3002"
	DB_TRANSACTION_FAILED         = "JOBS_3003"
	DB_RECORD_NOT_FOUND           = "JOBS_3004"
	DB_DUPLICATE_ENTRY            = "JOBS_3005"
	DB_CONSTRAINT_VIOLATION       = "JOBS_3006"
	DB_MIGRATION_FAILED           = "JOBS_3007"

	// VALIDATION - Input validation errors (4000-4099)
	VALIDATION_INVALID_INPUT      = "JOBS_4001"
	VALIDATION_MISSING_FIELD      = "JOBS_4002"
	VALIDATION_FIELD_TOO_LONG     = "JOBS_4003"
	VALIDATION_FIELD_TOO_SHORT    = "JOBS_4004"
	VALIDATION_INVALID_UUID       = "JOBS_4005"
	VALIDATION_INVALID_DATE       = "JOBS_4006"
	VALIDATION_INVALID_URL        = "JOBS_4007"
	VALIDATION_INVALID_STATUS     = "JOBS_4008"

	// REDIS - Redis/Cache errors (5000-5099)
	REDIS_CONNECTION_FAILED       = "JOBS_5001"
	REDIS_GET_FAILED              = "JOBS_5002"
	REDIS_SET_FAILED              = "JOBS_5003"
	REDIS_DELETE_FAILED           = "JOBS_5004"

	// RABBITMQ - Message queue errors (6000-6099)
	RABBITMQ_CONNECTION_FAILED    = "JOBS_6001"
	RABBITMQ_CHANNEL_FAILED       = "JOBS_6002"
	RABBITMQ_PUBLISH_FAILED       = "JOBS_6003"
	RABBITMQ_CONSUME_FAILED       = "JOBS_6004"
	RABBITMQ_QUEUE_DECLARE_FAILED = "JOBS_6005"
	RABBITMQ_EXCHANGE_DECLARE_FAILED = "JOBS_6006"

	// JOB_APPLICATION - Job application specific errors (7000-7099)
	JOB_APP_NOT_FOUND             = "JOBS_7001"
	JOB_APP_INVALID_STATUS        = "JOBS_7002"
	JOB_APP_ALREADY_EXISTS        = "JOBS_7003"
	JOB_APP_UPDATE_FAILED         = "JOBS_7004"
	JOB_APP_DELETE_FAILED         = "JOBS_7005"

	// RESUME - Resume specific errors (7100-7199)
	RESUME_NOT_FOUND              = "JOBS_7101"
	RESUME_GENERATION_FAILED      = "JOBS_7102"
	RESUME_ALREADY_EXISTS         = "JOBS_7103"
	RESUME_JOB_NOT_FOUND          = "JOBS_7104"
	RESUME_JOB_FAILED             = "JOBS_7105"
	RESUME_INVALID_LANGUAGE       = "JOBS_7106"
	RESUME_SERVICE_ERROR          = "JOBS_7107"

	// COVER_LETTER - Cover letter specific errors (7200-7299)
	COVER_LETTER_NOT_FOUND        = "JOBS_7201"
	COVER_LETTER_GENERATION_FAILED = "JOBS_7202"
	COVER_LETTER_INVALID_TEMPLATE = "JOBS_7203"

	// AI_SERVICE - AI service errors (8000-8099)
	AI_SERVICE_UNAVAILABLE        = "JOBS_8001"
	AI_SERVICE_REQUEST_FAILED     = "JOBS_8002"
	AI_SERVICE_INVALID_RESPONSE   = "JOBS_8003"
	AI_SERVICE_TIMEOUT            = "JOBS_8004"

	// SERVER - Server/System errors (9000-9099)
	SERVER_INTERNAL_ERROR         = "JOBS_9001"
	SERVER_SERVICE_UNAVAILABLE    = "JOBS_9002"
	SERVER_TIMEOUT                = "JOBS_9003"
	SERVER_CONTEXT_CANCELLED      = "JOBS_9004"
	SERVER_RATE_LIMIT_EXCEEDED    = "JOBS_9005"
)

// Error messages - human-readable descriptions
var errorMessages = map[string]string{
	// Authentication
	AUTH_JWT_INVALID_SIGNATURE:    "JWT token signature is invalid",
	AUTH_JWT_EXPIRED:              "JWT token has expired",
	AUTH_JWT_MISSING_CLAIMS:       "JWT token is missing required claims",
	AUTH_JWT_MALFORMED:            "JWT token is malformed",
	AUTH_TOKEN_MISSING:            "Authentication token is missing",
	AUTH_TOKEN_INVALID_FORMAT:     "Authentication token has invalid format",
	AUTH_UNAUTHORIZED:             "Unauthorized access",
	AUTH_SERVICE_UNAVAILABLE:      "Authentication service is unavailable",
	AUTH_SERVICE_ERROR:            "Authentication service error",
	AUTH_USER_NOT_FOUND:           "User not found",
	AUTH_EMAIL_ALREADY_EXISTS:     "Email address is already registered",
	AUTH_USERNAME_ALREADY_EXISTS:  "Username is already taken",
	AUTH_INVALID_CREDENTIALS:      "Invalid email or password",
	AUTH_PASSWORD_TOO_SHORT:       "Password must be at least 8 characters",
	AUTH_WEAK_PASSWORD:            "Password is too weak",
	AUTH_TOKEN_EXPIRED:            "Verification token has expired",
	AUTH_TOKEN_ALREADY_USED:       "Verification token has already been used",

	// CSRF
	CSRF_TOKEN_EXPIRED:           "CSRF token has expired",
	CSRF_TOKEN_INVALID:           "CSRF token validation failed",
	CSRF_TOKEN_MISSING:           "CSRF token is missing from request",
	CSRF_TOKEN_GENERATION_FAILED: "Failed to generate CSRF token",
	CSRF_TOKEN_MISMATCH:          "CSRF token does not match stored value",

	// Database
	DB_CONNECTION_FAILED:    "Failed to connect to database",
	DB_QUERY_FAILED:         "Database query execution failed",
	DB_TRANSACTION_FAILED:   "Database transaction failed",
	DB_RECORD_NOT_FOUND:     "Requested record not found",
	DB_DUPLICATE_ENTRY:      "Record already exists",
	DB_CONSTRAINT_VIOLATION: "Database constraint violation",
	DB_MIGRATION_FAILED:     "Database migration failed",

	// Validation
	VALIDATION_INVALID_INPUT:  "Input validation failed",
	VALIDATION_MISSING_FIELD:  "Required field is missing",
	VALIDATION_FIELD_TOO_LONG: "Field value exceeds maximum length",
	VALIDATION_FIELD_TOO_SHORT: "Field value is below minimum length",
	VALIDATION_INVALID_UUID:   "Invalid UUID format",
	VALIDATION_INVALID_DATE:   "Invalid date format",
	VALIDATION_INVALID_URL:    "Invalid URL format",
	VALIDATION_INVALID_STATUS: "Invalid status value",

	// Redis
	REDIS_CONNECTION_FAILED: "Failed to connect to Redis",
	REDIS_GET_FAILED:        "Failed to retrieve data from cache",
	REDIS_SET_FAILED:        "Failed to store data in cache",
	REDIS_DELETE_FAILED:     "Failed to delete data from cache",

	// RabbitMQ
	RABBITMQ_CONNECTION_FAILED:       "Failed to connect to RabbitMQ",
	RABBITMQ_CHANNEL_FAILED:          "Failed to create RabbitMQ channel",
	RABBITMQ_PUBLISH_FAILED:          "Failed to publish message to queue",
	RABBITMQ_CONSUME_FAILED:          "Failed to consume messages from queue",
	RABBITMQ_QUEUE_DECLARE_FAILED:    "Failed to declare queue",
	RABBITMQ_EXCHANGE_DECLARE_FAILED: "Failed to declare exchange",

	// Job Applications
	JOB_APP_NOT_FOUND:      "Job application not found",
	JOB_APP_INVALID_STATUS: "Invalid job application status",
	JOB_APP_ALREADY_EXISTS: "Job application already exists",
	JOB_APP_UPDATE_FAILED:  "Failed to update job application",
	JOB_APP_DELETE_FAILED:  "Failed to delete job application",

	// Resumes
	RESUME_NOT_FOUND:           "Resume not found",
	RESUME_GENERATION_FAILED:   "Resume generation failed",
	RESUME_ALREADY_EXISTS:      "Resume already exists for this job application",
	RESUME_JOB_NOT_FOUND:       "Resume generation job not found",
	RESUME_JOB_FAILED:          "Resume generation job failed",
	RESUME_INVALID_LANGUAGE:    "Invalid resume language",
	RESUME_SERVICE_ERROR:       "Resume service error",

	// Cover Letters
	COVER_LETTER_NOT_FOUND:         "Cover letter not found",
	COVER_LETTER_GENERATION_FAILED: "Cover letter generation failed",
	COVER_LETTER_INVALID_TEMPLATE:  "Invalid cover letter template",

	// AI Service
	AI_SERVICE_UNAVAILABLE:     "AI service is unavailable",
	AI_SERVICE_REQUEST_FAILED:  "AI service request failed",
	AI_SERVICE_INVALID_RESPONSE: "AI service returned invalid response",
	AI_SERVICE_TIMEOUT:         "AI service request timeout",

	// Server
	SERVER_INTERNAL_ERROR:      "Internal server error occurred",
	SERVER_SERVICE_UNAVAILABLE: "Service is temporarily unavailable",
	SERVER_TIMEOUT:             "Request timeout",
	SERVER_CONTEXT_CANCELLED:   "Request was cancelled",
	SERVER_RATE_LIMIT_EXCEEDED: "Rate limit exceeded",
}

// GetMessage returns the human-readable message for an error code
func GetMessage(code string) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "Unknown error occurred"
}
