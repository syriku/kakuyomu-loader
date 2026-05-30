package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Novel represents a parsed novel chapter.
type Novel struct {
	Title      string
	Content    string
	Vocabulary map[string]string
}

// KakuyomuLoader defines the interface for loading novel content.
type KakuyomuLoader interface {
	Load(ctx context.Context, url string) (*Novel, error)
}

// DefaultKakuyomuLoader is the default implementation of KakuyomuLoader.
type DefaultKakuyomuLoader struct {
	client *http.Client
}

// NewKakuyomuLoader creates a new instance of KakuyomuLoader.
func NewKakuyomuLoader() KakuyomuLoader {
	return &DefaultKakuyomuLoader{
		client: &http.Client{},
	}
}

// Load fetches and parses the novel content from the given URL.
func (l *DefaultKakuyomuLoader) Load(ctx context.Context, url string) (*Novel, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	reader := NewHtmlReader(string(bodyBytes))
	return reader.KakuyomuNovel()
}
