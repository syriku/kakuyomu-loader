package loader

import (
	"context"
	"strings"
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
