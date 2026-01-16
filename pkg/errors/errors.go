package errors

import "fmt"

type ErrorCode string

const (
	CodeValid       ErrorCode = "CODE_VALID"
	CodeInvalid     ErrorCode = "CODE_INVALID"
	CodeExpired     ErrorCode = "CODE_EXPIRED"
	CodeAlreadyUsed ErrorCode = "CODE_ALREADY_USED"
	CodeNotEligible ErrorCode = "CODE_NOT_ELIGIBLE"
	RateLimited     ErrorCode = "RATE_LIMITED"
	RewardFailed    ErrorCode = "REWARD_FAILED"
	FraudBlocked    ErrorCode = "FRAUD_BLOCKED"
	SystemError     ErrorCode = "SYSTEM_ERROR"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewError(code ErrorCode, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func WrapError(code ErrorCode, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

func NewInvalidCodeError() *AppError {
	return &AppError{Code: CodeInvalid, Message: "Invalid promo code"}
}

func NewExpiredCodeError() *AppError {
	return &AppError{Code: CodeExpired, Message: "Promo code has expired"}
}

func NewAlreadyUsedError() *AppError {
	return &AppError{Code: CodeAlreadyUsed, Message: "Promo code has already been used"}
}

func NewRateLimitError() *AppError {
	return &AppError{Code: RateLimited, Message: "Rate limit exceeded"}
}

func NewFraudBlockedError() *AppError {
	return &AppError{Code: FraudBlocked, Message: "Transaction blocked for security reasons"}
}

func NewRewardFailedError(err error) *AppError {
	return &AppError{Code: RewardFailed, Message: "Failed to deliver reward", Err: err}
}

func NewSystemError(err error) *AppError {
	return &AppError{Code: SystemError, Message: "Internal system error", Err: err}
}
