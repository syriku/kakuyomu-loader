package loader

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
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

// HtmlReader defines the interface for parsing HTML content into a Novel.
type HtmlReader interface {
	Novel() (*Novel, error)
}

// defaultHtmlReader is the default implementation of HtmlReader.
type defaultHtmlReader struct {
	html  string
	once  sync.Once
	novel *Novel
	err   error
}

// NewHtmlReader creates a new instance of HtmlReader.
func NewHtmlReader(html string) HtmlReader {
	return &defaultHtmlReader{
		html: html,
	}
}

// Novel returns the parsed Novel, performing parsing lazily on the first call.
func (r *defaultHtmlReader) Novel() (*Novel, error) {
	r.once.Do(func() {
		// Parse title
		titleRe := regexp.MustCompile(`(?is)<p class="widget-episodeTitle[^>]*>(.*?)</p>`)
		titleMatch := titleRe.FindStringSubmatch(r.html)
		title := ""
		if len(titleMatch) > 1 {
			title = cleanParagraph(titleMatch[1])
		}

		// Parse content paragraphs
		pRe := regexp.MustCompile(`(?is)<p\s+id="p\d+"[^>]*>(.*?)</p>`)
		pMatches := pRe.FindAllStringSubmatch(r.html, -1)

		var paragraphs []string
		vocabulary := make(map[string]string)
		for _, m := range pMatches {
			extractVocabulary(m[1], vocabulary)
			cleaned := cleanParagraph(m[1])
			paragraphs = append(paragraphs, cleaned)
		}

		content := strings.Join(paragraphs, "\n")

		r.novel = &Novel{
			Title:      title,
			Content:    content,
			Vocabulary: vocabulary,
		}
	})

	return r.novel, r.err
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
	return reader.Novel()
}

func cleanParagraph(s string) string {
	rtRe := regexp.MustCompile(`(?is)<rt[^>]*>.*?</rt>`)
	rpRe := regexp.MustCompile(`(?is)<rp[^>]*>.*?</rp>`)
	brRe := regexp.MustCompile(`(?is)<br\s*/?>`)
	tagRe := regexp.MustCompile(`(?is)<[^>]+>`)

	s = rtRe.ReplaceAllString(s, "")
	s = rpRe.ReplaceAllString(s, "")
	s = brRe.ReplaceAllString(s, "\n")
	s = tagRe.ReplaceAllString(s, "")

	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}

func extractVocabulary(s string, vocab map[string]string) {
	rubyRe := regexp.MustCompile(`(?is)<ruby[^>]*>(.*?)</ruby>`)
	rtRe := regexp.MustCompile(`(?is)<rt[^>]*>(.*?)</rt>`)
	rpRe := regexp.MustCompile(`(?is)<rp[^>]*>.*?</rp>`)
	tagRe := regexp.MustCompile(`(?is)<[^>]+>`)

	matches := rubyRe.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		rubyContent := m[1]

		var prons []string
		rtMatches := rtRe.FindAllStringSubmatch(rubyContent, -1)
		for _, rtm := range rtMatches {
			prons = append(prons, strings.TrimSpace(html.UnescapeString(tagRe.ReplaceAllString(rtm[1], ""))))
		}
		pronunciation := strings.Join(prons, "")

		word := rtRe.ReplaceAllString(rubyContent, "")
		word = rpRe.ReplaceAllString(word, "")
		word = tagRe.ReplaceAllString(word, "")
		word = strings.TrimSpace(html.UnescapeString(word))

		if word != "" && pronunciation != "" {
			vocab[word] = pronunciation
		}
	}
}
