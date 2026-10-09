package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/repository"
	"siakad-mini/internal/response"
	"siakad-mini/internal/service"
	"siakad-mini/internal/validator"
)

type CourseHandler struct {
	svc *service.CourseService
}

func NewCourseHandler(svc *service.CourseService) *CourseHandler {
	return &CourseHandler{svc: svc}
}

// List: GET /api/v1/courses
func (h *CourseHandler) List(c *gin.Context) {
	errs := validator.Errors{}

	var semester *int
	if v := c.Query("semester"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			errs.Add("semester", "Semester harus berupa angka")
		} else {
			semester = &n
		}
	}
	if errs.Any() {
		response.Fail(c, apperror.Validation(errs))
		return
	}

	items, err := h.svc.List(c.Request.Context(), repository.CourseFilter{
		Semester:  semester,
		Search:    strings.TrimSpace(c.Query("search")),
		Available: c.Query("available") == "true",
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, "Data mata kuliah berhasil diambil", items)
}
