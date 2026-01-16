package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/whatsapp-promo-poc/api/internal/mock"
	"github.com/whatsapp-promo-poc/api/internal/repository"
	"github.com/whatsapp-promo-poc/pkg/errors"
	"github.com/whatsapp-promo-poc/pkg/i18n"
	"github.com/whatsapp-promo-poc/pkg/models"
)

type PromoService struct {
	promoRepo     *repository.PromoRepository
	txRepo        *repository.TransactionRepository
	rateLimitRepo *repository.RateLimitRepository
	validator     *mock.PromoValidator
	telecom       *mock.TelecomAPI
	antifraud     *mock.AntiFraudService
	i18n          *i18n.Manager
	minuteLimit   int
	hourLimit     int
}

func NewPromoService(
	db *sql.DB,
	i18nManager *i18n.Manager,
	telecomDelayMs int,
	antifraudEnabled bool,
	minuteLimit, hourLimit int,
) *PromoService {
	return &PromoService{
		promoRepo:     repository.NewPromoRepository(db),
		txRepo:        repository.NewTransactionRepository(db),
		rateLimitRepo: repository.NewRateLimitRepository(db),
		validator:     mock.NewPromoValidator(),
		telecom:       mock.NewTelecomAPI(telecomDelayMs),
		antifraud:     mock.NewAntiFraudService(antifraudEnabled),
		i18n:          i18nManager,
		minuteLimit:   minuteLimit,
		hourLimit:     hourLimit,
	}
}

type RedeemResult struct {
	Success       bool
	TransactionID string
	ErrorCode     errors.ErrorCode
	MessageKey    string
	MessageFR     string
	MessageEN     string
	RewardType    string
	RewardAmount  float64
	RewardDesc    string
	Currency      string
	WaitMinutes   int
}

func (s *PromoService) RedeemCode(ctx context.Context, phone, code, language, messageID string) (*RedeemResult, error) {
	txID := generateTransactionID()

	// Step 1: Rate limiting check
	_, allowed, err := s.rateLimitRepo.CheckAndIncrement(phone, s.minuteLimit, s.hourLimit)
	if err != nil {
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	if !allowed {
		// Create failed transaction record
		tx := &models.Transaction{
			TransactionID:  txID,
			PhoneNumber:    phone,
			PromoCodeInput: code,
			Status:         models.TxStatusFailed,
			ErrorCode:      string(errors.RateLimited),
			FailureReason:  "Rate limit exceeded",
			Language:       language,
			WAMessageID:    messageID,
		}
		s.txRepo.Create(tx)

		return &RedeemResult{
			Success:       false,
			TransactionID: txID,
			ErrorCode:     errors.RateLimited,
			MessageKey:    "rate_limited",
			MessageFR:     s.i18n.Get("fr", "rate_limited", 1),
			MessageEN:     s.i18n.Get("en", "rate_limited", 1),
			WaitMinutes:   1,
		}, nil
	}

	// Step 2: Anti-fraud check
	fraudResult, err := s.antifraud.CheckTransaction(ctx, phone, code)
	if err != nil {
		return nil, fmt.Errorf("fraud check failed: %w", err)
	}

	if !fraudResult.Allowed {
		tx := &models.Transaction{
			TransactionID:  txID,
			PhoneNumber:    phone,
			PromoCodeInput: code,
			Status:         models.TxStatusFraud,
			ErrorCode:      string(errors.FraudBlocked),
			FailureReason:  fraudResult.RiskReason,
			Language:       language,
			WAMessageID:    messageID,
		}
		s.txRepo.Create(tx)

		return &RedeemResult{
			Success:       false,
			TransactionID: txID,
			ErrorCode:     errors.FraudBlocked,
			MessageKey:    "fraud_blocked",
			MessageFR:     s.i18n.Get("fr", "fraud_blocked"),
			MessageEN:     s.i18n.Get("en", "fraud_blocked"),
		}, nil
	}

	// Step 3: Find promo code
	promo, err := s.promoRepo.FindByCode(code)
	if err != nil {
		return nil, fmt.Errorf("failed to find promo code: %w", err)
	}

	if promo == nil {
		tx := &models.Transaction{
			TransactionID:  txID,
			PhoneNumber:    phone,
			PromoCodeInput: code,
			Status:         models.TxStatusFailed,
			ErrorCode:      string(errors.CodeInvalid),
			FailureReason:  "Code not found",
			Language:       language,
			WAMessageID:    messageID,
		}
		s.txRepo.Create(tx)

		return &RedeemResult{
			Success:       false,
			TransactionID: txID,
			ErrorCode:     errors.CodeInvalid,
			MessageKey:    "promo_not_found",
			MessageFR:     s.i18n.Get("fr", "promo_not_found"),
			MessageEN:     s.i18n.Get("en", "promo_not_found"),
		}, nil
	}

	// Step 4: Validate promo code
	validationResult := s.validator.Validate(promo)
	if !validationResult.Valid {
		tx := &models.Transaction{
			TransactionID:  txID,
			PromoCodeID:    &promo.ID,
			PhoneNumber:    phone,
			PromoCodeInput: code,
			Status:         models.TxStatusFailed,
			ErrorCode:      string(validationResult.ErrorCode),
			FailureReason:  validationResult.Message,
			Language:       language,
			WAMessageID:    messageID,
		}
		s.txRepo.Create(tx)

		var msgFR, msgEN string
		if validationResult.ErrorCode == errors.CodeExpired {
			if promo.ValidUntil != nil {
				dateStr := promo.ValidUntil.Format("02/01/2006")
				msgFR = s.i18n.Get("fr", "promo_expired", dateStr)
				msgEN = s.i18n.Get("en", "promo_expired", dateStr)
			} else {
				msgFR = s.i18n.Get("fr", "promo_expired_simple")
				msgEN = s.i18n.Get("en", "promo_expired_simple")
			}
		} else {
			msgFR = s.i18n.Get("fr", validationResult.Message)
			msgEN = s.i18n.Get("en", validationResult.Message)
		}

		return &RedeemResult{
			Success:       false,
			TransactionID: txID,
			ErrorCode:     validationResult.ErrorCode,
			MessageKey:    validationResult.Message,
			MessageFR:     msgFR,
			MessageEN:     msgEN,
		}, nil
	}

	// Step 5: Create pending transaction
	tx := &models.Transaction{
		TransactionID:  txID,
		PromoCodeID:    &promo.ID,
		PhoneNumber:    phone,
		PromoCodeInput: code,
		Status:         models.TxStatusPending,
		RewardType:     promo.RewardType,
		RewardAmount:   promo.RewardAmount,
		Language:       language,
		WAMessageID:    messageID,
	}

	_, err = s.txRepo.Create(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Step 6: Deliver reward via telecom API
	deliveryResult, err := s.telecom.DeliverReward(ctx, phone, promo.RewardType, promo.RewardAmount, promo.Currency)
	if err != nil {
		s.txRepo.UpdateStatus(txID, models.TxStatusFailed, string(errors.RewardFailed), err.Error())
		return nil, fmt.Errorf("telecom delivery failed: %w", err)
	}

	if !deliveryResult.Success {
		s.txRepo.UpdateStatus(txID, models.TxStatusFailed, string(errors.RewardFailed), deliveryResult.ErrorMessage)

		return &RedeemResult{
			Success:       false,
			TransactionID: txID,
			ErrorCode:     errors.RewardFailed,
			MessageKey:    "reward_failed",
			MessageFR:     s.i18n.Get("fr", "reward_failed"),
			MessageEN:     s.i18n.Get("en", "reward_failed"),
		}, nil
	}

	// Step 7: Mark success and increment usage
	s.txRepo.UpdateStatusWithDelivery(txID, models.TxStatusSuccess, true)
	s.promoRepo.IncrementUsage(promo.ID)

	// Prepare success message
	rewardDesc := promo.RewardDescriptionFR
	if language == "en" && promo.RewardDescriptionEN != "" {
		rewardDesc = promo.RewardDescriptionEN
	}

	var msgFR, msgEN string
	if promo.RewardType == "data" {
		msgFR = s.i18n.Get("fr", "promo_success_data", promo.RewardDescriptionFR)
		msgEN = s.i18n.Get("en", "promo_success_data", promo.RewardDescriptionEN)
	} else if promo.RewardType == "credit" {
		msgFR = s.i18n.Get("fr", "promo_success_credit", promo.RewardDescriptionFR)
		msgEN = s.i18n.Get("en", "promo_success_credit", promo.RewardDescriptionEN)
	} else {
		msgFR = s.i18n.Get("fr", "promo_success_generic", fmt.Sprintf("%.0f", promo.RewardAmount), promo.Currency)
		msgEN = s.i18n.Get("en", "promo_success_generic", fmt.Sprintf("%.0f", promo.RewardAmount), promo.Currency)
	}

	return &RedeemResult{
		Success:       true,
		TransactionID: txID,
		MessageKey:    "promo_success",
		MessageFR:     msgFR,
		MessageEN:     msgEN,
		RewardType:    promo.RewardType,
		RewardAmount:  promo.RewardAmount,
		RewardDesc:    rewardDesc,
		Currency:      promo.Currency,
	}, nil
}

func generateTransactionID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("txn_%s", hex.EncodeToString(b))
}
