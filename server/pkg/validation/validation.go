package validation

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// ValidateString validates a string's length
func ValidateString(s string, min, max int, fieldName string) error {
	if len(s) < min {
		return fmt.Errorf("%s must be at least %d characters", fieldName, min)
	}
	if len(s) > max {
		return fmt.Errorf("%s must be at most %d characters", fieldName, max)
	}
	return nil
}

// ValidateNoSQLInjection checks for common SQL injection patterns
func ValidateNoSQLInjection(s string) error {
	sqlPatterns := []string{
		"';",
		"--",
		"/*",
		"*/",
		"xp_",
		"sp_",
		"exec(",
		"execute(",
		"union select",
		"union all select",
		"drop table",
		"drop database",
		"delete from",
		"insert into",
		"update set",
		"alter table",
		"create table",
		"truncate table",
	}
	
	lower := strings.ToLower(s)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lower, pattern) {
			return errors.New("potential SQL injection detected")
		}
	}
	return nil
}

// ValidateNoXSS checks for common XSS patterns
func ValidateNoXSS(s string) error {
	xssPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"onerror=",
		"onload=",
		"onclick=",
		"onmouseover=",
		"<iframe",
		"<img",
		"<svg",
		"<object",
		"<embed",
		"<link",
		"<style",
		"expression(",
	}
	
	lower := strings.ToLower(s)
	for _, pattern := range xssPatterns {
		if strings.Contains(lower, pattern) {
			return errors.New("potential XSS attack detected")
		}
	}
	return nil
}

// ValidateURL validates a URL string
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return errors.New("URL cannot be empty")
	}
	
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}
	
	if parsedURL.Scheme == "" {
		return errors.New("URL must include a scheme (http:// or https://)")
	}
	
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("URL scheme must be http or https")
	}
	
	if parsedURL.Host == "" {
		return errors.New("URL must include a host")
	}
	
	return nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(uuidStr string) error {
	if uuidStr == "" {
		return errors.New("UUID cannot be empty")
	}
	
	_, err := uuid.Parse(uuidStr)
	if err != nil {
		return fmt.Errorf("invalid UUID format: %w", err)
	}
	
	return nil
}

// ValidateFileExtension validates a file extension
func ValidateFileExtension(filename string, allowedExts []string) error {
	if filename == "" {
		return errors.New("filename cannot be empty")
	}
	
	lowerFilename := strings.ToLower(filename)
	for _, ext := range allowedExts {
		if strings.HasSuffix(lowerFilename, strings.ToLower(ext)) {
			return nil
		}
	}
	
	return fmt.Errorf("file extension not allowed. Allowed extensions: %v", allowedExts)
}

// ValidateFileSize validates a file size
func ValidateFileSize(size, maxSize int64) error {
	if size < 0 {
		return errors.New("file size cannot be negative")
	}
	
	if size > maxSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}
	
	return nil
}

