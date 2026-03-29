package marketdata

import "testing"

// assertPanics は対象関数が panic を発生させることを検証するテストヘルパー。
func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic, but did not panic")
		}
	}()
	fn()
}
