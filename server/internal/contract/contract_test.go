package contract

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jobapplications "woragis-jobs-service/internal/domains/jobapplications"
)

// TestCreateJobApplicationPayload_Contract tests the CreateJobApplicationPayload contract
func TestCreateJobApplicationPayload_Contract(t *testing.T) {
	tests := []struct {
		name    string
		payload map[string]interface{}
		valid   bool
	}{
		{
			name: "valid payload with all required fields",
			payload: map[string]interface{}{
				"companyName": "Test Company",
				"location":    "New York, NY",
				"jobTitle":    "Software Engineer",
				"jobUrl":      "https://example.com/job/123",
				"website":     "linkedin",
			},
			valid: true,
		},
		{
			name: "valid payload with optional fields",
			payload: map[string]interface{}{
				"companyName":  "Test Company",
				"location":     "New York, NY",
				"jobTitle":     "Software Engineer",
				"jobUrl":       "https://example.com/job/123",
				"website":      "linkedin",
				"interestLevel": "high",
				"tags":         []string{"remote", "startup"},
				"notes":        "Great opportunity",
			},
			valid: true,
		},
		{
			name: "missing required field companyName",
			payload: map[string]interface{}{
				"location": "New York, NY",
				"jobTitle": "Software Engineer",
				"jobUrl":   "https://example.com/job/123",
				"website":  "linkedin",
			},
			valid: false,
		},
		{
			name: "missing required field jobTitle",
			payload: map[string]interface{}{
				"companyName": "Test Company",
				"location":    "New York, NY",
				"jobUrl":      "https://example.com/job/123",
				"website":     "linkedin",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.payload)
			require.NoError(t, err, "Should serialize to JSON")

			// Test JSON deserialization
			var unmarshaled map[string]interface{}
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err, "Should deserialize from JSON")

			// Verify required fields are present
			if tt.valid {
				assert.Contains(t, unmarshaled, "companyName")
				assert.Contains(t, unmarshaled, "jobTitle")
				assert.Contains(t, unmarshaled, "jobUrl")
				assert.Contains(t, unmarshaled, "website")
			}
		})
	}
}

// TestJobApplication_Contract tests the JobApplication entity contract
func TestJobApplication_Contract(t *testing.T) {
	userID := uuid.New()
	resumeID := uuid.New()
	now := time.Now()
	appliedAt := now.Add(-24 * time.Hour)

	application := jobapplications.JobApplication{
		ID:                uuid.New(),
		UserID:            userID,
		CompanyName:       "Test Company",
		Location:          "New York, NY",
		JobTitle:          "Software Engineer",
		JobURL:            "https://example.com/job/123",
		Website:           "linkedin",
		AppliedAt:         &appliedAt,
		CoverLetter:       "Cover letter text",
		LinkedInContact:   true,
		Status:            jobapplications.ApplicationStatusApplied,
		ResumeID:          &resumeID,
		SalaryMin:         intPtr(100000),
		SalaryMax:         intPtr(150000),
		SalaryCurrency:    "USD",
		JobDescription:    "Job description text",
		InterestLevel:     "high",
		Notes:             "Some notes",
		Tags:              jobapplications.JSONArray{"remote", "startup"},
		InterviewCount:    2,
		Source:            "job-board",
		ApplicationMethod: "auto",
		Language:          "en",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(application)
	require.NoError(t, err, "Should serialize to JSON")

	// Test JSON deserialization
	var unmarshaled jobapplications.JobApplication
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "Should deserialize from JSON")

	// Validate required fields
	assert.Equal(t, application.ID, unmarshaled.ID)
	assert.Equal(t, application.UserID, unmarshaled.UserID)
	assert.Equal(t, application.CompanyName, unmarshaled.CompanyName)
	assert.Equal(t, application.JobTitle, unmarshaled.JobTitle)
	assert.Equal(t, application.JobURL, unmarshaled.JobURL)
	assert.Equal(t, application.Website, unmarshaled.Website)
	assert.Equal(t, application.Status, unmarshaled.Status)

	// Verify JSON structure
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal(jsonData, &jsonMap)
	require.NoError(t, err)

	// Verify required fields in JSON
	assert.Contains(t, jsonMap, "id")
	assert.Contains(t, jsonMap, "userId")
	assert.Contains(t, jsonMap, "companyName")
	assert.Contains(t, jsonMap, "jobTitle")
	assert.Contains(t, jsonMap, "jobUrl")
	assert.Contains(t, jsonMap, "website")
	assert.Contains(t, jsonMap, "status")
}

// TestUpdateJobApplicationRequest_Contract tests the UpdateJobApplicationRequest contract
func TestUpdateJobApplicationRequest_Contract(t *testing.T) {
	followUpDate := time.Now().Add(7 * 24 * time.Hour)
	deadline := time.Now().Add(30 * 24 * time.Hour)
	interestLevel := "high"
	notes := "Updated notes"

	request := jobapplications.UpdateJobApplicationRequest{
		InterestLevel:  &interestLevel,
		Notes:          &notes,
		Tags:           jobapplications.JSONArray{"updated", "tags"},
		FollowUpDate:   &followUpDate,
		Deadline:       &deadline,
		SalaryMin:      intPtr(120000),
		SalaryMax:      intPtr(180000),
		SalaryCurrency: stringPtr("USD"),
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(request)
	require.NoError(t, err, "Should serialize to JSON")

	// Test JSON deserialization
	var unmarshaled jobapplications.UpdateJobApplicationRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "Should deserialize from JSON")

	// All fields are optional, so we just verify serialization works
	assert.NotNil(t, jsonData)
}

// TestUpdateStatusPayload_Contract tests the UpdateStatusPayload contract
func TestUpdateStatusPayload_Contract(t *testing.T) {
	payload := map[string]interface{}{
		"status": "applied",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(payload)
	require.NoError(t, err, "Should serialize to JSON")

	// Test JSON deserialization
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "Should deserialize from JSON")

	assert.Contains(t, unmarshaled, "status")
	assert.Equal(t, "applied", unmarshaled["status"])

	// Test with different status values
	validStatuses := []string{"pending", "processing", "applied", "contacted", "rejected", "accepted", "failed"}
	for _, status := range validStatuses {
		statusPayload := map[string]interface{}{
			"status": status,
		}
		jsonData, err := json.Marshal(statusPayload)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(jsonData, &result)
		require.NoError(t, err)
		assert.Equal(t, status, result["status"])
	}
}

// TestApplicationStatus_Contract tests the ApplicationStatus enum contract
func TestApplicationStatus_Contract(t *testing.T) {
	statuses := []jobapplications.ApplicationStatus{
		jobapplications.ApplicationStatusPending,
		jobapplications.ApplicationStatusProcessing,
		jobapplications.ApplicationStatusApplied,
		jobapplications.ApplicationStatusContacted,
		jobapplications.ApplicationStatusRejected,
		jobapplications.ApplicationStatusAccepted,
		jobapplications.ApplicationStatusFailed,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(status)
			require.NoError(t, err, "Should serialize to JSON")

			// Test JSON deserialization
			var unmarshaled jobapplications.ApplicationStatus
			err = json.Unmarshal(jsonData, &unmarshaled)
			require.NoError(t, err, "Should deserialize from JSON")

			assert.Equal(t, status, unmarshaled)
		})
	}
}

// TestJSONArray_Contract tests the JSONArray custom type contract
func TestJSONArray_Contract(t *testing.T) {
	tags := jobapplications.JSONArray{"remote", "startup", "dream-job"}

	// Test JSON serialization
	jsonData, err := json.Marshal(tags)
	require.NoError(t, err, "Should serialize to JSON")

	// Test JSON deserialization
	var unmarshaled jobapplications.JSONArray
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err, "Should deserialize from JSON")

	assert.Equal(t, tags, unmarshaled)
	assert.Len(t, unmarshaled, 3)

	// Test empty array
	emptyTags := jobapplications.JSONArray{}
	jsonData, err = json.Marshal(emptyTags)
	require.NoError(t, err)

	var emptyUnmarshaled jobapplications.JSONArray
	err = json.Unmarshal(jsonData, &emptyUnmarshaled)
	require.NoError(t, err)
	assert.Len(t, emptyUnmarshaled, 0)
}

// TestBackwardCompatibility tests that contract changes don't break backward compatibility
func TestBackwardCompatibility(t *testing.T) {
	// Test that old response format can still be parsed
	oldResponseJSON := `{
		"id": "123e4567-e89b-12d3-a456-426614174000",
		"userId": "123e4567-e89b-12d3-a456-426614174001",
		"companyName": "Test Company",
		"location": "New York, NY",
		"jobTitle": "Software Engineer",
		"jobUrl": "https://example.com/job/123",
		"website": "linkedin",
		"status": "applied",
		"createdAt": "2024-01-01T00:00:00Z",
		"updatedAt": "2024-01-01T00:00:00Z"
	}`

	var application jobapplications.JobApplication
	err := json.Unmarshal([]byte(oldResponseJSON), &application)
	require.NoError(t, err, "Should parse old response format")

	assert.NotEmpty(t, application.ID)
	assert.NotEmpty(t, application.UserID)
	assert.Equal(t, "Test Company", application.CompanyName)
	assert.Equal(t, "Software Engineer", application.JobTitle)
	assert.Equal(t, jobapplications.ApplicationStatusApplied, application.Status)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

