package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type RateLimitRepository struct {
	db *sql.DB
}

func NewRateLimitRepository(db *sql.DB) *RateLimitRepository {
	return &RateLimitRepository{db: db}
}

type RateLimitResult struct {
	MinuteCount  int
	HourCount    int
	MinuteWindow time.Time
	HourWindow   time.Time
}

func (r *RateLimitRepository) CheckAndIncrement(phone string, minuteLimit, hourLimit int) (*RateLimitResult, bool, error) {
	now := time.Now().UTC()
	minuteWindow := now.Truncate(time.Minute)
	hourWindow := now.Truncate(time.Hour)

	tx, err := r.db.Begin()
	if err != nil {
		return nil, false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Upsert minute counter
	minuteCount, err := r.upsertCounter(tx, phone, "minute", minuteWindow)
	if err != nil {
		return nil, false, err
	}

	// Upsert hour counter
	hourCount, err := r.upsertCounter(tx, phone, "hour", hourWindow)
	if err != nil {
		return nil, false, err
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result := &RateLimitResult{
		MinuteCount:  minuteCount,
		HourCount:    hourCount,
		MinuteWindow: minuteWindow,
		HourWindow:   hourWindow,
	}

	allowed := minuteCount <= minuteLimit && hourCount <= hourLimit
	return result, allowed, nil
}

func (r *RateLimitRepository) upsertCounter(tx *sql.Tx, phone, windowType string, windowStart time.Time) (int, error) {
	// First try to update existing record
	result, err := tx.Exec(`
		UPDATE rate_limits
		SET request_count = CASE
			WHEN window_start < ? THEN 1
			ELSE request_count + 1
		END,
		window_start = CASE
			WHEN window_start < ? THEN ?
			ELSE window_start
		END
		WHERE phone_number = ? AND window_type = ?
	`, windowStart, windowStart, windowStart, phone, windowType)

	if err != nil {
		return 0, fmt.Errorf("failed to update rate limit: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Insert new record
		_, err = tx.Exec(`
			INSERT INTO rate_limits (phone_number, window_start, request_count, window_type)
			VALUES (?, ?, 1, ?)
		`, phone, windowStart, windowType)
		if err != nil {
			return 0, fmt.Errorf("failed to insert rate limit: %w", err)
		}
		return 1, nil
	}

	// Get the current count
	var count int
	err = tx.QueryRow(`
		SELECT request_count FROM rate_limits
		WHERE phone_number = ? AND window_type = ?
	`, phone, windowType).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	return count, nil
}

func (r *RateLimitRepository) GetLimits(phone string) (*RateLimitResult, error) {
	now := time.Now().UTC()
	minuteWindow := now.Truncate(time.Minute)
	hourWindow := now.Truncate(time.Hour)

	result := &RateLimitResult{
		MinuteWindow: minuteWindow,
		HourWindow:   hourWindow,
	}

	// Get minute count
	var minuteCount int
	var minuteStart time.Time
	err := r.db.QueryRow(`
		SELECT request_count, window_start FROM rate_limits
		WHERE phone_number = ? AND window_type = 'minute'
	`, phone).Scan(&minuteCount, &minuteStart)

	if err == nil && !minuteStart.Before(minuteWindow) {
		result.MinuteCount = minuteCount
	}

	// Get hour count
	var hourCount int
	var hourStart time.Time
	err = r.db.QueryRow(`
		SELECT request_count, window_start FROM rate_limits
		WHERE phone_number = ? AND window_type = 'hour'
	`, phone).Scan(&hourCount, &hourStart)

	if err == nil && !hourStart.Before(hourWindow) {
		result.HourCount = hourCount
	}

	return result, nil
}
