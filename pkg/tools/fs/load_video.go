package fstools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/media"
)

// LoadVideoTool loads a local video file into the MediaStore and returns a
// media:// reference. The provider handles the actual video processing —
// multimodal models that support video can analyze it directly.
type LoadVideoTool struct {
	workspace   string
	restrict    bool
	maxFileSize int
	mediaStore  media.MediaStore
	allowPaths  []*regexp.Regexp

	defaultChannel string
	defaultChatID  string
}

func NewLoadVideoTool(
	workspace string,
	restrict bool,
	maxFileSize int,
	store media.MediaStore,
	allowPaths ...[]*regexp.Regexp,
) *LoadVideoTool {
	if maxFileSize <= 0 {
		maxFileSize = config.DefaultMaxMediaSize
	}
	var patterns []*regexp.Regexp
	if len(allowPaths) > 0 {
		patterns = allowPaths[0]
	}
	return &LoadVideoTool{
		workspace:   workspace,
		restrict:    restrict,
		maxFileSize: maxFileSize,
		mediaStore:  store,
		allowPaths:  patterns,
	}
}

func (t *LoadVideoTool) Name() string { return "load_video" }

func (t *LoadVideoTool) Description() string {
	return "Load a local video file so you can analyze its contents. " +
		"Supported formats: MP4, WebM, MOV, AVI, MKV and other common video formats. " +
		"After calling this tool, describe or analyze the video in your next response."
}

func (t *LoadVideoTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the local video file. Relative paths are resolved from workspace.",
			},
		},
		"required": []string{"path"},
	}
}

func (t *LoadVideoTool) SetContext(channel, chatID string) {
	t.defaultChannel = channel
	t.defaultChatID = chatID
}

func (t *LoadVideoTool) SetMediaStore(store media.MediaStore) {
	t.mediaStore = store
}

func (t *LoadVideoTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	path, _ := args["path"].(string)
	if strings.TrimSpace(path) == "" {
		return ErrorResult("path is required")
	}

	channel := ToolChannel(ctx)
	if channel == "" {
		channel = t.defaultChannel
	}
	chatID := ToolChatID(ctx)
	if chatID == "" {
		chatID = t.defaultChatID
	}
	if channel == "" || chatID == "" {
		return ErrorResult("no target channel/chat available")
	}

	if t.mediaStore == nil {
		return ErrorResult("media store not configured")
	}

	resolved, err := validatePathWithAllowPaths(path, t.workspace, t.restrict, t.allowPaths)
	if err != nil {
		return ErrorResult(fmt.Sprintf("invalid path: %v", err))
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return ErrorResult(fmt.Sprintf("file not found: %v", err))
	}
	if info.IsDir() {
		return ErrorResult("path is a directory, expected a video file")
	}
	if info.Size() > int64(t.maxFileSize) {
		return ErrorResult(fmt.Sprintf(
			"file too large: %d bytes (max %d bytes)", info.Size(), t.maxFileSize,
		))
	}

	mediaType := detectMediaType(resolved)
	if !strings.HasPrefix(mediaType, "video/") {
		return ErrorResult(fmt.Sprintf(
			"file does not appear to be a video (detected type: %s)", mediaType,
		))
	}

	filename := filepath.Base(resolved)
	scope := fmt.Sprintf("tool:load_video:%s:%s", channel, chatID)

	ref, err := t.mediaStore.Store(resolved, media.MediaMeta{
		Filename:      filename,
		ContentType:   mediaType,
		Source:        "tool:load_video",
		CleanupPolicy: media.CleanupPolicyForgetOnly,
	}, scope)
	if err != nil {
		return ErrorResult(fmt.Sprintf("failed to register video in media store: %v", err))
	}

	msg := fmt.Sprintf("Video loaded: %s\n[video: %s]", filename, resolved)

	return &ToolResult{
		ForLLM:  msg,
		ForUser: fmt.Sprintf("Loaded video: %s", filename),
		Media:   []string{ref},
	}
}
