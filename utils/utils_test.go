package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO: Mock timings so that tests do not need to actually wait "N" seconds for a time-based test to finish: https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/mocking#mocking
func TestRunWhileFalse_Instant(t *testing.T) {
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
	// given
	ret := false
	f := func() bool {
		return ret
	}
	time.AfterFunc(1*time.Second, func() { ret = true })

	// when
	result := RunWhileFalse(f, 5*time.Second, time.Second)

	// then
	assert.True(t, result)
}

func TestGenerateRandomString(t *testing.T) {
	assert := assert.New(t)
	testCases := []struct {
		name            string
		inputLength     int
		expectedToPanic bool
	}{
		{"zero-length random string", 0, false},
		{"single character random string", 1, false},
		{"multi-character random string", 5, false},
		{"invalid negative input length", -1, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); tc.expectedToPanic {
					assert.NotNil(r, "expected goroutine to panic")
				} else {
					assert.Nil(r, "expected goroutine not to panic")
				}
			}()
			actualOutput := GenerateRandomString(tc.inputLength)
			assert.Len(actualOutput, tc.inputLength)
		})
	}
}

func TestGenerateRandomStringSeeded(t *testing.T) {
	randSeed := int64(1)
	assert := assert.New(t)
	testCases := []struct {
		name           string
		inputLength    int
		expectedOutput string
		expectedErr    error
	}{
		{"zero-length random string", 0, "", nil},
		{"single character random string", 1, "X", nil},
		{"multi-character random string", 5, "XVlBz", nil},
		{"invalid negative input length", -1, "", errors.New("requested length of random string must be greater than or equal to zero")},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOutput, actualError := generateRandomStringSeeded(&randSeed, tc.inputLength)
			assert.Equal(tc.expectedOutput, actualOutput)
			assert.Equal(tc.expectedErr, actualError)
		})
	}
}
