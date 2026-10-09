package middleware

import (
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/model"
	"siakad-mini/internal/repository"
	"siakad-mini/internal/response"
	"siakad-mini/internal/token"
)

const (
	ctxUserID    = "user_id"
	ctxRole      = "role"
	ctxStudentID = "student_id"
)

const msgUnauthorized = "Token tidak valid atau tidak ditemukan"

// Auth = AUTHENTICATION: memastikan token Bearer valid lalu menyimpan identitas ke context.
// Token milik mahasiswa yang sudah di-soft delete ditolak (401).
func Auth(secret string, db *sql.DB) gin.HandlerFunc {
	students := repository.NewStudentRepository()

	return func(c *gin.Context) {
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			response.Fail(c, apperror.Unauthorized(msgUnauthorized))
			return
		}

		claims, err := token.Parse(secret, strings.TrimSpace(parts[1]))
		if err != nil {
			response.Fail(c, apperror.Unauthorized(msgUnauthorized))
			return
		}

		var studentID int64
		if claims.Role == model.RoleMahasiswa {
			st, err := students.FindActiveByUserID(c.Request.Context(), db, claims.UserID)
			if err != nil {
				response.Fail(c, err)
				return
			}
			if st == nil {
				response.Fail(c, apperror.Unauthorized(msgUnauthorized))
				return
			}
			studentID = st.ID
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, claims.Role)
		c.Set(ctxStudentID, studentID)
		c.Next()
	}
}

// RequireRole = AUTHORIZATION berbasis role: role tidak sesuai -> 403.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := Role(c)
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		response.Fail(c, apperror.Forbidden("Anda tidak memiliki akses ke resource ini"))
	}
}

func UserID(c *gin.Context) int64 {
	v, _ := c.Get(ctxUserID)
	id, _ := v.(int64)
	return id
}

func Role(c *gin.Context) string {
	v, _ := c.Get(ctxRole)
	role, _ := v.(string)
	return role
}

// StudentID adalah id tabel students milik user yang login (0 untuk admin).
// Selalu diambil dari token, tidak pernah dari body/query.
func StudentID(c *gin.Context) int64 {
	v, _ := c.Get(ctxStudentID)
	id, _ := v.(int64)
	return id
}
