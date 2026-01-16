package models

import "time"

type PromoStatus string

const (
	PromoStatusActive  PromoStatus = "active"
	PromoStatusExpired PromoStatus = "expired"
	PromoStatusUsed    PromoStatus = "used"
	PromoStatusInvalid PromoStatus = "invalid"
)

type PromoCode struct {
	ID                  int64       `json:"id"`
	Code                string      `json:"code"`
	RewardType          string      `json:"reward_type"`
	RewardAmount        float64     `json:"reward_amount"`
	RewardDescriptionFR string      `json:"reward_description_fr"`
	RewardDescriptionEN string      `json:"reward_description_en"`
	Currency            string      `json:"currency"`
	Status              PromoStatus `json:"status"`
	MaxUses             int         `json:"max_uses"`
	CurrentUses         int         `json:"current_uses"`
	ValidFrom           *time.Time  `json:"valid_from"`
	ValidUntil          *time.Time  `json:"valid_until"`
	CreatedAt           time.Time   `json:"created_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
}
