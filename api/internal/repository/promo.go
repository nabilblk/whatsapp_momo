package repository

import (
	"database/sql"
	"fmt"

	"github.com/whatsapp-promo-poc/pkg/models"
)

type PromoRepository struct {
	db *sql.DB
}

func NewPromoRepository(db *sql.DB) *PromoRepository {
	return &PromoRepository{db: db}
}

func (r *PromoRepository) FindByCode(code string) (*models.PromoCode, error) {
	var promo models.PromoCode
	var validFrom, validUntil sql.NullTime
	var rewardDescFR, rewardDescEN sql.NullString

	err := r.db.QueryRow(`
		SELECT id, code, reward_type, reward_amount, reward_description_fr,
		       reward_description_en, currency, status, max_uses, current_uses,
		       valid_from, valid_until, created_at, updated_at
		FROM promo_codes WHERE UPPER(code) = UPPER(?)
	`, code).Scan(
		&promo.ID, &promo.Code, &promo.RewardType, &promo.RewardAmount,
		&rewardDescFR, &rewardDescEN, &promo.Currency, &promo.Status,
		&promo.MaxUses, &promo.CurrentUses, &validFrom, &validUntil,
		&promo.CreatedAt, &promo.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find promo code: %w", err)
	}

	if validFrom.Valid {
		promo.ValidFrom = &validFrom.Time
	}
	if validUntil.Valid {
		promo.ValidUntil = &validUntil.Time
	}
	if rewardDescFR.Valid {
		promo.RewardDescriptionFR = rewardDescFR.String
	}
	if rewardDescEN.Valid {
		promo.RewardDescriptionEN = rewardDescEN.String
	}

	return &promo, nil
}

func (r *PromoRepository) IncrementUsage(id int64) error {
	_, err := r.db.Exec(`
		UPDATE promo_codes
		SET current_uses = current_uses + 1,
		    updated_at = CURRENT_TIMESTAMP,
		    status = CASE
		        WHEN current_uses + 1 >= max_uses THEN 'used'
		        ELSE status
		    END
		WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("failed to increment promo usage: %w", err)
	}

	return nil
}

func (r *PromoRepository) GetAll() ([]*models.PromoCode, error) {
	rows, err := r.db.Query(`
		SELECT id, code, reward_type, reward_amount, status, max_uses, current_uses
		FROM promo_codes
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all promo codes: %w", err)
	}
	defer rows.Close()

	var promos []*models.PromoCode
	for rows.Next() {
		var p models.PromoCode
		if err := rows.Scan(&p.ID, &p.Code, &p.RewardType, &p.RewardAmount, &p.Status, &p.MaxUses, &p.CurrentUses); err != nil {
			return nil, fmt.Errorf("failed to scan promo code: %w", err)
		}
		promos = append(promos, &p)
	}

	return promos, nil
}
