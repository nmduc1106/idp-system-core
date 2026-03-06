package handlers

import (
	"fmt"
	"idp-api-gateway/internal/core/ports"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service ports.AuthService
}

func NewAuthHandler(service ports.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account with email, password, and full name.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Register Info"
// @Success      201 {object} map[string]string "User registered successfully"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      500 {object} map[string]string "Internal Server Error"
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Register(c.Request.Context(), req.Email, req.Password, req.FullName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// Login godoc
// @Summary      Login
// @Description  Authenticate user and return JWT Token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login Info"
// @Success      200 {object} map[string]string "Login successful"
// @Failure      400 {object} map[string]string "Invalid input"
// @Failure      401 {object} map[string]string "Invalid credentials"
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Thiết lập HttpOnly Cookie để lưu trữ JWT an toàn
	// maxAge: 86400 (24 giờ), path: "/", domain: "", secure: false (để test localhost), httpOnly: true
	c.SetCookie("access_token", token, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

// Logout godoc
// @Summary      Logout
// @Description  Logout user and clear JWT cookie.
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string "Logout successful"
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetMe godoc
// @Summary      Get current user profile
// @Description  Get the profile of the currently authenticated user using JWT cookie or Bearer token fallback.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} UserResponse "Returns user details without password"
// @Failure      401 {object} map[string]string "Unauthorized"
// @Failure      404 {object} map[string]string "User not found"
// @Router       /api/v1/users/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDExtracted, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// userID là kiểu uuid.UUID (từ Middleware), convert sang string
	userIDStr := fmt.Sprintf("%v", userIDExtracted)

	user, err := h.service.GetMe(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Trả về UserResponse DTO để ẩn password_hash
	resp := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, resp)
}