package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jobapplications "woragis-jobs-service/internal/domains/jobapplications"
	"woragis-jobs-service/pkg/response"
)

// TestCreateJobApplicationEndpoint_Contract tests the CreateJobApplication endpoint HTTP contract
func TestCreateJobApplicationEndpoint_Contract(t *testing.T) {
	app := fiber.New()

	app.Post("/api/v1/job-applications", func(c *fiber.Ctx) error {
		var payload map[string]interface{}
		if err := c.BodyParser(&payload); err != nil {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{
				"message": "invalid request payload",
			})
		}

		// Contract: Valid payload should return 201
		if payload["companyName"] != nil && payload["jobTitle"] != nil {
			application := jobapplications.JobApplication{
				ID:          uuid.New(),
				UserID:      uuid.New(),
				CompanyName: payload["companyName"].(string),
				JobTitle:    payload["jobTitle"].(string),
				JobURL:      payload["jobUrl"].(string),
				Website:     payload["website"].(string),
				Status:      jobapplications.ApplicationStatusPending,
			}
			return response.Success(c, fiber.StatusCreated, application)
		}

		return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{
			"message": "invalid request payload",
		})
	})

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		validateResponse func(t *testing.T, resp *http.Response)
	}{
		{
			name: "valid payload returns 201",
			payload: map[string]interface{}{
				"companyName": "Test Company",
				"location":    "New York, NY",
				"jobTitle":    "Software Engineer",
				"jobUrl":      "https://example.com/job/123",
				"website":     "linkedin",
			},
			expectedStatus: 201,
			validateResponse: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

				var responseBody map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&responseBody)
				require.NoError(t, err)

				// Contract: Response should have standard structure
				assert.Contains(t, responseBody, "success")
				assert.Contains(t, responseBody, "data")

				// Contract: Data should contain job application
				data := responseBody["data"].(map[string]interface{})
				assert.Contains(t, data, "id")
				assert.Contains(t, data, "userId")
				assert.Contains(t, data, "companyName")
				assert.Contains(t, data, "jobTitle")
				assert.Contains(t, data, "status")
			},
		},
		{
			name: "missing required field returns 400",
			payload: map[string]interface{}{
				"companyName": "Test Company",
				// Missing jobTitle
				"jobUrl":  "https://example.com/job/123",
				"website": "linkedin",
			},
			expectedStatus: 400,
			validateResponse: func(t *testing.T, resp *http.Response) {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/job-applications", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validateResponse != nil {
				tt.validateResponse(t, resp)
			}
		})
	}
}

// TestGetJobApplicationEndpoint_Contract tests the GetJobApplication endpoint HTTP contract
func TestGetJobApplicationEndpoint_Contract(t *testing.T) {
	app := fiber.New()

	appID := uuid.New()
	app.Get("/api/v1/job-applications/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		if id == "invalid" {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{
				"message": "invalid application id",
			})
		}

		application := jobapplications.JobApplication{
			ID:          appID,
			UserID:      uuid.New(),
			CompanyName: "Test Company",
			JobTitle:    "Software Engineer",
			JobURL:      "https://example.com/job/123",
			Website:     "linkedin",
			Status:      jobapplications.ApplicationStatusApplied,
		}

		return response.Success(c, fiber.StatusOK, application)
	})

	tests := []struct {
		name           string
		applicationID string
		expectedStatus int
	}{
		{
			name:           "valid application ID returns 200",
			applicationID:  appID.String(),
			expectedStatus: 200,
		},
		{
			name:           "invalid application ID returns 400",
			applicationID:  "invalid",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/job-applications/"+tt.applicationID, nil)
			req.Header.Set("Authorization", "Bearer test-token")

			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		})
	}
}

// TestUpdateJobApplicationStatusEndpoint_Contract tests the UpdateStatus endpoint HTTP contract
func TestUpdateJobApplicationStatusEndpoint_Contract(t *testing.T) {
	app := fiber.New()

	app.Patch("/api/v1/job-applications/:id/status", func(c *fiber.Ctx) error {
		var payload map[string]interface{}
		if err := c.BodyParser(&payload); err != nil {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{
				"message": "invalid request payload",
			})
		}

		status, ok := payload["status"].(string)
		if !ok || status == "" {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{
				"message": "status is required",
			})
		}

		application := jobapplications.JobApplication{
			ID:     uuid.New(),
			Status: jobapplications.ApplicationStatus(status),
		}

		return response.Success(c, fiber.StatusOK, application)
	})

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid status update returns 200",
			payload: map[string]interface{}{
				"status": "applied",
			},
			expectedStatus: 200,
		},
		{
			name: "missing status returns 400",
			payload: map[string]interface{}{
				// Missing status
			},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			req := httptest.NewRequest("PATCH", "/api/v1/job-applications/"+uuid.New().String()+"/status", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer test-token")

			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
		})
	}
}

// TestListJobApplicationsEndpoint_Contract tests the ListJobApplications endpoint HTTP contract
func TestListJobApplicationsEndpoint_Contract(t *testing.T) {
	app := fiber.New()

	app.Get("/api/v1/job-applications", func(c *fiber.Ctx) error {
		applications := []jobapplications.JobApplication{
			{
				ID:          uuid.New(),
				CompanyName: "Company 1",
				JobTitle:    "Engineer",
				Status:      jobapplications.ApplicationStatusApplied,
			},
		}

		responseData := fiber.Map{
			"applications": applications,
			"total":        1,
			"limit":        50,
			"offset":       0,
		}

		return response.Success(c, fiber.StatusOK, responseData)
	})

	tests := []struct {
		name           string
		queryParams   string
		expectedStatus int
	}{
		{
			name:           "list applications without params",
			queryParams:    "",
			expectedStatus: 200,
		},
		{
			name:           "list applications with filters",
			queryParams:    "?status=applied&website=linkedin",
			expectedStatus: 200,
		},
		{
			name:           "list applications with pagination",
			queryParams:    "?limit=10&offset=0",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/job-applications"+tt.queryParams, nil)
			req.Header.Set("Authorization", "Bearer test-token")

			resp, err := app.Test(req)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			if resp.StatusCode == 200 {
				var responseBody map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&responseBody)
				require.NoError(t, err)

				// Contract: Response should have applications and pagination
				data := responseBody["data"].(map[string]interface{})
				assert.Contains(t, data, "applications")
			}
		})
	}
}

// TestErrorResponse_Contract tests that error responses follow the contract
func TestErrorResponse_Contract(t *testing.T) {
	app := fiber.New()

	app.Post("/api/v1/job-applications/test-error", func(c *fiber.Ctx) error {
		return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{
			"message": "Test error message",
		})
	})

	req := httptest.NewRequest("POST", "/api/v1/job-applications/test-error", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, 400, resp.StatusCode)

	var errorBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorBody)
	require.NoError(t, err)

	// Contract: Error response should have standard structure
	assert.Contains(t, errorBody, "success")
	assert.Contains(t, errorBody, "data")
	assert.Equal(t, false, errorBody["success"])
}

// TestContentType_Contract tests that all endpoints return JSON content type
func TestContentType_Contract(t *testing.T) {
	app := fiber.New()

	endpoints := []struct {
		method  string
		path    string
		handler fiber.Handler
	}{
		{"POST", "/api/v1/job-applications", func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{"message": "test"})
		}},
		{"GET", "/api/v1/job-applications/:id", func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{"message": "test"})
		}},
		{"PATCH", "/api/v1/job-applications/:id/status", func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{"message": "test"})
		}},
		{"GET", "/api/v1/job-applications", func(c *fiber.Ctx) error {
			return response.Error(c, fiber.StatusBadRequest, 400, fiber.Map{"message": "test"})
		}},
	}

	for _, endpoint := range endpoints {
		app.Add(endpoint.method, endpoint.path, endpoint.handler)

		path := endpoint.path
		if path == "/api/v1/job-applications/:id" || path == "/api/v1/job-applications/:id/status" {
			path = "/api/v1/job-applications/" + uuid.New().String()
			if endpoint.method == "PATCH" {
				path += "/status"
			}
		}

		req := httptest.NewRequest(endpoint.method, path, bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)

		// Contract: All endpoints should return JSON
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"),
			"Endpoint %s %s should return JSON", endpoint.method, endpoint.path)
	}
}

