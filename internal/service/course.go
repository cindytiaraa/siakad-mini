package service

import (
	"context"
	"database/sql"

	"siakad-mini/internal/model"
	"siakad-mini/internal/repository"
)

type CourseService struct {
	db      *sql.DB
	courses *repository.CourseRepository
}

func NewCourseService(db *sql.DB, courses *repository.CourseRepository) *CourseService {
	return &CourseService{db: db, courses: courses}
}

func (s *CourseService) List(ctx context.Context, f repository.CourseFilter) ([]model.Course, error) {
	return s.courses.List(ctx, s.db, f)
}
