package handlers

import (
	"idp-api-gateway/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// allowedMimeTypes defines the set of acceptable file types for document upload.
var allowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/tiff":      true,
	"application/pdf": true,
}

type HTTPHandler struct {
	service ports.IDPService
}

func NewHTTPHandler(service ports.IDPService) *HTTPHandler {
	return &HTTPHandler{service: service}
}

// Upload godoc
// @Summary      Upload Document
// @Description  Upload a PDF or Image file. Requires Bearer Token.
// @Tags         jobs
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Document file (PDF, PNG, JPG, TIFF)"
// @Success      200 {object} map[string]string "Returns job_id and status"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      400 {object} map[string]string "No file uploaded or unsupported file type"
// @Router       /api/v1/upload [post]
func (h *HTTPHandler) Upload(c *gin.Context) {
	// 1. Extract UserID from Context (injected by Auth Middleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing User Context"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// 2. Get file from request
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 3. Validate MIME type (allow images + PDF)
	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedMimeTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type. Allowed: JPEG, PNG, TIFF, PDF"})
		return
	}

	// 4. Open file stream
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open file"})
		return
	}
	defer file.Close()

	// 5. Call Service (pass userID for ownership)
	job, err := h.service.UploadDocument(c.Request.Context(), userID, fileHeader.Filename, fileHeader.Size, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6. Return result
	c.JSON(http.StatusOK, gin.H{
		"message": "Upload successful",
		"job_id":  job.ID,
		"doc_id":  job.DocumentID,
		"status":  job.State,
	})
}

// GetJob godoc
// @Summary      Get Job Status
// @Tags         jobs
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Job ID (UUID)"
// @Success      200 {object} domain.Job
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