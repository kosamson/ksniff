package utils

import (
	"context"
	"math/rand"
	"time"
)

func RunWhileFalse(fn func() bool, timeout time.Duration, delay time.Duration) bool {
	var ctx context.Context
	var cancel context.CancelFunc
	if fn() {
		return true
	}

	// Timeout 0 is infinite timeout
	if timeout == 0 {
		ctx, cancel = context.WithCancel(context.Background())
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	}
	delayTick := time.NewTicker(delay)

	defer delayTick.Stop()
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-delayTick.C:
			if fn() {
				cancel()
				return true
			}
		}
	}
}

func GenerateRandomString(length int) string {
	return generateRandomString(nil, length)
}

func generateRandomString(random *rand.Rand, length int) string {
	if length < 0 {
		panic("length of random string must be greater than or equal to zero")
	}

	if random == nil {
		random = rand.New(rand.NewSource(time.Now().Unix()))
	}

	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)

	for i := range b {
		b[i] = letterRunes[random.Intn(len(letterRunes))]
	}

	return string(b)
}
