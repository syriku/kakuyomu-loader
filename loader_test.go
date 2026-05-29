package loader

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func TestKakuyomuLoader_Load(t *testing.T) {
	loader := NewKakuyomuLoader()
	url := "https://kakuyomu.jp/works/16817330657430657810/episodes/16817330657430704350"

	novel, err := loader.Load(context.Background(), url)
	if err != nil {
		t.Fatalf("failed to load novel: %v", err)
	}

	if novel.Title == "" {
		t.Error("expected non-empty title")
	}

	if novel.Content == "" {
		t.Error("expected non-empty content")
	}

	if !strings.Contains(novel.Content, "篠崎翼") {
		t.Errorf("expected content to contain '篠崎翼' without ruby tags, got: %s...", novel.Content[:100])
	}

	if pron, ok := novel.Vocabulary["篠崎翼"]; !ok {
		t.Error("expected vocabulary to contain '篠崎翼'")
	} else if pron != "しのざき つばさ" {
		t.Errorf("expected pronunciation 'しのざき つばさ', got '%s'", pron)
	}
}

func TestHtmlReader(t *testing.T) {
	sampleHTML := `
<p class="widget-episodeTitle">Test Title</p>
<p id="p1">Hello <ruby>篠崎翼<rt>しのざき つばさ</rt></ruby>!</p>
<p id="p2">Paragraph 2</p>
`
	reader := NewHtmlReader(sampleHTML)

	// Call Novel() first time
	novel, err := reader.Novel()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if novel.Title != "Test Title" {
		t.Errorf("expected 'Test Title', got '%s'", novel.Title)
	}

	expectedContent := "Hello 篠崎翼!\nParagraph 2"
	if novel.Content != expectedContent {
		t.Errorf("expected content %q, got %q", expectedContent, novel.Content)
	}

	if pron, ok := novel.Vocabulary["篠崎翼"]; !ok {
		t.Error("expected vocabulary to contain '篠崎翼'")
	} else if pron != "しのざき つばさ" {
		t.Errorf("expected pronunciation 'しのざき つばさ', got '%s'", pron)
	}

	// Call Novel() second time to test lazy loading (returning same instance)
	novel2, err := reader.Novel()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if novel != novel2 {
		t.Error("expected the same Novel pointer to be returned (lazy loading)")
	}
}

func TestHtmlReader_Concurrency(t *testing.T) {
	sampleHTML := `
<p class="widget-episodeTitle">Concurrent Test</p>
<p id="p1">Content</p>
`
	reader := NewHtmlReader(sampleHTML)

	var wg sync.WaitGroup
	const numGoroutines = 10
	novels := make([]*Novel, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			n, err := reader.Novel()
			novels[idx] = n
			errors[idx] = err
		}(i)
	}
	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		if errors[i] != nil {
			t.Errorf("goroutine %d got error: %v", i, errors[i])
		}
		if novels[i] != novels[0] {
			t.Errorf("goroutine %d returned different novel pointer: %p vs %p", i, novels[i], novels[0])
		}
	}
}
