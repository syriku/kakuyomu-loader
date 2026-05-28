# kakuyomu-loader

`kakuyomu-loader` 是一用于抓取并解析「カクヨム（Kakuyomu）」小说内容的 Go 语言库。

## 功能特性

- **轻量级解析**：仅依赖 Go 标准库（`net/http`、`regexp` 等），无需引入无头浏览器或其他庞大的 HTML 解析依赖。
- **纯净文本提取**：能够从网页中准确提取小说的章节标题与正文段落，并自动去除所有无关的 HTML 标签，返回纯净的可阅读文本。
- **生词/注音提取**：自动解析日语小说特有的 `<ruby>`（振假名注音）标签内容。在剥离正文注音的同时，将其汇总至 `Vocabulary`（生词表）字典中（映射格式为 `单词 -> 假名读音`），方便后续处理或展示。

## 项目关联

此库是专门为 [transmas](https://github.com/syriku/transmas) 开发的配套模块。

## 快速使用

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fhyxz/kakuyomu-loader"
)

func main() {
	loader := loader.NewKakuyomuLoader()
	url := "https://kakuyomu.jp/works/16817330657430657810/episodes/16817330657430704350"

	novel, err := loader.Load(context.Background(), url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("标题: %s\n", novel.Title)
	fmt.Printf("正文长度: %d 字符\n", len(novel.Content))
	
	fmt.Println("生词表:")
	for word, pron := range novel.Vocabulary {
		fmt.Printf("- %s: %s\n", word, pron)
	}
}
```
