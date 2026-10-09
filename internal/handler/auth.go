package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/dto"
	"siakad-mini/internal/middleware"
	"siakad-mini/internal/response"
	"siakad-mini/internal/service"
)

type AuthHandler struct {
	svc     *service.AuthService
	limiter *middleware.LoginLimiter
}

func NewAuthHandler(svc *service.AuthService, limiter *middleware.LoginLimiter) *AuthHandler {
	return &AuthHandler{svc: svc, limiter: limiter}
}

// Login: POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bindError(err))
		return
	}

	key := c.ClientIP() + "|" + strings.ToLower(strings.TrimSpace(req.Email))
	if h.limiter.Blocked(key) {
		c.Header("Retry-After", "60")
		response.Fail(c, apperror.TooManyRequests("Terlalu banyak percobaan login. Coba lagi nanti"))
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.Status == http.StatusUnauthorized {
			h.limiter.Fail(key) // hanya kredensial salah yang dihitung
		}
		response.Fail(c, err)
		return
	}

	h.limiter.Reset(key)
	response.OK(c, "Login berhasil", result)
}

// Me: GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	me, err := h.svc.Me(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, "Data pengguna berhasil diambil", me)
}
