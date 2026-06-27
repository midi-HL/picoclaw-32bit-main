package tts

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// MimoTTSVariant represents which MiMo TTS model variant to use.
type MimoTTSVariant string

const (
	MimoTTSVariantPreset     MimoTTSVariant = "mimo-v2.5-tts"
	MimoTTSVariantVoiceDesign MimoTTSVariant = "mimo-v2.5-tts-voicedesign"
	MimoTTSVariantVoiceClone  MimoTTSVariant = "mimo-v2.5-tts-voiceclone"
)

// MimoTTSOptions holds configuration for MiMo TTS synthesis.
type MimoTTSOptions struct {
	// Variant selects the model variant (preset, voicedesign, voiceclone).
	Variant MimoTTSVariant
	// Voice is the preset voice name (e.g. "冰糖", "Chloe").
	// Only used for the preset variant.
	Voice string
	// VoiceDesignText is the voice description for voicedesign variant.
	VoiceDesignText string
	// VoiceCloneData is the base64-encoded reference audio for voiceclone variant.
	// Format: "data:audio/mpeg;base64,..." or "data:audio/wav;base64,..."
	VoiceCloneData string
}

type MimoTTSProvider struct {
	apiKey     string
	apiBase    string
	model      string
	options    MimoTTSOptions
	httpClient *http.Client
}

func NewMimoTTSProvider(apiKey string, apiBase string, model string, proxyURL string, opts ...MimoTTSOptions) *MimoTTSProvider {
	if apiBase == "" {
		apiBase = "https://api.xiaomimimo.com/v1/chat/completions"
	} else {
		if u, err := url.Parse(apiBase); err == nil && u.Scheme != "" && u.Host != "" {
			path := u.Path
			if u.Host == "api.xiaomimimo.com" {
				if path == "" || path == "/" || path == "/v1" || path == "/v1/" {
					path = "/v1/chat/completions"
				} else {
					if !strings.HasPrefix(path, "/") {
						path = "/" + path
					}
					if !strings.HasPrefix(path, "/v1/") {
						path = "/v1" + strings.TrimSuffix(path, "/")
					}
					if !strings.HasSuffix(path, "/chat/completions") {
						path = strings.TrimSuffix(path, "/") + "/chat/completions"
					}
				}
			} else {
				if !strings.HasSuffix(path, "/chat/completions") {
					path = strings.TrimSuffix(path, "/") + "/chat/completions"
				}
			}
			u.Path = path
			apiBase = u.String()
		} else {
			if apiBase == "https://api.xiaomimimo.com/v1" {
				apiBase = "https://api.xiaomimimo.com/v1/chat/completions"
			} else if !strings.HasSuffix(apiBase, "/chat/completions") {
				apiBase = strings.TrimSuffix(apiBase, "/") + "/chat/completions"
			}
		}
	}

	model = strings.TrimSpace(model)
	if model == "" {
		model = "mimo-v2.5-tts"
	}

	options := MimoTTSOptions{
		Variant: MimoTTSVariantPreset,
		Voice:   "mimo_default",
	}
	if len(opts) > 0 {
		options = opts[0]
		if options.Variant == "" {
			options.Variant = MimoTTSVariantPreset
		}
		if options.Voice == "" && options.Variant == MimoTTSVariantPreset {
			options.Voice = "mimo_default"
		}
	}

	client := &http.Client{Timeout: 60 * time.Second}
	if proxyURL != "" {
		if pURL, err := url.Parse(proxyURL); err == nil {
			client.Transport = &http.Transport{Proxy: http.ProxyURL(pURL)}
		}
	}

	return &MimoTTSProvider{
		apiKey:     apiKey,
		apiBase:    apiBase,
		model:      model,
		options:    options,
		httpClient: client,
	}
}

func (t *MimoTTSProvider) Name() string {
	return "mimo-tts"
}

func (t *MimoTTSProvider) Synthesize(ctx context.Context, text string) (io.ReadCloser, error) {
	logger.DebugCF("voice-tts", "Starting TTS synthesis", map[string]any{
		"text_len": len(text),
		"provider": t.Name(),
		"variant":  string(t.options.Variant),
	})

	// Build messages based on variant
	messages := make([]map[string]string, 0, 2)

	switch t.options.Variant {
	case MimoTTSVariantVoiceDesign:
		// voicedesign: user message = voice description (required)
		designText := strings.TrimSpace(t.options.VoiceDesignText)
		if designText == "" {
			designText = "A natural, clear voice."
		}
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": designText,
		})

	case MimoTTSVariantVoiceClone:
		// voiceclone: user message = empty string
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": "",
		})

	default:
		// preset: user message = style instruction (optional)
		// We don't add a user message by default for preset voices.
	}

	// assistant message = text to synthesize
	messages = append(messages, map[string]string{
		"role":    "assistant",
		"content": text,
	})

	// Build audio config based on variant
	audioConfig := map[string]string{
		"format": "mp3",
	}

	switch t.options.Variant {
	case MimoTTSVariantPreset:
		if t.options.Voice != "" {
			audioConfig["voice"] = t.options.Voice
		}

	case MimoTTSVariantVoiceClone:
		if t.options.VoiceCloneData != "" {
			audioConfig["voice"] = t.options.VoiceCloneData
		}

	case MimoTTSVariantVoiceDesign:
		// No voice field for voicedesign
	}

	reqBody := map[string]any{
		"model":    t.model,
		"messages": messages,
		"audio":    audioConfig,
		"stream":   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.apiBase, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Choices []struct {
			Message struct {
				Audio struct {
					Data string `json:"data"`
				} `json:"audio"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(payload.Choices) == 0 || payload.Choices[0].Message.Audio.Data == "" {
		return nil, fmt.Errorf("invalid TTS response: missing audio data")
	}

	audioBytes, err := base64.StdEncoding.DecodeString(payload.Choices[0].Message.Audio.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio data: %w", err)
	}

	return io.NopCloser(bytes.NewReader(audioBytes)), nil
}
