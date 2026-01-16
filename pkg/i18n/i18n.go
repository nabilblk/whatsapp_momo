package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed messages/*.json
var messagesFS embed.FS

type Manager struct {
	messages map[string]map[string]string
	mu       sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

func NewManager() (*Manager, error) {
	var initErr error
	once.Do(func() {
		instance = &Manager{
			messages: make(map[string]map[string]string),
		}

		for _, lang := range []string{"en", "fr"} {
			data, err := messagesFS.ReadFile(fmt.Sprintf("messages/%s.json", lang))
			if err != nil {
				initErr = fmt.Errorf("failed to read %s.json: %w", lang, err)
				return
			}

			var msgs map[string]string
			if err := json.Unmarshal(data, &msgs); err != nil {
				initErr = fmt.Errorf("failed to parse %s.json: %w", lang, err)
				return
			}

			instance.messages[lang] = msgs
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return instance, nil
}

func (m *Manager) Get(lang, key string, args ...interface{}) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if msgs, ok := m.messages[lang]; ok {
		if msg, ok := msgs[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(msg, args...)
			}
			return msg
		}
	}

	// Fallback to French (default language)
	if lang != "fr" {
		if msgs, ok := m.messages["fr"]; ok {
			if msg, ok := msgs[key]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(msg, args...)
				}
				return msg
			}
		}
	}

	return key
}

func (m *Manager) GetBilingual(key string, args ...interface{}) (fr, en string) {
	fr = m.Get("fr", key, args...)
	en = m.Get("en", key, args...)
	return
}

func (m *Manager) SupportedLanguages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	langs := make([]string, 0, len(m.messages))
	for lang := range m.messages {
		langs = append(langs, lang)
	}
	return langs
}
