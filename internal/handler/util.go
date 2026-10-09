package handler

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/response"
)

// bindError mengubah kegagalan decode JSON menjadi 422.
func bindError(err error) *apperror.AppError {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) && typeErr.Field != "" {
		return apperror.Validation(map[string][]string{typeErr.Field: {"Tipe data tidak valid"}})
	}
	return apperror.Validation(map[string][]string{"body": {"Format JSON tidak valid"}})
}

// parseID membaca :id dari path. Id yang tidak valid diperlakukan sebagai data tidak ditemukan.
func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		response.Fail(c, apperror.NotFound("Data tidak ditemukan"))
		return 0, false
	}
	return id, true
}
