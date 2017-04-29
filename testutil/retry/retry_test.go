package retry

import (
	"testing"
	"time"
)

type failer struct{ calls int }

func (f *failer) FailNow() { f.calls++ }

// delta defines the time band a test run should complete in.
var delta = 5 * time.Millisecond

func TestRetryer(t *testing.T) {
	tests := []struct {
		desc string
		r    Retryer
	}{
		{"counter", &Counter{Count: 3, Wait: 10 * time.Millisecond}},
		{"timer", &Timer{Timeout: 20 * time.Millisecond, Wait: 10 * time.Millisecond}},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var n int
			f := new(failer)
			start := time.Now()
			for tt.r.Next(f) {
				n++
			}
			dur := time.Since(start)
			if got, want := n, 3; got != want {
				t.Fatalf("got %d retries want %d", got, want)
			}
			if got, want := f.calls, 1; got != want {
				t.Fatalf("got %d FailNow calls want %d", got, want)
			}
			// since the first iteration happens immediately
			// the retryer waits only twice for three iterations.
			// order of events: (true, wait, true, wait, true, false)
			if got, want := dur, 20*time.Millisecond; got < (want-delta) || got > (want+delta) {
				t.Fatalf("loop took %v want %v (+/- %v)", got, want, delta)
			}
		})
	}
}
