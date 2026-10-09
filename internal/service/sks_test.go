package service

import "testing"

func ptr(v float64) *float64 { return &v }

func TestBatasSKS(t *testing.T) {
	cases := []struct {
		name string
		ipk  *float64
		want int
	}{
		{"4.00", ptr(4.00), 24},
		{"3.00 batas bawah 24", ptr(3.00), 24},
		{"2.99", ptr(2.99), 21},
		{"2.50 batas bawah 21", ptr(2.50), 21},
		{"2.49", ptr(2.49), 18},
		{"0.00", ptr(0.00), 18},
		{"NULL", nil, 18},
	}
	for _, tc := range cases {
		if got := BatasSKS(tc.ipk); got != tc.want {
			t.Errorf("%s: BatasSKS = %d, want %d", tc.name, got, tc.want)
		}
	}
}
