package erp

import (
	"fmt"
	"strings"
)

type AdvisorQARecord struct {
	ID        string `json:"id"`
	Company   string `json:"company"`
	Year      int    `json:"year"`
	Quarter   int    `json:"quarter"`
	StepCode  string `json:"stepCode"`
	StepName  string `json:"stepName"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Mode      string `json:"mode"`
	AIUsed    bool   `json:"aiUsed"`
	CreatedAt string `json:"createdAt"`
}

func FormatAdvisorQAHistory(records []AdvisorQARecord) string {
	if len(records) == 0 {
		return "暂无历史建议记录。\n"
	}

	var b strings.Builder
	b.WriteString("最近历史建议记录（Q/A）：\n")
	for _, record := range records {
		question := strings.TrimSpace(record.Question)
		if question == "" {
			question = "未填写具体问题"
		}
		answer := firstNonEmptyLine(record.Answer)
		if answer == "" {
			answer = "暂无回答摘要"
		}
		b.WriteString(fmt.Sprintf("- Y%dQ%d %s：Q：%s；A：%s\n",
			record.Year, record.Quarter, record.StepName, question, answer))
	}
	return b.String()
}

func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}
