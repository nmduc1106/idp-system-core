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
// @Description  Authenticate user and set HttpOnly cookies (access_token 15m + refresh_token 7d).
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

	accessToken, refreshToken, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Access Token Cookie: 15 minutes, path "/" for all routes
	c.SetCookie("access_token", accessToken, 15*60, "/", "", false, true)

	// Refresh Token Cookie: 7 days, path restricted to /api/v1/auth/refresh for security
	c.SetCookie("refresh_token", refreshToken, 7*24*60*60, "/api/v1/auth/refresh", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

// Refresh godoc
// @Summary      Refresh Access Token
// @Description  Use refresh_token cookie to obtain a new short-lived access_token.
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string "Token refreshed"
// @Failure      401 {object} map[string]string "Invalid or expired refresh token"
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token required"})
		return
	}

	newAccessToken, err := h.service.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Set the new access_token cookie (15 minutes)
	c.SetCookie("access_token", newAccessToken, 15*60, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed"})
}

// Logout godoc
// @Summary      Logout
// @Description  Logout user, invalidate refresh token in Redis, and clear all auth cookies.
// @Tags         auth
// @Produce      json
// @Success      200 {object} map[string]string "Logout successful"
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract userID from context if available (Logout is in public group, so may not have middleware)
	userIDExtracted, exists := c.Get("userID")
	if exists {
		userIDStr := fmt.Sprintf("%v", userIDExtracted)
		_ = h.service.Logout(c.Request.Context(), userIDStr)
	}

	// Clear both cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth/refresh", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
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

	resp := UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FullName:  user.FullName,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, resp)
}