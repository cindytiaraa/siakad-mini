package handler

import (
	"github.com/gin-gonic/gin"

	"siakad-mini/internal/dto"
	"siakad-mini/internal/middleware"
	"siakad-mini/internal/response"
	"siakad-mini/internal/service"
)

type EnrollmentHandler struct {
	svc *service.EnrollmentService
}

func NewEnrollmentHandler(svc *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{svc: svc}
}

// Create: POST /api/v1/enrollments
func (h *EnrollmentHandler) Create(c *gin.Context) {
	var req dto.EnrollmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bindError(err))
		return
	}
	// Identitas mahasiswa selalu dari token, bukan dari body.
	e, err := h.svc.Create(c.Request.Context(), middleware.StudentID(c), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, "Mata kuliah berhasil ditambahkan ke KRS", e)
}

// Delete: DELETE /api/v1/enrollments/:id
func (h *EnrollmentHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), middleware.StudentID(c), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}
