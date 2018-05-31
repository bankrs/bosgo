package bosgo

import (
	"testing"
	"time"
)

func TestRetryPolicyWait(t *testing.T) {
	testCases := []struct {
		policy   RetryPolicy
		attempts int
		expected time.Duration
	}{
		{
			policy: RetryPolicy{
				MaxRetries: 20,
			},
			attempts: 2,
			expected: 0,
		},
		{
			policy: RetryPolicy{
				Wait:       time.Second,
				MaxRetries: 20,
			},
			attempts: 10,
			expected: time.Second,
		},
		{
			policy: RetryPolicy{
				Wait:       30 * time.Millisecond,
				MaxWait:    time.Second,
				Multiplier: 1.5,
				MaxRetries: 20,
			},
			attempts: 1,
			expected: 30 * time.Millisecond,
		},
		{
			policy: RetryPolicy{
				Wait:       30 * time.Millisecond,
				MaxWait:    time.Second,
				Multiplier: 1.5,
				MaxRetries: 20,
			},
			attempts: 2,
			expected: 45 * time.Millisecond,
		},
		{
			policy: RetryPolicy{
				Wait:       60 * time.Millisecond,
				MaxWait:    time.Second,
				Multiplier: 1.5,
				MaxRetries: 20,
			},
			attempts: 6,
			expected: 455625000,
		},
		{
			policy: RetryPolicy{
				Wait:       60 * time.Millisecond,
				MaxWait:    time.Second,
				Multiplier: 1.5,
				MaxRetries: 20,
			},
			attempts: 10,
			expected: time.Second,
		},
		{
			policy: RetryPolicy{
				Wait:       30 * time.Millisecond,
				MaxWait:    time.Second,
				Multiplier: 1.2,
				MaxRetries: 20,
			},
			attempts: 10,
			expected: 154793410,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			wait := tc.policy.NextWait(tc.attempts)
			if wait != tc.expected {
				t.Errorf("got %s (%[1]d), wanted %s (%[2]d)", wait, tc.expected)
			}
		})
	}
}

func TestRetryPolicyCanRetry(t *testing.T) {
	testCases := []struct {
		policy   RetryPolicy
		attempts int
		expected bool
	}{
		{
			policy: RetryPolicy{
				Wait:       time.Second,
				MaxRetries: 5,
			},
			attempts: 5, // 4 retries
			expected: true,
		},
		{
			policy: RetryPolicy{
				Wait:       time.Second,
				MaxRetries: 5,
			},
			attempts: 6, // 5 retries
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			allowed := tc.policy.canRetry(tc.attempts)
			if allowed != tc.expected {
				t.Errorf("got %v, wanted %v", allowed, tc.expected)
			}
		})
	}
}
