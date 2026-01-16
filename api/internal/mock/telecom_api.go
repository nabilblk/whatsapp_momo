package mock

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type TelecomAPI struct {
	simulatedDelayMs int
	failureRate      float64
}

func NewTelecomAPI(delayMs int) *TelecomAPI {
	return &TelecomAPI{
		simulatedDelayMs: delayMs,
		failureRate:      0.02, // 2% simulated failure rate
	}
}

type DeliveryResult struct {
	Success       bool
	TransactionID string
	ErrorCode     string
	ErrorMessage  string
}

func (t *TelecomAPI) DeliverReward(ctx context.Context, phone, rewardType string, amount float64, currency string) (*DeliveryResult, error) {
	// Simulate network delay
	select {
	case <-time.After(time.Duration(t.simulatedDelayMs) * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Simulate occasional failures
	if rand.Float64() < t.failureRate {
		return &DeliveryResult{
			Success:      false,
			ErrorCode:    "TELECOM_UNAVAILABLE",
			ErrorMessage: "Telecom service temporarily unavailable",
		}, nil
	}

	return &DeliveryResult{
		Success:       true,
		TransactionID: generateMockTxID(),
	}, nil
}

func generateMockTxID() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("TX%s", string(b))
}

func (t *TelecomAPI) CheckBalance(ctx context.Context, phone string) (float64, error) {
	// Mock implementation - always return a valid balance
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	return 1000.0, nil
}
