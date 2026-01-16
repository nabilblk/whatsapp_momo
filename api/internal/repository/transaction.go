package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/whatsapp-promo-poc/pkg/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *models.Transaction) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO transactions (
			transaction_id, promo_code_id, phone_number, promo_code_input,
			status, error_code, failure_reason, reward_type, reward_amount,
			reward_delivered, language, whatsapp_message_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		tx.TransactionID, tx.PromoCodeID, tx.PhoneNumber, tx.PromoCodeInput,
		tx.Status, tx.ErrorCode, tx.FailureReason, tx.RewardType, tx.RewardAmount,
		tx.RewardDelivered, tx.Language, tx.WAMessageID,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (r *TransactionRepository) UpdateStatus(txID string, status models.TransactionStatus, errorCode, failureReason string) error {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE transactions
		SET status = ?, error_code = ?, failure_reason = ?, processed_at = ?
		WHERE transaction_id = ?
	`, status, errorCode, failureReason, now, txID)

	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}

func (r *TransactionRepository) UpdateStatusWithDelivery(txID string, status models.TransactionStatus, delivered bool) error {
	now := time.Now()
	_, err := r.db.Exec(`
		UPDATE transactions
		SET status = ?, reward_delivered = ?, processed_at = ?
		WHERE transaction_id = ?
	`, status, delivered, now, txID)

	if err != nil {
		return fmt.Errorf("failed to update transaction with delivery: %w", err)
	}

	return nil
}

func (r *TransactionRepository) FindByID(txID string) (*models.Transaction, error) {
	var tx models.Transaction
	var promoCodeID sql.NullInt64
	var processedAt sql.NullTime
	var errorCode, failureReason, rewardType, waMessageID sql.NullString
	var rewardAmount sql.NullFloat64

	err := r.db.QueryRow(`
		SELECT id, transaction_id, promo_code_id, phone_number, promo_code_input,
		       status, error_code, failure_reason, reward_type, reward_amount,
		       reward_delivered, language, whatsapp_message_id, created_at, processed_at
		FROM transactions WHERE transaction_id = ?
	`, txID).Scan(
		&tx.ID, &tx.TransactionID, &promoCodeID, &tx.PhoneNumber, &tx.PromoCodeInput,
		&tx.Status, &errorCode, &failureReason, &rewardType, &rewardAmount,
		&tx.RewardDelivered, &tx.Language, &waMessageID, &tx.CreatedAt, &processedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}

	if promoCodeID.Valid {
		tx.PromoCodeID = &promoCodeID.Int64
	}
	if processedAt.Valid {
		tx.ProcessedAt = &processedAt.Time
	}
	if errorCode.Valid {
		tx.ErrorCode = errorCode.String
	}
	if failureReason.Valid {
		tx.FailureReason = failureReason.String
	}
	if rewardType.Valid {
		tx.RewardType = rewardType.String
	}
	if rewardAmount.Valid {
		tx.RewardAmount = rewardAmount.Float64
	}
	if waMessageID.Valid {
		tx.WAMessageID = waMessageID.String
	}

	return &tx, nil
}

func (r *TransactionRepository) CountByPhone(phone string, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM transactions
		WHERE phone_number = ? AND created_at >= ?
	`, phone, since).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return count, nil
}
