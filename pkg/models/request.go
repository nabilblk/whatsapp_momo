package models

type RedeemRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Code        string `json:"code" binding:"required"`
	MessageID   string `json:"message_id,omitempty"`
	Language    string `json:"language,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

type RedeemResponse struct {
	Status        string  `json:"status"`
	TransactionID string  `json:"transaction_id,omitempty"`
	ErrorCode     string  `json:"error_code,omitempty"`
	Reward        *Reward `json:"reward,omitempty"`
	Message       Message `json:"message"`
}

type Reward struct {
	Type        string  `json:"type"`
	Amount      string  `json:"amount"`
	Description string  `json:"description"`
}

type Message struct {
	FR string `json:"fr"`
	EN string `json:"en"`
}

type HealthResponse struct {
	Status            string `json:"status"`
	WhatsAppConnected bool   `json:"whatsapp_connected"`
	DBConnected       bool   `json:"db_connected"`
	UptimeSeconds     int64  `json:"uptime_seconds"`
}
