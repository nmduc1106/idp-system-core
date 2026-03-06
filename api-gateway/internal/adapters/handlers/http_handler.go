package handlers

import (
	"idp-api-gateway/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
// @Param        file formData file true "Document file (PDF, PNG, JPG)"
// @Success      200 {object} map[string]string "Returns job_id and status"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      400 {object} map[string]string "No file uploaded"
// @Router       /api/v1/upload [post]
func (h *HTTPHandler) Upload(c *gin.Context) {
	// 1. [QUAN TRỌNG] Lấy UserID từ Context (Do Auth Middleware nạp vào)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing User Context"})
		return
	}
	userID := userIDVal.(uuid.UUID) // Ép kiểu về UUID

	// 2. Lấy file từ request
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 3. Mở luồng đọc file (Stream)
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to open file"})
		return
	}
	defer file.Close()

	// 4. Gọi Service xử lý (Truyền userID vào)
	job, err := h.service.UploadDocument(c.Request.Context(), userID, fileHeader.Filename, fileHeader.Size, file, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5. Trả về kết quả
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
// @Router       /api/v1/jobs/{id} [get]
func (h *HTTPHandler) GetJob(c *gin.Context) {
	id := c.Param("id")
	job, err := h.service.GetJobStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}
	c.JSON(http.StatusOK, job)
}