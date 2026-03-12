package handlers

import (
	"idp-api-gateway/internal/core/domain"
	"idp-api-gateway/internal/core/ports"
	"net/http"
	"strconv"
	"strings"

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
// @Summary      Get All Jobs (Paginated)
// @Description  Returns all jobs with pagination, optional status/file_code filters, and associated user/document info. ADMIN only.
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Items per page (default: 10, max: 100)"
// @Param        status query string false "Filter by status"
// @Param        file_code query string false "Filter by file code"
// @Success      200 {object} domain.PaginatedResponse
// @Failure      403 {object} map[string]string "Forbidden"
// @Router       /api/v1/admin/jobs [get]
func (h *AdminHandler) GetJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	q := domain.PaginationQuery{
		Page:     page,
		Limit:    limit,
		Status:   strings.TrimSpace(c.Query("status")),
		FileCode: strings.TrimSpace(c.Query("file_code")),
	}
	q.Normalize()

	result, err := h.service.GetAllJobs(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
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
