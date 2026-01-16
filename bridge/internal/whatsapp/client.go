package whatsapp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

type Client struct {
	wa        *whatsmeow.Client
	container *sqlstore.Container
	onMessage func(evt *events.Message)
	log       waLog.Logger
}

type Config struct {
	SessionDBPath string
	LogLevel      string
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	log := waLog.Stdout("WhatsApp", cfg.LogLevel, true)
	dbLog := waLog.Stdout("Database", cfg.LogLevel, true)

	// Create directory if it doesn't exist
	dir := filepath.Dir(cfg.SessionDBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	container, err := sqlstore.New(
		ctx,
		"sqlite3",
		fmt.Sprintf("file:%s?_foreign_keys=on", cfg.SessionDBPath),
		dbLog,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlstore: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, log)

	return &Client{
		wa:        client,
		container: container,
		log:       log,
	}, nil
}

func (c *Client) SetMessageHandler(handler func(evt *events.Message)) {
	c.onMessage = handler
}

func (c *Client) Connect(ctx context.Context) error {
	c.wa.AddEventHandler(c.eventHandler)

	if c.wa.Store.ID == nil {
		// New device - need QR code
		qrChan, err := c.wa.GetQRChannel(ctx)
		if err != nil {
			return fmt.Errorf("failed to get QR channel: %w", err)
		}

		if err := c.wa.Connect(); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		for evt := range qrChan {
			switch evt.Event {
			case "code":
				DisplayQR(evt.Code)
			case "success":
				c.log.Infof("Login successful!")
				return nil
			case "timeout":
				return fmt.Errorf("QR code timeout - please restart and try again")
			default:
				c.log.Warnf("Login event: %s", evt.Event)
			}
		}
	} else {
		// Existing device - just connect
		if err := c.wa.Connect(); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		c.log.Infof("Connected with existing session: %s", c.wa.Store.ID)
	}

	return nil
}

func (c *Client) eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		if c.onMessage != nil && !v.Info.IsFromMe {
			c.onMessage(v)
		}
	case *events.Connected:
		c.log.Infof("WhatsApp connected")
	case *events.Disconnected:
		c.log.Warnf("WhatsApp disconnected")
	case *events.LoggedOut:
		c.log.Errorf("WhatsApp logged out - session invalidated")
		os.Exit(1)
	case *events.StreamReplaced:
		c.log.Warnf("Stream replaced - another device connected")
	}
}

func (c *Client) SendText(ctx context.Context, to string, message string) error {
	jid, err := c.parseJID(to)
	if err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}

	_, err = c.wa.SendMessage(ctx, jid, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (c *Client) parseJID(phone string) (types.JID, error) {
	// Remove any non-numeric characters except leading +
	cleaned := phone
	if len(cleaned) > 0 && cleaned[0] == '+' {
		cleaned = cleaned[1:]
	}

	// Remove any remaining non-numeric characters
	result := ""
	for _, ch := range cleaned {
		if ch >= '0' && ch <= '9' {
			result += string(ch)
		}
	}

	if len(result) < 8 {
		return types.JID{}, fmt.Errorf("phone number too short: %s", phone)
	}

	return types.NewJID(result, types.DefaultUserServer), nil
}

func (c *Client) WaitForShutdown() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	c.log.Infof("Shutting down...")
	c.wa.Disconnect()
}

func (c *Client) IsConnected() bool {
	return c.wa.IsConnected()
}

func (c *Client) GetPhoneNumber() string {
	if c.wa.Store.ID != nil {
		return c.wa.Store.ID.User
	}
	return ""
}
