package two_factor_auth_test

import (
	two_factor_auth_service "compost-bin/service/two_factor_auth"
	"crypto/rand"
	"testing"
)

var (
	plains4Benchmark      = []string{}
	cipherTexts4Benchmark = []string{}
)

func TestEncryptAndDecrypt(t *testing.T) {
	mockData := []string{
		"123.com",
		"19981024",
		"you@123.com",
	}

	for i, plain := range mockData {
		cipherText := two_factor_auth_service.Encrypt(plain)
		if cipherText == "" {
			t.Fatalf("Failed to encrypt %dth plain text %s.", i, plain)
		}

		got := two_factor_auth_service.Decrypt(cipherText)
		if got == "" {
			t.Fatalf("Failed to decrypt %dth cipher text %s.", i, got)
		}

		if got != plain {
			t.Fatalf("Failed to decrypt %s, got %s, expect %s.", cipherText, got, plain)
		}
	}
}

func BenchmarkEncrypt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		plains4Benchmark = append(plains4Benchmark, rand.Text())
		b.StartTimer()
		cipherTexts4Benchmark = append(cipherTexts4Benchmark, two_factor_auth_service.Encrypt(plains4Benchmark[i]))
		b.StopTimer()
	}
}

func BenchmarkDecrypt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		idx := i % len(cipherTexts4Benchmark)
		if plains4Benchmark[idx] != two_factor_auth_service.Decrypt(cipherTexts4Benchmark[idx]) {
			b.Fatalf("Failed to decrypt")
		}
	}
}
