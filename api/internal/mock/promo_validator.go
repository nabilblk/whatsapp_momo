package mock

import (
	"time"

	"github.com/whatsapp-promo-poc/pkg/errors"
	"github.com/whatsapp-promo-poc/pkg/models"
)

type PromoValidator struct{}

func NewPromoValidator() *PromoValidator {
	return &PromoValidator{}
}

type ValidationResult struct {
	Valid     bool
	ErrorCode errors.ErrorCode
	Message   string
}

func (v *PromoValidator) Validate(promo *models.PromoCode) *ValidationResult {
	if promo == nil {
		return &ValidationResult{
			Valid:     false,
			ErrorCode: errors.CodeInvalid,
			Message:   "promo_invalid",
		}
	}

	// Check status
	switch promo.Status {
	case models.PromoStatusExpired:
		return &ValidationResult{
			Valid:     false,
			ErrorCode: errors.CodeExpired,
			Message:   "promo_expired",
		}
	case models.PromoStatusUsed:
		return &ValidationResult{
			Valid:     false,
			ErrorCode: errors.CodeAlreadyUsed,
			Message:   "promo_used",
		}
	case models.PromoStatusInvalid:
		return &ValidationResult{
			Valid:     false,
			ErrorCode: errors.CodeInvalid,
			Message:   "promo_invalid",
		}
	}

	// Check usage limits
	if promo.CurrentUses >= promo.MaxUses {
		return &ValidationResult{
			Valid:     false,
			ErrorCode: errors.CodeAlreadyUsed,
			Message:   "promo_used",
		}
	}

	// Check validity period
	now := time.Now()
	if promo.ValidFrom != nil && now.Before(*promo.ValidFrom) {
		return &ValidationResult{
			Valid:     false,
			ErrorCode: errors.CodeInvalid,
			Message:   "promo_invalid",
		}
	}

	if promo.ValidUntil != nil && now.After(*promo.ValidUntil) {
		return &ValidationResult{
			Valid:     false,
			ErrorCode: errors.CodeExpired,
			Message:   "promo_expired",
		}
	}

	return &ValidationResult{
		Valid:   true,
		Message: "promo_valid",
	}
}
