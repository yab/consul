// Package retry provides a generic retry mechanism
// which can be used in tests.
package retry

import "time"

const (
	// Timeout is the default time span for which an operation
	// should be retried.
	Timeout = time.Second

	// Wait is the time between two retry attempts.
	Wait = 25 * time.Millisecond
)

// Retryer provides an interface for retrying an operation
// repeatedly until it either succeeds or times out. The
// Failer will be called when on timeout.
//
// Retryer does not accept a callback function to execute
// the tests to keep the file:line information generated
// by test output correct.
//
//   func TestX(t *testing.T) {
//       for r := retry.For{}{}; r.Next(); {
// 		     if err := f(); err != nil {
// 			     t.Log("f: ", err)
// 			     continue
// 		     }
// 		     break
// 	     }
//   }
type Retryer interface {
	// Next returns true if the operation can be retried.
	// If not, it will call t.FailNow() and return false.
	Next(t Failer) bool

	// Reset configures the retryer for re-use.
	Reset()
}

// Failer is an interface compatible with *testing.T.
type Failer interface {
	FailNow()
}

// R returns a Timer with default configuration.
func R() *Timer {
	return &Timer{Timeout: Timeout, Wait: Wait}
}

// Times returns a Counter with default configuration.
func Times(n int) *Counter {
	return &Counter{Count: n, Wait: Wait}
}

// Counter implements a retryer which retries
// an operation a specicific number of times.
// The first operation will be executed immediately
// and all subsequent operations will return after
// the wait period.
type Counter struct {
	Count int
	Wait  time.Duration

	count int
}

// Next returns true as long as the number
// of retries has not been reached. The
// first invocation will return immediately.
// All subsequent calls will return after the
// Wait period.
func (r *Counter) Next(t Failer) bool {
	if r.count == r.Count {
		t.FailNow()
		return false
	}
	if r.count > 0 {
		time.Sleep(r.Wait)
	}
	r.count++
	return true
}

// Reset configures the retryer for re-use.
func (r *Counter) Reset() {
	r.count = 0
}

// Timer implements a time-based retryer
// which iterates until a certain amount of
// time has elapsed. Between iterations it
// will wait the time set in Wait.
type Timer struct {
	Timeout time.Duration
	Wait    time.Duration

	// stop is the timeout deadline.
	// Set on the first invocation of Next().
	stop time.Time
}

// Next returns true as long as the timeout
// has not elapsed. The first invocation will
// set the deadline for the timeout and
// will return immediately. All subsequent
// calls will return after the Wait period.
func (r *Timer) Next(t Failer) bool {
	if r.stop.IsZero() {
		r.stop = time.Now().Add(r.Timeout)
		return true
	}
	if time.Now().After(r.stop) {
		t.FailNow()
		return false
	}
	time.Sleep(r.Wait)
	return true
}

// Reset configures the retryer for re-use.
func (r *Timer) Reset() {
	r.stop = time.Time{}
}
