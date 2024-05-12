package utils

import (
	"context"
	"errors"
	"fmt"
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
	if randStr, err := generateRandomStringSeeded(nil, length); err != nil {
		panic(fmt.Sprintf("could not generate random string: %v", err))
	} else {
		return randStr
	}
}

func generateRandomStringSeeded(seed *int64, length int) (string, error) {
	if length < 0 {
		return "", errors.New("requested length of random string must be greater than or equal to zero")
	}
	var randomIntGenerator func(int) int
	if seed != nil {
		randomIntGenerator = rand.New(rand.NewSource(*seed)).Intn
	} else {
		randomIntGenerator = rand.Intn
	}
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[randomIntGenerator(len(letterRunes))]
	}
	return string(b), nil
}
