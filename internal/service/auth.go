package service

import (
	"context"
	"database/sql"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"siakad-mini/internal/apperror"
	"siakad-mini/internal/dto"
	"siakad-mini/internal/model"
	"siakad-mini/internal/repository"
	"siakad-mini/internal/token"
	"siakad-mini/internal/validator"
)

type AuthService struct {
	db        *sql.DB
	users     *repository.UserRepository
	students  *repository.StudentRepository
	secret    string
	expiresIn int // detik
}

func NewAuthService(db *sql.DB, users *repository.UserRepository, students *repository.StudentRepository, secret string, expiresIn int) *AuthService {
	return &AuthService{db: db, users: users, students: students, secret: secret, expiresIn: expiresIn}
}

type LoginUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type LoginResult struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	User        LoginUser `json:"user"`
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	errs := validator.Errors{}
	if email == "" {
		errs.Add("email", "Email wajib diisi")
	} else if !validator.IsEmail(email) {
		errs.Add("email", "Format email tidak valid")
	}
	if req.Password == "" {
		errs.Add("password", "Password wajib diisi")
	} else if utf8.RuneCountInString(req.Password) < 8 {
		errs.Add("password", "Password minimal 8 karakter")
	}
	if errs.Any() {
		return nil, apperror.Validation(errs)
	}

	// Pesan sama untuk semua kegagalan kredensial agar akun tidak bisa ditebak.
	invalid := apperror.Unauthorized("Email atau password salah")

	user, err := s.users.FindByEmail(ctx, s.db, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, invalid
	}

	// Mahasiswa yang sudah di-soft delete tidak dapat login.
	if user.Role == model.RoleMahasiswa {
		st, err := s.students.FindActiveByUserID(ctx, s.db, user.ID)
		if err != nil {
			return nil, err
		}
		if st == nil {
			return nil, invalid
		}
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, invalid
	}

	tok, err := token.Generate(s.secret, user.ID, user.Role, time.Duration(s.expiresIn)*time.Second)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken: tok,
		TokenType:   "Bearer",
		ExpiresIn:   s.expiresIn,
		User:        LoginUser{ID: user.ID, Email: user.Email, Role: user.Role},
	}, nil
}

type MeStudent struct {
	ID       int64  `json:"id"`
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}

type MeResponse struct {
	ID       int64      `json:"id"`
	Email    string     `json:"email"`
	Role     string     `json:"role"`
	Students *MeStudent `json:"students,omitempty"`
}

func (s *AuthService) Me(ctx context.Context, userID int64) (*MeResponse, error) {
	user, err := s.users.FindByID(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.Unauthorized("Token tidak valid atau tidak ditemukan")
	}

	res := &MeResponse{ID: user.ID, Email: user.Email, Role: user.Role}
	if user.Role == model.RoleMahasiswa {
		st, err := s.students.FindActiveByUserID(ctx, s.db, user.ID)
		if err != nil {
			return nil, err
		}
		if st == nil {
			return nil, apperror.Unauthorized("Token tidak valid atau tidak ditemukan")
		}
		res.Students = &MeStudent{ID: st.ID, NIM: st.NIM, Nama: st.Nama, Prodi: st.Prodi, Angkatan: st.Angkatan}
	}
	return res, nil
}
