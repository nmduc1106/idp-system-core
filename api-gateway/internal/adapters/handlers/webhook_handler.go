package handlers

import (
	"idp-api-gateway/internal/core/ports"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// WebhookHandler handles internal-only webhook calls from microservices (e.g., Python worker).
type WebhookHandler struct {
	idpService    ports.IDPService
	webhookSecret string
}

func NewWebhookHandler(idpService ports.IDPService) *WebhookHandler {
	secret := os.Getenv("WEBHOOK_SECRET")
	if secret == "" {
		panic("❌ WEBHOOK_SECRET environment variable is required")
	}
	return &WebhookHandler{idpService: idpService, webhookSecret: secret}
}

// JobCompleted godoc
// @Summary      Internal Webhook: Notify Job Completion
// @Description  Called by the Python worker after updating a job to COMPLETED/FAILED. Triggers cache invalidation. Requires X-Webhook-Secret header.
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        X-Webhook-Secret header string true "Shared webhook secret"
// @Param        body body map[string]string true "Payload: {\"job_id\": \"uuid\"}"
// @Success      200 {object} map[string]string "Cache invalidated"
// @Failure      400 {object} map[string]string "Missing job_id"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      500 {object} map[string]string "Invalidation failed"
// @Router       /internal/webhook/job-completed [post]
func (h *WebhookHandler) JobCompleted(c *gin.Context) {
	log.Printf("[WEBHOOK] 📥 Received job-completed webhook request")

	// 1. Authenticate: check X-Webhook-Secret header
	secret := c.GetHeader("X-Webhook-Secret")
	if secret == "" || secret != h.webhookSecret {
		log.Printf("[WEBHOOK] ❌ Secret validation failed. Received: '%s'", secret)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
	log.Printf("[WEBHOOK] ✅ Secret validated successfully")

	// 2. Parse body
	var body struct {
		JobID string `json:"job_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Printf("[WEBHOOK] ❌ Missing or invalid job_id in payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_id is required"})
		return
	}

	log.Printf("[WEBHOOK] Processing caching invalidation for JobID: %s", body.JobID)

	// 3. Invalidate caches
	if err := h.idpService.InvalidateJobCaches(c.Request.Context(), body.JobID); err != nil {
		log.Printf("[WEBHOOK] ❌ Cache invalidation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WEBHOOK] ✅ Cache invalidation completed successfully for JobID: %s", body.JobID)
	c.JSON(http.StatusOK, gin.H{"message": "Cache invalidated"})
}
