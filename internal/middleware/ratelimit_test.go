package middleware

import (
	"testing"
	"time"
)

func TestLoginLimiterBlocksAfterMaxFailures(t *testing.T) {
	l := NewLoginLimiter(5, time.Minute)

	for i := 0; i < 5; i++ {
		if l.Blocked("k") {
			t.Fatalf("tidak boleh terblokir pada percobaan ke-%d", i+1)
		}
		l.Fail("k")
	}
	if !l.Blocked("k") {
		t.Fatal("percobaan ke-6 harus terblokir")
	}
	if l.Blocked("lain") {
		t.Fatal("key lain tidak boleh ikut terblokir")
	}

	l.Reset("k")
	if l.Blocked("k") {
		t.Fatal("setelah reset tidak boleh terblokir")
	}
}

func TestLoginLimiterWindowExpires(t *testing.T) {
	l := NewLoginLimiter(2, 50*time.Millisecond)
	l.Fail("k")
	l.Fail("k")
	if !l.Blocked("k") {
		t.Fatal("harus terblokir")
	}
	time.Sleep(80 * time.Millisecond)
	if l.Blocked("k") {
		t.Fatal("harus terbuka setelah window lewat")
	}
}
