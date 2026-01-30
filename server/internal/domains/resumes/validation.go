package resumes

import (
	"fmt"
	"strings"

	"woragis-jobs-service/pkg/validation"
)

// createResumePayload represents the payload for CreateResume
type createResumePayload struct {
	Title    string   `json:"title"`
	FilePath string   `json:"filePath"`
	FileName string   `json:"fileName"`
	FileSize int64    `json:"fileSize"`
	Tags     []string `json:"tags"`
}

// updateResumePayload represents the payload for UpdateResume
type updateResumePayload struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

// generateResumePayload represents the payload for GenerateResume
type generateResumePayload struct {
	JobApplicationID string `json:"jobApplicationId"`
	JobDescription   string `json:"jobDescription"`
	Language         string `json:"language"`
	Template         string `json:"template,omitempty"`
}

// ValidateCreateResumePayload validates create resume payload
func ValidateCreateResumePayload(payload *createResumePayload) error {
	// Validate title (required, 1-200 chars)
	if err := validation.ValidateString(payload.Title, 1, 200, "title"); err != nil {
		return fmt.Errorf("title: %w", err)
	}
	// Check for SQL injection and XSS
	if err := validation.ValidateNoSQLInjection(payload.Title); err != nil {
		return fmt.Errorf("title: %w", err)
	}
	if err := validation.ValidateNoXSS(payload.Title); err != nil {
		return fmt.Errorf("title: %w", err)
	}

	// Validate file path (optional, but if provided, validate)
	if payload.FilePath != "" {
		if err := validation.ValidateString(payload.FilePath, 1, 500, "filePath"); err != nil {
			return fmt.Errorf("filePath: %w", err)
		}
		// Check for path traversal
		if strings.Contains(payload.FilePath, "..") {
			return fmt.Errorf("filePath: invalid path")
		}
	}

	// Validate file name (optional, but if provided, validate)
	if payload.FileName != "" {
		if err := validation.ValidateString(payload.FileName, 1, 255, "fileName"); err != nil {
			return fmt.Errorf("fileName: %w", err)
		}
		// Validate file extension
		allowedExts := []string{".pdf", ".doc", ".docx"}
		if err := validation.ValidateFileExtension(payload.FileName, allowedExts); err != nil {
			return fmt.Errorf("fileName: %w", err)
		}
	}

	// Validate file size (optional, but if provided, validate)
	if payload.FileSize > 0 {
		maxSize := int64(10 * 1024 * 1024) // 10MB
		if err := validation.ValidateFileSize(payload.FileSize, maxSize); err != nil {
			return fmt.Errorf("fileSize: %w", err)
		}
	}

	// Validate tags (optional, but if provided, validate each tag)
	if len(payload.Tags) > 0 {
		if len(payload.Tags) > 20 {
			return fmt.Errorf("tags: too many tags (maximum 20)")
		}
		for i, tag := range payload.Tags {
			if err := validation.ValidateString(tag, 1, 50, fmt.Sprintf("tags[%d]", i)); err != nil {
				return fmt.Errorf("tags[%d]: %w", i, err)
			}
			// Check for SQL injection and XSS
			if err := validation.ValidateNoSQLInjection(tag); err != nil {
				return fmt.Errorf("tags[%d]: %w", i, err)
			}
			if err := validation.ValidateNoXSS(tag); err != nil {
				return fmt.Errorf("tags[%d]: %w", i, err)
			}
		}
	}

	return nil
}

// ValidateUpdateResumePayload validates update resume payload
func ValidateUpdateResumePayload(payload *updateResumePayload) error {
	// Validate title (optional, but if provided, validate)
	if payload.Title != "" {
		if err := validation.ValidateString(payload.Title, 1, 200, "title"); err != nil {
			return fmt.Errorf("title: %w", err)
		}
		// Check for SQL injection and XSS
		if err := validation.ValidateNoSQLInjection(payload.Title); err != nil {
			return fmt.Errorf("title: %w", err)
		}
		if err := validation.ValidateNoXSS(payload.Title); err != nil {
			return fmt.Errorf("title: %w", err)
		}
	}

	// Validate tags (optional, but if provided, validate each tag)
	if len(payload.Tags) > 0 {
		if len(payload.Tags) > 20 {
			return fmt.Errorf("tags: too many tags (maximum 20)")
		}
		for i, tag := range payload.Tags {
			if err := validation.ValidateString(tag, 1, 50, fmt.Sprintf("tags[%d]", i)); err != nil {
				return fmt.Errorf("tags[%d]: %w", i, err)
			}
			// Check for SQL injection and XSS
			if err := validation.ValidateNoSQLInjection(tag); err != nil {
				return fmt.Errorf("tags[%d]: %w", i, err)
			}
			if err := validation.ValidateNoXSS(tag); err != nil {
				return fmt.Errorf("tags[%d]: %w", i, err)
			}
		}
	}

	return nil
}

// ValidateGenerateResumePayload validates generate resume payload
func ValidateGenerateResumePayload(payload *generateResumePayload) error {
	// Validate that either jobApplicationId or jobDescription is provided
	if payload.JobApplicationID == "" && strings.TrimSpace(payload.JobDescription) == "" {
		return fmt.Errorf("either jobApplicationId or jobDescription is required")
	}

	// If jobApplicationId provided, validate UUID
	if payload.JobApplicationID != "" {
		if err := validation.ValidateUUID(payload.JobApplicationID); err != nil {
			return fmt.Errorf("jobApplicationId: %w", err)
		}
	}

	// Validate language (optional). Accept either 2-letter ISO codes or full language names.
	if payload.Language != "" {
		l := strings.ToLower(strings.TrimSpace(payload.Language))
		allowedFull := map[string]bool{"english": true, "portuguese": true, "spanish": true, "french": true, "german": true}
		if len(l) == 2 {
			// accept two-letter codes (e.g., en, pt, es)
			if l != strings.ToLower(l) {
				return fmt.Errorf("language: must be lowercase")
			}
		} else if !allowedFull[l] {
			return fmt.Errorf("language: unsupported value")
		}
	}

	// Validate template (optional, but if provided, validate)
	if payload.Template != "" {
		if err := validation.ValidateString(payload.Template, 1, 100, "template"); err != nil {
			return fmt.Errorf("template: %w", err)
		}
		// Check for SQL injection and XSS
		if err := validation.ValidateNoSQLInjection(payload.Template); err != nil {
			return fmt.Errorf("template: %w", err)
		}
		if err := validation.ValidateNoXSS(payload.Template); err != nil {
			return fmt.Errorf("template: %w", err)
		}
	}

	return nil
}

// ValidateListResumesQueryParams validates query parameters for ListResumes
func ValidateListResumesQueryParams(limit, offset int, search string) error {
	// Validate limit
	if limit < 1 {
		return fmt.Errorf("limit: must be at least 1")
	}
	if limit > 200 {
		return fmt.Errorf("limit: must be at most 200")
	}

	// Validate offset
	if offset < 0 {
		return fmt.Errorf("offset: must be at least 0")
	}

	// Validate search (optional, but if provided, validate length)
	if search != "" {
		if err := validation.ValidateString(search, 1, 200, "search"); err != nil {
			return fmt.Errorf("search: %w", err)
		}
		// Check for SQL injection and XSS
		if err := validation.ValidateNoSQLInjection(search); err != nil {
			return fmt.Errorf("search: %w", err)
		}
		if err := validation.ValidateNoXSS(search); err != nil {
			return fmt.Errorf("search: %w", err)
		}
	}

	return nil
}

// ValidateUploadResumeFile validates uploaded file
func ValidateUploadResumeFile(filename string, size int64, contentType string) error {
	// Validate file extension
	allowedExts := []string{".pdf"}
	if err := validation.ValidateFileExtension(filename, allowedExts); err != nil {
		return fmt.Errorf("file: %w", err)
	}

	// Validate file size (max 10MB)
	maxSize := int64(10 * 1024 * 1024)
	if err := validation.ValidateFileSize(size, maxSize); err != nil {
		return fmt.Errorf("file: %w", err)
	}

	// Validate content type
	if contentType != "" && contentType != "application/pdf" {
		return fmt.Errorf("file: only PDF files are allowed")
	}

	return nil
}

// normalizeLanguage maps short language codes and variants to canonical full language names.
func normalizeLanguage(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	switch l {
	case "en", "eng", "english":
		return "english"
	case "pt", "pt-br", "pt-pt", "portuguese":
		return "portuguese"
	case "es", "spa", "spanish":
		return "spanish"
	case "fr", "french":
		return "french"
	case "de", "german":
		return "german"
	default:
		return l
	}
}

