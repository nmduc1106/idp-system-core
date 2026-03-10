package handlers

import (
	"idp-api-gateway/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service ports.AdminService
}

func NewAdminHandler(service ports.AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

// GetStats godoc
// @Summary      Get System Statistics
// @Description  Returns total users, total jobs, and job counts grouped by state. ADMIN only.
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]interface{} "System statistics"
// @Failure      403 {object} map[string]string "Forbidden"
// @Router       /api/v1/admin/stats [get]
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetSystemStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetJobs godoc
// @Summary      Get All Jobs
// @Description  Returns all jobs with associated user info ordered by created_at DESC. ADMIN only.
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} map[string]interface{} "List of all jobs"
// @Failure      403 {object} map[string]string "Forbidden"
// @Router       /api/v1/admin/jobs [get]
func (h *AdminHandler) GetJobs(c *gin.Context) {
	jobs, err := h.service.GetAllJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// GetUsers godoc
// @Summary      Get All Users
// @Description  Returns all registered users (passwords excluded). ADMIN only.
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Success      200 {array} domain.User "List of all users"
// @Failure      403 {object} map[string]string "Forbidden"
// @Router       /api/v1/admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}
