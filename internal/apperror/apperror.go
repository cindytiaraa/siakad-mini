package apperror

import "net/http"

// AppError adalah error yang sudah punya status HTTP dan pesan untuk client.
type AppError struct {
	Status  int
	Message string
	Errors  map[string][]string
}

func (e *AppError) Error() string { return e.Message }

func New(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

func Unauthorized(message string) *AppError    { return New(http.StatusUnauthorized, message) }
func Forbidden(message string) *AppError       { return New(http.StatusForbidden, message) }
func NotFound(message string) *AppError        { return New(http.StatusNotFound, message) }
func Conflict(message string) *AppError        { return New(http.StatusConflict, message) }
func TooManyRequests(message string) *AppError { return New(http.StatusTooManyRequests, message) }

// Unprocessable menghasilkan 422 dengan pesan khusus (dipakai untuk business rule).
func Unprocessable(message string, errs map[string][]string) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Message: message, Errors: errs}
}

// Validation menghasilkan 422 "Validasi gagal" sesuai contoh PDF.
func Validation(errs map[string][]string) *AppError {
	return Unprocessable("Validasi gagal", errs)
}
