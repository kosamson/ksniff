package utils

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString(t *testing.T) {
	t.Parallel()

	randSeed := int64(1)
	assert := assert.New(t)
	testCases := []struct {
		name            string
		inputLength     int
		expectedOutput  string
		expectedToPanic bool
	}{
		{"zero-length random string", 0, "", false},
		{"single character random string", 1, "X", false},
		{"multi-character random string", 5, "XVlBz", false},
		{"invalid negative input length", -1, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				if r := recover(); tc.expectedToPanic {
					assert.NotNil(r, "expected goroutine to panic")
				} else {
					assert.Nil(r, "expected goroutine not to panic")
				}
			}()

			actualOutput := generateRandomString(rand.New(rand.NewSource(randSeed)), tc.inputLength)

			assert.Equal(tc.expectedOutput, actualOutput)
		})
	}
}
