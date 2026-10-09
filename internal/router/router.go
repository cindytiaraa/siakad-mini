package router

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/config"
	"siakad-mini/internal/handler"
	"siakad-mini/internal/middleware"
	"siakad-mini/internal/model"
	"siakad-mini/internal/repository"
	"siakad-mini/internal/response"
	"siakad-mini/internal/service"
)

func New(cfg *config.Config, db *sql.DB) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	_ = r.SetTrustedProxies(nil) // ClientIP = alamat koneksi langsung (untuk rate limiting)
	r.Use(gin.Logger(), middleware.Recovery())
	r.NoRoute(func(c *gin.Context) {
		response.Fail(c, apperror.NotFound("Endpoint tidak ditemukan"))
	})

	// repository
	userRepo := repository.NewUserRepository()
	studentRepo := repository.NewStudentRepository()
	courseRepo := repository.NewCourseRepository()
	enrollRepo := repository.NewEnrollmentRepository()

	// service
	authSvc := service.NewAuthService(db, userRepo, studentRepo, cfg.JWTSecret, cfg.JWTExpires)
	studentSvc := service.NewStudentService(db, userRepo, studentRepo, enrollRepo)
	courseSvc := service.NewCourseService(db, courseRepo)
	enrollSvc := service.NewEnrollmentService(db, studentRepo, courseRepo, enrollRepo)

	// handler
	authH := handler.NewAuthHandler(authSvc, middleware.NewLoginLimiter(5, time.Minute))
	studentH := handler.NewStudentHandler(studentSvc)
	courseH := handler.NewCourseHandler(courseSvc)
	enrollH := handler.NewEnrollmentHandler(enrollSvc)

	admin := middleware.RequireRole(model.RoleAdmin)
	mahasiswa := middleware.RequireRole(model.RoleMahasiswa)

	v1 := r.Group("/api/v1")

	// 1. publik
	v1.POST("/auth/login", authH.Login)

	// semua endpoint lain wajib token
	secured := v1.Group("", middleware.Auth(cfg.JWTSecret, db))
	secured.GET("/auth/me", authH.Me)                                // 2. semua role
	secured.GET("/students", admin, studentH.List)                   // 3. admin
	secured.POST("/students", admin, studentH.Create)                // 4. admin
	secured.GET("/students/:id", studentH.Detail)                    // 5. admin / mahasiswa (milik sendiri)
	secured.PUT("/students/:id", admin, studentH.Update)             // 6. admin
	secured.DELETE("/students/:id", admin, studentH.Delete)          // 7. admin
	secured.GET("/courses", courseH.List)                            // 8. semua role
	secured.POST("/enrollments", mahasiswa, enrollH.Create)          // 9. mahasiswa
	secured.DELETE("/enrollments/:id", mahasiswa, enrollH.Delete)    // 10. mahasiswa (milik sendiri)

	return r
}
