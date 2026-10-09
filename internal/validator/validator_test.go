package validator

import "testing"

func TestIsNIM(t *testing.T) {
	cases := map[string]bool{
		"187221000001":  true,
		"18722100000":   false,
		"1872210000012": false,
		"18722100000a":  false,
		"":              false,
	}
	for in, want := range cases {
		if got := IsNIM(in); got != want {
			t.Errorf("IsNIM(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsEmail(t *testing.T) {
	cases := map[string]bool{
		"a@b.com":            true,
		"rina@student.test":  true,
		"bukan-email":        false,
		"Nama <a@b.com>":     false,
		"a b@c.com":          false,
		"":                   false,
	}
	for in, want := range cases {
		if got := IsEmail(in); got != want {
			t.Errorf("IsEmail(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsTahunAkademik(t *testing.T) {
	cases := map[string]bool{
		"2026/2027-Ganjil": true,
		"2026/2027-Genap":  true,
		"2026/2028-Ganjil": false,
		"2026/2027-ganjil": false,
		"2026-2027-Ganjil": false,
		"2026/2027":        false,
		"":                 false,
	}
	for in, want := range cases {
		if got := IsTahunAkademik(in); got != want {
			t.Errorf("IsTahunAkademik(%q) = %v, want %v", in, got, want)
		}
	}
}
