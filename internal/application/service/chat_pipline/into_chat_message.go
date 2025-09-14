package chatpipline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

	"github.com/Tencent/WeKnowRust/internal/logger"
	"github.com/Tencent/WeKnowRust/internal/types"
)

// PluginIntoChatMessage handles the transformation of search results into chat messages
type PluginIntoChatMessage struct{}

// NewPluginIntoChatMessage creates and registers a new PluginIntoChatMessage instance
func NewPluginIntoChatMessage(eventManager *EventManager) *PluginIntoChatMessage {
	res := &PluginIntoChatMessage{}
	eventManager.Register(res)
	return res
}

// ActivationEvents returns the event types this plugin handles
func (p *PluginIntoChatMessage) ActivationEvents() []types.EventType {
	return []types.EventType{types.INTO_CHAT_MESSAGE}
}

// OnEvent processes the INTO_CHAT_MESSAGE event to format chat message content
func (p *PluginIntoChatMessage) OnEvent(ctx context.Context,
	eventType types.EventType, chatManage *types.ChatManage, next func() *PluginError,
) *PluginError {
	// Extract content from merge results
	passages := make([]string, len(chatManage.MergeResult))
	for i, result := range chatManage.MergeResult {
		// Merge content and image information
		passages[i] = getEnrichedPassageForChat(ctx, result)
	}

	// Parse the context template
	tmpl, err := template.New("searchContent").Parse(chatManage.SummaryConfig.ContextTemplate)
	if err != nil {
		return ErrTemplateParse.WithError(err)
	}

	// Prepare weekday names for template (English)
	weekdayName := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	var userContent bytes.Buffer

	// Execute template with context data
	err = tmpl.Execute(&userContent, map[string]interface{}{
		"Query":       chatManage.Query,                         // User's original query
		"Contexts":    passages,                                 // Extracted passages from search results
		"CurrentTime": time.Now().Format("2006-01-02 15:04:05"), // Formatted current time
		"CurrentWeek": weekdayName[time.Now().Weekday()],        // Current weekday name
	})
	if err != nil {
		return ErrTemplateExecute.WithError(err)
	}

	// Set formatted content back to chat management
	chatManage.UserContent = userContent.String()
	return next()
}

// getEnrichedPassageForChat merges Content and ImageInfo text for chat messages
func getEnrichedPassageForChat(ctx context.Context, result *types.SearchResult) string {
	// If there is no image info and no content, return empty
	if result.Content == "" && result.ImageInfo == "" {
		return ""
	}

	// If only content is present, return it
	if result.ImageInfo == "" {
		return result.Content
	}

	// Merge image information into content
	return enrichContentWithImageInfo(ctx, result.Content, result.ImageInfo)
}

// Regular expression to match Markdown image links
var markdownImageRegex = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

// enrichContentWithImageInfo merges image information with content
func enrichContentWithImageInfo(ctx context.Context, content string, imageInfoJSON string) string {
	// Parse ImageInfo
	var imageInfos []types.ImageInfo
	err := json.Unmarshal([]byte(imageInfoJSON), &imageInfos)
	if err != nil {
		logger.Warnf(ctx, "Failed to parse ImageInfo: %v, using content only", err)
		return content
	}

	if len(imageInfos) == 0 {
		return content
	}

	// Build map from image URL to info
	imageInfoMap := make(map[string]*types.ImageInfo)
	for i := range imageInfos {
		if imageInfos[i].URL != "" {
			imageInfoMap[imageInfos[i].URL] = &imageInfos[i]
		}
		// Also map OriginalURL
		if imageInfos[i].OriginalURL != "" {
			imageInfoMap[imageInfos[i].OriginalURL] = &imageInfos[i]
		}
	}

	// Find all Markdown image links in content
	matches := markdownImageRegex.FindAllStringSubmatch(content, -1)

	// Track processed image URLs
	processedURLs := make(map[string]bool)

	logger.Infof(ctx, "Found %d Markdown image links in content", len(matches))

	// Replace each image link by appending caption/OCR text
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		// Extract image URL (ignore alt text)
		imgURL := match[2]

		// Mark URL processed
		processedURLs[imgURL] = true

		// Lookup matching image info
		imgInfo, found := imageInfoMap[imgURL]

		// If found, append caption and OCR text
		if found && imgInfo != nil {
			replacement := match[0] + "\n"
			if imgInfo.Caption != "" {
				replacement += fmt.Sprintf("Image caption: %s\n", imgInfo.Caption)
			}
			if imgInfo.OCRText != "" {
				replacement += fmt.Sprintf("Image text: %s\n", imgInfo.OCRText)
			}
			content = strings.Replace(content, match[0], replacement, 1)
		}
	}

	// Append info for images not present in content but in ImageInfo
	var additionalImageTexts []string
	for _, imgInfo := range imageInfos {
		// Skip already processed URLs
		if processedURLs[imgInfo.URL] || processedURLs[imgInfo.OriginalURL] {
			continue
		}

		var imgTexts []string
		if imgInfo.Caption != "" {
			imgTexts = append(imgTexts, fmt.Sprintf("Image %s caption: %s", imgInfo.URL, imgInfo.Caption))
		}
		if imgInfo.OCRText != "" {
			imgTexts = append(imgTexts, fmt.Sprintf("Image %s text: %s", imgInfo.URL, imgInfo.OCRText))
		}

		if len(imgTexts) > 0 {
			additionalImageTexts = append(additionalImageTexts, imgTexts...)
		}
	}

	// Append additional image information at the end
	if len(additionalImageTexts) > 0 {
		if content != "" {
			content += "\n\n"
		}
		content += "Additional image information:\n" + strings.Join(additionalImageTexts, "\n")
	}

	logger.Debugf(ctx, "Enhanced content with image info: found %d Markdown images, added %d additional images",
		len(matches), len(additionalImageTexts))

	return content
}
