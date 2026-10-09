package service

import "math"

// BatasSKS menentukan batas SKS per semester berdasarkan IPK terakhir:
//   IPK >= 3,00        -> 24 SKS
//   IPK 2,50 - 2,99    -> 21 SKS
//   IPK < 2,50         -> 18 SKS
// IPK kosong (NULL) diperlakukan sebagai < 2,50 (asumsi, lihat README).
func BatasSKS(ipk *float64) int {
	if ipk == nil {
		return 18
	}
	cents := int(math.Round(*ipk * 100))
	switch {
	case cents >= 300:
		return 24
	case cents >= 250:
		return 21
	default:
		return 18
	}
}
