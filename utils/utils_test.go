package utils

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO: Mock timings so that tests do not need to actually wait "N" seconds for a time-based test to finish: https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking#mocking
func TestRunWhileFalse_Instant(t *testing.T) {
	t.Parallel()

	// given
	f := func() bool {
		return true
	}

	// when
	result := RunWhileFalse(f, time.Minute, time.Minute)

	// then
	assert.True(t, result)
}

func TestRunWhileFalse_1SecTimeoutFalse(t *testing.T) {
	t.Parallel()

	// given
	f := func() bool {
		return false
	}

	// when
	begin := time.Now()
	result := RunWhileFalse(f, time.Second, time.Second)
	end := time.Now()
	diff := end.Sub(begin)

	// then
	assert.False(t, result)
	assert.True(t, (diff.Seconds() > 0 && diff.Seconds() < 2))
}

func TestRunWhileFalse_NoTimeout(t *testing.T) {
	t.Parallel()

	// given
	f := func() bool {
		return false
	}
	// This part is tricky since we don't want our test case to run forever.
	// Adding a timeout outside scope of RunWhileFalse
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// when
	go func() {
		RunWhileFalse(f, 0*time.Second, time.Second)
		cancel()
	}()

	// then
	<-ctx.Done()
	assert.Equal(t, context.DeadlineExceeded, ctx.Err())
}

func TestRunWhileFalse_1SecTimeoutTrue(t *testing.T) {
	t.Parallel()

	// without the mutex, this raises a
	// gorace data race error
	var mu sync.Mutex

	ret := false
	f := func() bool {
		mu.Lock()
		defer mu.Unlock()
		return ret
	}

	time.AfterFunc(1*time.Second, func() {
		mu.Lock()
		defer mu.Unlock()
		ret = true
	})

	result := RunWhileFalse(f, 5*time.Second, time.Second)

	assert.True(t, result)
}
