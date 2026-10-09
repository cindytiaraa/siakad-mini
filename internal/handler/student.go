package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/dto"
	"siakad-mini/internal/middleware"
	"siakad-mini/internal/repository"
	"siakad-mini/internal/response"
	"siakad-mini/internal/service"
	"siakad-mini/internal/validator"
)

type StudentHandler struct {
	svc *service.StudentService
}

func NewStudentHandler(svc *service.StudentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

// List: GET /api/v1/students
func (h *StudentHandler) List(c *gin.Context) {
	errs := validator.Errors{}

	page := 1
	if v := c.Query("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			errs.Add("page", "Page harus bilangan bulat minimal 1")
		} else {
			page = n
		}
	}

	perPage := 10
	if v := c.Query("per_page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			errs.Add("per_page", "Per page harus bilangan bulat antara 1 dan 50")
		} else {
			perPage = n
		}
	}

	var angkatan *int
	if v := c.Query("angkatan"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			errs.Add("angkatan", "Angkatan harus berupa angka")
		} else {
			angkatan = &n
		}
	}

	sort := c.Query("sort")
	if sort != "" && sort != "nama" && sort != "-ipk_terakhir" {
		errs.Add("sort", "Sort hanya boleh 'nama' atau '-ipk_terakhir'")
	}

	if errs.Any() {
		response.Fail(c, apperror.Validation(errs))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), repository.StudentFilter{
		Page:     page,
		PerPage:  perPage,
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: angkatan,
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     sort,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}

	lastPage := (total + perPage - 1) / perPage
	if lastPage < 1 {
		lastPage = 1
	}
	response.List(c, "Data mahasiswa berhasil diambil", items, response.Meta{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		LastPage:    lastPage,
	})
}

// Create: POST /api/v1/students
func (h *StudentHandler) Create(c *gin.Context) {
	var req dto.StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bindError(err))
		return
	}
	st, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Created(c, "Mahasiswa berhasil ditambahkan", st)
}

// Detail: GET /api/v1/students/:id
func (h *StudentHandler) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	detail, err := h.svc.Detail(c.Request.Context(), middleware.Role(c), middleware.StudentID(c), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, "Detail mahasiswa berhasil diambil", detail)
}

// Update: PUT /api/v1/students/:id
func (h *StudentHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bindError(err))
		return
	}
	st, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, "Data mahasiswa berhasil diperbarui", st)
}

// Delete: DELETE /api/v1/students/:id (soft delete)
func (h *StudentHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.NoContent(c)
}
