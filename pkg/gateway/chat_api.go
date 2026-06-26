package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/identity"
	"github.com/sipeed/picoclaw/pkg/logger"
)

// ChatAPIHandler serves REST API endpoints for chat (/api/chat, /api/chat/stream).
type ChatAPIHandler struct {
	bus *bus.MessageBus
}

// NewChatAPIHandler creates a new ChatAPIHandler.
func NewChatAPIHandler(msgBus *bus.MessageBus) *ChatAPIHandler {
	return &ChatAPIHandler{bus: msgBus}
}

type chatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Model     string `json:"model,omitempty"`
}

type chatResponse struct {
	SessionID string         `json:"session_id"`
	Model     string         `json:"model,omitempty"`
	Reply     string         `json:"reply"`
	ToolCalls []toolCallInfo `json:"tool_calls,omitempty"`
	Timestamp string         `json:"timestamp"`
}

type toolCallInfo struct {
	Name   string `json:"name"`
	Args   string `json:"args,omitempty"`
	Result string `json:"result,omitempty"`
}

// pendingChat tracks an in-flight chat request waiting for agent responses.
type pendingChat struct {
	done     chan struct{}
	reply    strings.Builder
	toolCall []toolCallInfo
	model    string
	once     sync.Once
}

func (p *pendingChat) finalize() {
	p.once.Do(func() { close(p.done) })
}

// handleChat handles POST /api/chat — synchronous, waits for full reply.
func (h *ChatAPIHandler) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		writeJSONError(w, http.StatusBadRequest, "message is required")
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	chatID := "api:" + sessionID
	senderID := "api-user"
	msgID := uuid.New().String()

	sender := bus.SenderInfo{
		Platform:    "api",
		PlatformID:  senderID,
		CanonicalID: identity.BuildCanonicalID("api", senderID),
	}

	metadata := map[string]string{
		"platform":   "api",
		"session_id": sessionID,
	}
	if req.Model != "" {
		metadata["model_override"] = req.Model
	}

	inboundCtx := bus.InboundContext{
		Channel:   "api",
		ChatID:    chatID,
		ChatType:  "direct",
		SenderID:  senderID,
		MessageID: msgID,
		Raw:       metadata,
	}

	msg := bus.InboundMessage{
		Context:    inboundCtx,
		Sender:     sender,
		Content:    req.Message,
		MediaScope: "api:" + chatID + ":" + msgID,
	}
	msg = bus.NormalizeInboundMessage(msg)

	pending := &pendingChat{
		done: make(chan struct{}),
	}

	// Register response collector before publishing.
	registerPendingChat(chatID, pending)
	defer unregisterPendingChat(chatID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	if err := h.bus.PublishInbound(ctx, msg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to publish message: "+err.Error())
		return
	}

	logger.DebugCF("api", "Chat message published", map[string]any{
		"session_id": sessionID,
		"preview":    truncateRunes(req.Message, 50),
	})

	// Wait for agent response.
	select {
	case <-pending.done:
		// ok
	case <-ctx.Done():
		writeJSONError(w, http.StatusGatewayTimeout, "request timed out waiting for agent response")
		return
	}

	resp := chatResponse{
		SessionID: sessionID,
		Model:     pending.model,
		Reply:     strings.TrimSpace(pending.reply.String()),
		ToolCalls: pending.toolCall,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleChatStream handles POST /api/chat/stream — SSE streaming response.
func (h *ChatAPIHandler) handleChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		writeJSONError(w, http.StatusBadRequest, "message is required")
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	chatID := "api:" + sessionID
	senderID := "api-user"
	msgID := uuid.New().String()

	sender := bus.SenderInfo{
		Platform:    "api",
		PlatformID:  senderID,
		CanonicalID: identity.BuildCanonicalID("api", senderID),
	}

	metadata := map[string]string{
		"platform":   "api",
		"session_id": sessionID,
	}
	if req.Model != "" {
		metadata["model_override"] = req.Model
	}

	inboundCtx := bus.InboundContext{
		Channel:   "api",
		ChatID:    chatID,
		ChatType:  "direct",
		SenderID:  senderID,
		MessageID: msgID,
		Raw:       metadata,
	}

	msg := bus.InboundMessage{
		Context:    inboundCtx,
		Sender:     sender,
		Content:    req.Message,
		MediaScope: "api:" + chatID + ":" + msgID,
	}
	msg = bus.NormalizeInboundMessage(msg)

	pending := &pendingChat{
		done: make(chan struct{}),
	}

	registerPendingChat(chatID, pending)
	defer unregisterPendingChat(chatID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	if err := h.bus.PublishInbound(ctx, msg); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to publish message: "+err.Error())
		return
	}

	// Set SSE headers.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, canFlush := w.(http.Flusher)

	// Send start event.
	fmt.Fprintf(w, "event: start\ndata: %s\n\n", mustJSON(map[string]string{
		"session_id": sessionID,
	}))
	if canFlush {
		flusher.Flush()
	}

	// Stream outbound messages until done or cancelled.
	outChan := h.bus.OutboundChan()
	for {
		select {
		case outMsg, ok := <-outChan:
			if !ok {
				return
			}
			if outMsg.ChatID != chatID && outMsg.SessionKey != chatID {
				// Not for this request — re-publish is not possible,
				// so we rely on the bus subscriber pattern.
				// In practice, the agent loop sends directly to the channel,
				// and the Pico channel handles delivery. For the REST API,
				// we need a different approach — see response collector below.
				continue
			}

			content := strings.TrimSpace(outMsg.Content)
			if content != "" {
				fmt.Fprintf(w, "event: text\ndata: %s\n\n", mustJSON(content))
				if canFlush {
					flusher.Flush()
				}
			}

		case <-pending.done:
			// Agent finished processing.
			fmt.Fprintf(w, "event: done\ndata: %s\n\n", mustJSON(map[string]any{
				"session_id": sessionID,
				"reply":      strings.TrimSpace(pending.reply.String()),
			}))
			if canFlush {
				flusher.Flush()
			}
			return

		case <-ctx.Done():
			fmt.Fprintf(w, "event: done\ndata: %s\n\n", mustJSON(map[string]string{
				"session_id": sessionID,
				"error":      "timeout",
			}))
			if canFlush {
				flusher.Flush()
			}
			return
		}
	}
}

// --- Response collector ---
// The agent loop sends OutboundMessages to channels. For the REST API,
// we need to intercept these. We use a simple approach: register a
// per-chatID collector that the gateway's outbound relay calls.

var (
	pendingChatsMu sync.RWMutex
	pendingChats   = make(map[string]*pendingChat)
)

func registerPendingChat(chatID string, p *pendingChat) {
	pendingChatsMu.Lock()
	pendingChats[chatID] = p
	pendingChatsMu.Unlock()
}

func unregisterPendingChat(chatID string) {
	pendingChatsMu.Lock()
	delete(pendingChats, chatID)
	pendingChatsMu.Unlock()
}

// CollectOutboundResponse is called by the outbound relay when a message
// is destined for an API chat. Returns true if the message was consumed.
func CollectOutboundResponse(msg bus.OutboundMessage) bool {
	pendingChatsMu.RLock()
	p, ok := pendingChats[msg.ChatID]
	pendingChatsMu.RUnlock()
	if !ok {
		return false
	}

	content := strings.TrimSpace(msg.Content)
	if content != "" {
		p.reply.WriteString(content)
	}

	if msg.AgentID != "" {
		p.model = msg.AgentID
	}

	// Check if this is a finalizable message (not thought/tool_calls/tool_feedback).
	kind := ""
	if msg.Context.Raw != nil {
		kind = strings.TrimSpace(msg.Context.Raw["message_kind"])
	}
	isIntermediate := kind == "thought" || kind == "tool_calls" || kind == "tool_feedback"

	if !isIntermediate {
		p.finalize()
	}

	return true
}

// CollectOutboundMediaResponse handles media messages for API chats.
func CollectOutboundMediaResponse(msg bus.OutboundMediaMessage) bool {
	pendingChatsMu.RLock()
	_, ok := pendingChats[msg.ChatID]
	pendingChatsMu.RUnlock()
	return ok
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
