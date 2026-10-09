package validator

import (
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

// Errors berformat sama dengan field "errors" pada response 422 di PDF.
type Errors map[string][]string

func (e Errors) Add(field, message string) { e[field] = append(e[field], message) }
func (e Errors) Any() bool                 { return len(e) > 0 }

var (
	nimRe   = regexp.MustCompile(`^[0-9]{12}$`)
	tahunRe = regexp.MustCompile(`^([0-9]{4})/([0-9]{4})-(Ganjil|Genap)$`)
)

func IsEmail(s string) bool {
	if s == "" || strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s
}

func IsNIM(s string) bool { return nimRe.MatchString(s) }

// IsTahunAkademik memeriksa format 2026/2027-Ganjil (tahun kedua = tahun pertama + 1).
func IsTahunAkademik(s string) bool {
	m := tahunRe.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	return b == a+1
}
