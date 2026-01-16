package models

import "time"

type TransactionStatus string

const (
	TxStatusPending TransactionStatus = "pending"
	TxStatusSuccess TransactionStatus = "success"
	TxStatusFailed  TransactionStatus = "failed"
	TxStatusFraud   TransactionStatus = "fraud_blocked"
)

type Transaction struct {
	ID              int64             `json:"id"`
	TransactionID   string            `json:"transaction_id"`
	PromoCodeID     *int64            `json:"promo_code_id"`
	PhoneNumber     string            `json:"phone_number"`
	PromoCodeInput  string            `json:"promo_code_input"`
	Status          TransactionStatus `json:"status"`
	ErrorCode       string            `json:"error_code,omitempty"`
	FailureReason   string            `json:"failure_reason,omitempty"`
	RewardType      string            `json:"reward_type,omitempty"`
	RewardAmount    float64           `json:"reward_amount,omitempty"`
	RewardDelivered bool              `json:"reward_delivered"`
	Language        string            `json:"language"`
	WAMessageID     string            `json:"whatsapp_message_id,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	ProcessedAt     *time.Time        `json:"processed_at,omitempty"`
}
