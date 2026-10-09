package response

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"siakad-mini/internal/apperror"
)

type body struct {
	Success bool                `json:"success"`
	Message string              `json:"message"`
	Data    any                 `json:"data,omitempty"`
	Meta    any                 `json:"meta,omitempty"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, body{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, body{Success: true, Message: message, Data: data})
}

func List(c *gin.Context, message string, data any, meta Meta) {
	c.JSON(http.StatusOK, body{Success: true, Message: message, Data: data, Meta: meta})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

// Fail menulis response error. AppError dipakai apa adanya,
// error lain dicatat di log server lalu dijawab 500 generik (tanpa stack trace).
func Fail(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.AbortWithStatusJSON(appErr.Status, body{
			Success: false,
			Message: appErr.Message,
			Errors:  appErr.Errors,
		})
		return
	}
	log.Printf("[ERROR] %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	InternalError(c)
}

func InternalError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, body{
		Success: false,
		Message: "Terjadi kesalahan pada server",
	})
}
