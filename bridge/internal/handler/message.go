package handler

import (
	"context"
	"log"
	"regexp"
	"strings"

	"go.mau.fi/whatsmeow/types/events"

	"github.com/whatsapp-promo-poc/bridge/internal/client"
	"github.com/whatsapp-promo-poc/pkg/i18n"
)

type MessageHandler struct {
	apiClient       *client.APIClient
	i18n            *i18n.Manager
	defaultLanguage string
	sendReply       func(ctx context.Context, to, message string) error
}

func NewMessageHandler(
	apiClient *client.APIClient,
	i18nManager *i18n.Manager,
	defaultLang string,
	sendReply func(ctx context.Context, to, message string) error,
) *MessageHandler {
	return &MessageHandler{
		apiClient:       apiClient,
		i18n:            i18nManager,
		defaultLanguage: defaultLang,
		sendReply:       sendReply,
	}
}

func (h *MessageHandler) Handle(evt *events.Message) {
	ctx := context.Background()

	// Extract message text
	text := h.extractText(evt)
	if text == "" {
		return
	}

	phone := evt.Info.Sender.User
	messageID := evt.Info.ID

	log.Printf("Received message from %s: %s", phone, text)

	// Detect language from message
	language := h.detectLanguage(text)

	// Extract promo code
	promoCode := h.extractPromoCode(text)

	if promoCode == "" {
		// Check for help commands
		if h.isHelpCommand(text) {
			reply := h.i18n.Get(language, "help")
			h.sendReplyWithLog(ctx, phone, reply)
			return
		}

		// Send welcome message
		reply := h.i18n.Get(language, "welcome")
		h.sendReplyWithLog(ctx, phone, reply)
		return
	}

	// Call API to redeem
	result, err := h.apiClient.RedeemPromoCode(ctx, "+"+phone, promoCode, language, messageID)
	if err != nil {
		log.Printf("API error for %s: %v", phone, err)
		reply := h.i18n.Get(language, "error_generic")
		h.sendReplyWithLog(ctx, phone, reply)
		return
	}

	// Send appropriate response based on language
	var reply string
	if language == "en" {
		reply = result.Message.EN
	} else {
		reply = result.Message.FR
	}

	// Add emoji prefix based on status
	if result.Status == "success" {
		reply = "✅ " + reply
	} else {
		reply = "❌ " + reply
	}

	h.sendReplyWithLog(ctx, phone, reply)
}

func (h *MessageHandler) sendReplyWithLog(ctx context.Context, phone, message string) {
	log.Printf("Sending reply to %s: %s", phone, message)
	if err := h.sendReply(ctx, phone, message); err != nil {
		log.Printf("Failed to send reply to %s: %v", phone, err)
	}
}

func (h *MessageHandler) extractText(evt *events.Message) string {
	if evt.Message == nil {
		return ""
	}

	// Check various message types
	if conv := evt.Message.GetConversation(); conv != "" {
		return strings.TrimSpace(conv)
	}
	if extended := evt.Message.GetExtendedTextMessage(); extended != nil {
		return strings.TrimSpace(extended.GetText())
	}

	return ""
}

func (h *MessageHandler) extractPromoCode(text string) string {
	// Clean and uppercase
	text = strings.ToUpper(strings.TrimSpace(text))

	// Match alphanumeric codes (e.g., VALID100, PROMO2024)
	// Must be 4-20 characters, alphanumeric only
	re := regexp.MustCompile(`^[A-Z0-9]{4,20}$`)
	if re.MatchString(text) {
		return text
	}

	// Try to extract from longer message (first word that looks like a code)
	words := strings.Fields(text)
	for _, word := range words {
		cleaned := strings.ToUpper(word)
		if re.MatchString(cleaned) {
			return cleaned
		}
	}

	return ""
}

func (h *MessageHandler) detectLanguage(text string) string {
	text = strings.ToLower(text)

	// French keywords
	frenchKeywords := []string{"bonjour", "salut", "merci", "code", "promo", "aide", "svp", "s'il"}
	for _, kw := range frenchKeywords {
		if strings.Contains(text, kw) {
			return "fr"
		}
	}

	// English keywords
	englishKeywords := []string{"hello", "hi", "thanks", "help", "redeem", "please"}
	for _, kw := range englishKeywords {
		if strings.Contains(text, kw) {
			return "en"
		}
	}

	return h.defaultLanguage
}

func (h *MessageHandler) isHelpCommand(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	helpCommands := []string{"aide", "help", "?", "/help", "/aide"}
	for _, cmd := range helpCommands {
		if text == cmd {
			return true
		}
	}
	return false
}
