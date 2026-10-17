package utilities

import "testing"

func TestLooksLikeUUID(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"00000000-0000-0000-0000-000000000000", true},
		{"A1B2C3D4-E5F6-7890-ABCD-EF1234567890", true},
		{"a1b2c3d4-e5f6-7890-abcd-ef1234567890", true},
		{"my-pool", false},
		{"pool-with-dashes", false},
		{"", false},
		{"00000000-0000-0000-0000-00000000000", false},
		{"not-a-uuid-at-all", false},
		{"00000000000000000000000000000000", false},
	}
	n := len(cases)
	for i := 0; i < n; i++ {
		c := cases[i]
		got := LooksLikeUUID(c.in)
		if got != c.want {
			t.Fatalf("LooksLikeUUID(%q)=%v want %v", c.in, got, c.want)
		}
	}
}

func BenchmarkLooksLikeUUID(b *testing.B) {
	ref := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = LooksLikeUUID(ref)
	}
}
