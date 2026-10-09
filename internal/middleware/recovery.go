package middleware

import (
	"github.com/gin-gonic/gin"

	"siakad-mini/internal/response"
)

// Recovery menangkap panic: stack trace hanya masuk log server,
// client menerima 500 generik.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		response.InternalError(c)
	})
}
