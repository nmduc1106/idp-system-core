package handlers

import (
	"bytes"
	"encoding/json"
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// allowedMagicMimeTypes defines the MIME types permitted after magic bytes detection.
var allowedMagicMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"application/pdf": true,
}

// fileCodeRegex: alphanumeric, dashes, underscores only, max 50 chars.
var fileCodeRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]{1,50}$`)

type HTTPHandler struct {
	service ports.IDPService
}

func NewHTTPHandler(service ports.IDPService) *HTTPHandler {
	return &HTTPHandler{service: service}
}

// Upload godoc
// @Summary      Upload Document
// @Description  Upload a PDF or Image file with metadata. Requires Bearer Token. Validates file content via Magic Bytes.
// @Tags         jobs
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Document file (PDF, PNG, JPG)"
// @Param        file_code formData string true "User-defined document code (alphanumeric, dashes, max 50 chars)"
// @Param        notes formData string false "Optional notes about the document"
// @Success      200 {object} map[string]string "Returns job_id and status"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      400 {object} map[string]string "Validation error or unsupported file type"
// @Router       /api/v1/upload [post]
func (h *HTTPHandler) Upload(c *gin.Context) {
	// 1. Extract UserID from Context (injected by Auth Middleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing User Context"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// 2. Validate file_code (required, alphanumeric + dashes, max 50 chars)
	fileCode := strings.TrimSpace(c.PostForm("file_code"))
	if fileCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_code is required"})
		return
	}
	if !fileCodeRegex.MatchString(fileCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_code must be alphanumeric with dashes/underscores only, max 50 characters"})
		return
	}

	// 3. Sanitize optional notes
	notes := strings.TrimSpace(c.PostForm("notes"))

	// 4. Get file from request
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 5. Open file stream
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open file"})
		return
	}
	defer file.Close()

	// 6. Magic Bytes Validation: Read first 512 bytes to detect actual MIME type
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read file for validation"})
		return
	}
	detectedType := http.DetectContentType(buf[:n])

	if !allowedMagicMimeTypes[detectedType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "Unsupported file type detected. Allowed: JPEG, PNG, PDF",
			"detected_type": detectedType,
		})
		return
	}

	// 7. Reconstruct the full reader: prepend the already-read bytes back
	fullReader := io.MultiReader(bytes.NewReader(buf[:n]), file)

	// 8. Call Service (pass userID + metadata)
	job, err := h.service.UploadDocument(
		c.Request.Context(),
		userID,
		fileHeader.Filename,
		fileHeader.Size,
		fullReader,
		detectedType, // Use magic-bytes detected type, NOT client header
		fileCode,
		notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 9. Return result
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload successful",
		"job_id":  job.ID,
		"doc_id":  job.DocumentID,
		"status":  job.State,
	})
}

// GetUserJobs godoc
// @Summary      Get User's Jobs (Paginated)
// @Description  Returns the authenticated user's jobs with pagination and optional filters.
// @Tags         jobs
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Items per page (default: 10, max: 100)"
// @Param        status query string false "Filter by status (PENDING, EXTRACTING, COMPLETED, FAILED)"
// @Param        file_code query string false "Filter by file code (substring match)"
// @Success      200 {object} domain.PaginatedResponse
// @Failure      401 {object} map[string]string "Unauthorized"
// @Router       /api/v1/jobs [get]
func (h *HTTPHandler) GetUserJobs(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing User Context"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	q := domain.PaginationQuery{
		Page:     page,
		Limit:    limit,
		Status:   strings.TrimSpace(c.Query("status")),
		FileCode: strings.TrimSpace(c.Query("file_code")),
	}
	q.Normalize()

	result, err := h.service.GetUserJobs(c.Request.Context(), userID, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetJob godoc
// @Summary      Get Job Status
// @Tags         jobs
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Job ID (UUID)"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "Job not found"
// @Router       /api/v1/jobs/{id} [get]
func (h *HTTPHandler) GetJob(c *gin.Context) {
	// 1. Extract UserID from Context (injected by Auth Middleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing User Context"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// 2. Get job with ownership enforcement
	id := c.Param("id")
	job, err := h.service.GetJobStatus(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}

// StreamJob godoc
// @Summary      Stream Job Status (SSE)
// @Description  Subscribe to Server-Sent Events for real-time job updates.
// @Tags         jobs
// @Security     BearerAuth
// @Produce      text/event-stream
// @Param        id path string true "Job ID (UUID)"
// @Success      200 {string} string "Event stream"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "Job not found"
// @Router       /api/v1/jobs/{id}/stream [get]
func (h *HTTPHandler) StreamJob(c *gin.Context) {
	// 1. Extract UserID from Context
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing User Context"})
		return
	}
	userID := userIDVal.(uuid.UUID)
	jobID := c.Param("id")

	// 2. Get SSE Channel from Service (Verifies ownership internally)
	msgChan, err := h.service.StreamJobStatus(c.Request.Context(), userID, jobID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found or unauthorized"})
		return
	}

	// 3. Set SSE Headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// 4. Stream Event using Gin's Stream
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return false // Channel closed
			}

			// Clean the Double-Encoded JSON from Python
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(msg), &payload); err == nil {
				// Parse the nested "result" string if it exists
				if resultStr, ok := payload["result"].(string); ok {
					var rawResult json.RawMessage
					if json.Unmarshal([]byte(resultStr), &rawResult) == nil {
						payload["result"] = rawResult
					}
				}
				// Send the cleaned JSON
				c.SSEvent("message", payload)
			} else {
				// Fallback to raw string if unmarshal fails
				c.SSEvent("message", msg)
			}

			return false // Return false to close the stream after the first message
		case <-c.Request.Context().Done():
			return false // Client disconnected
		}
	})
}

// ExportJobsExcel godoc
// @Summary      Export Completed Jobs to Excel
// @Description  Downloads an Excel file containing extracted OCR data for completed jobs.
// @Tags         jobs
// @Security     BearerAuth
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param        file_code query string false "Filter by File Code"
// @Success      200 {file} file "IDP_Report.xlsx"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Failed to generate report"
// @Router       /api/v1/jobs/export [get]
func (h *HTTPHandler) ExportJobsExcel(c *gin.Context) {
	// 1. Extract UserID from Context
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing User Context"})
		return
	}
	userID := userIDVal.(uuid.UUID).String()

	// 2. Extract query params
	fileCode := c.Query("file_code")

	// 3. Generate Excel
	buf, err := h.service.ExportJobsToExcel(c.Request.Context(), userID, fileCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 4. Stream response
	c.Header("Content-Disposition", "attachment; filename=IDP_Report.xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Transfer-Encoding", "binary")
	
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
