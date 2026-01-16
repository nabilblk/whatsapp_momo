package mock

import (
	"context"
	"strings"
)

type AntiFraudService struct {
	enabled bool
}

func NewAntiFraudService(enabled bool) *AntiFraudService {
	return &AntiFraudService{enabled: enabled}
}

type FraudCheckResult struct {
	Allowed    bool
	RiskScore  float64
	RiskReason string
}

func (a *AntiFraudService) CheckTransaction(ctx context.Context, phone, promoCode string) (*FraudCheckResult, error) {
	if !a.enabled {
		return &FraudCheckResult{Allowed: true, RiskScore: 0}, nil
	}

	riskScore := 0.0
	riskReason := ""

	// Rule 1: Check for suspicious phone patterns
	if strings.HasPrefix(phone, "+000") || strings.HasPrefix(phone, "000") {
		riskScore = 0.9
		riskReason = "Suspicious phone number pattern"
	}

	// Rule 2: Check for test/fraud keywords in code
	upperCode := strings.ToUpper(promoCode)
	if strings.Contains(upperCode, "FRAUD") || strings.Contains(upperCode, "HACK") {
		riskScore = 1.0
		riskReason = "Suspicious promo code pattern"
	}

	// Rule 3: Check for very short or very long phone numbers
	cleanPhone := strings.TrimPrefix(phone, "+")
	if len(cleanPhone) < 8 || len(cleanPhone) > 15 {
		riskScore = 0.8
		riskReason = "Invalid phone number length"
	}

	return &FraudCheckResult{
		Allowed:    riskScore < 0.8,
		RiskScore:  riskScore,
		RiskReason: riskReason,
	}, nil
}

func (a *AntiFraudService) IsEnabled() bool {
	return a.enabled
}
