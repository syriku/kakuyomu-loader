package loader

import (
	"html"
	"regexp"
	"strings"
	"sync"
)

// HtmlReader defines the interface for parsing HTML content into a Novel.
type HtmlReader interface {
	KakuyomuNovel() (*Novel, error)
	AnyHtmlNovel() (*Novel, error)
}

// defaultHtmlReader is the default implementation of HtmlReader.
type defaultHtmlReader struct {
	html          string
	kakuyomuOnce  sync.Once
	kakuyomuNovel *Novel
	kakuyomuErr   error
	anyOnce       sync.Once
	anyNovel      *Novel
	anyErr        error
}

// NewHtmlReader creates a new instance of HtmlReader.
func NewHtmlReader(html string) HtmlReader {
	return &defaultHtmlReader{
		html: html,
	}
}

// KakuyomuNovel returns the parsed Novel from Kakuyomu format, performing parsing lazily on the first call.
func (r *defaultHtmlReader) KakuyomuNovel() (*Novel, error) {
	r.kakuyomuOnce.Do(func() {
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

		r.kakuyomuNovel = &Novel{
			Title:      title,
			Content:    content,
			Vocabulary: vocabulary,
		}
	})

	return r.kakuyomuNovel, r.kakuyomuErr
}

// AnyHtmlNovel returns the parsed Novel from any HTML format, converting it to plain text.
func (r *defaultHtmlReader) AnyHtmlNovel() (*Novel, error) {
	r.anyOnce.Do(func() {
		vocabulary := make(map[string]string)
		extractVocabulary(r.html, vocabulary)

		titleRe := regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
		titleMatch := titleRe.FindStringSubmatch(r.html)
		title := ""
		if len(titleMatch) > 1 {
			title = cleanParagraph(titleMatch[1])
		}

		content := cleanHtmlToText(r.html)

		r.anyNovel = &Novel{
			Title:      title,
			Content:    content,
			Vocabulary: vocabulary,
		}
	})

	return r.anyNovel, r.anyErr
}

func cleanHtmlToText(s string) string {
	rtRe := regexp.MustCompile(`(?is)<rt[^>]*>.*?</rt>`)
	rpRe := regexp.MustCompile(`(?is)<rp[^>]*>.*?</rp>`)
	s = rtRe.ReplaceAllString(s, "")
	s = rpRe.ReplaceAllString(s, "")

	brRe := regexp.MustCompile(`(?is)<br\s*/?>`)
	s = brRe.ReplaceAllString(s, "\n")

	blockEndRe := regexp.MustCompile(`(?is)</(div|p|h[1-6]|li)>`)
	s = blockEndRe.ReplaceAllString(s, "\n")

	tagRe := regexp.MustCompile(`(?is)<[^>]+>`)
	s = tagRe.ReplaceAllString(s, "")

	s = html.UnescapeString(s)

	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
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
