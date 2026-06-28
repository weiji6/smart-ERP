package erp

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	promptDocsOnce sync.Once
	promptDocsText string
	promptDocsErr  error
)

func PromptRuleDocuments() string {
	promptDocsOnce.Do(func() {
		files := []struct {
			Title string
			Path  string
		}{
			{Title: "ERP 沙盘规则摘要", Path: "docs/rules.md"},
			{Title: "年度运营流程表", Path: "docs/operation_flow.md"},
		}

		var b strings.Builder
		for _, file := range files {
			data, err := readPromptDoc(file.Path)
			if err != nil {
				promptDocsErr = err
				continue
			}
			b.WriteString(fmt.Sprintf("\n## %s\n", file.Title))
			b.WriteString(strings.TrimSpace(string(data)))
			b.WriteString("\n")
		}
		promptDocsText = strings.TrimSpace(b.String())
	})
	if promptDocsErr != nil && promptDocsText == "" {
		return "规则文档读取失败，请以系统内置规则和运营流程为准。"
	}
	return promptDocsText
}

func readPromptDoc(path string) ([]byte, error) {
	candidates := []string{path, "../" + path, "../../" + path, "../../../" + path}
	var lastErr error
	for _, candidate := range candidates {
		data, err := os.ReadFile(candidate)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
