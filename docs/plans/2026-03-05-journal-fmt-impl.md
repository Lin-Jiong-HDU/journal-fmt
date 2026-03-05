# journal-fmt 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个类似 gofmt 的 hledger 记账文件格式化工具

**Architecture:** 解析器将 .journal 文件解析为 AST，格式化器按规则重新生成格式化文本。CLI 支持 stdout 输出和 -w 覆盖原文件。

**Tech Stack:** Go 1.25, 标准库 flag/filepath 包

---

## Task 1: 创建项目结构和 types 模块

**Files:**
- Create: `internal/types/types.go`

**Step 1: 创建 types 文件**

```go
// internal/types/types.go

package types

// Journal 表示整个账本文件
type Journal struct {
	Items []Item
}

// Item 表示文件中的一个条目
type Item interface {
	isItem()
}

// Comment 注释行
type Comment struct {
	Text  string // 注释内容（不含分号）
	IsTag bool   // 是否是 tag（;  tag: 格式）
}

func (c *Comment) isItem() {}

// PriceDecl 价格声明 (P 2026/03/01 CNY 1.00 USD 7.20)
type PriceDecl struct {
	Date             string
	Commodity        string
	Price            string
	TargetCommodity  string
}

func (p *PriceDecl) isItem() {}

// Transaction 交易
type Transaction struct {
	Date        string     // 日期
	Status      string     // "*", "!", 或空
	Description string     // 描述
	Postings    []Posting  // 分录
	Comment     string     // 行尾注释（可选）
}

func (t *Transaction) isItem() {}

// Posting 分录（账户 + 金额）
type Posting struct {
	Account   string // 账户名
	Amount    string // 金额数值
	Commodity string // 货币单位
	Comment   string // 行尾注释（可选）
}

// EmptyLine 空行
type EmptyLine struct{}

func (e *EmptyLine) isItem() {}
```

**Step 2: 验证编译**

Run: `go build ./...`
Expected: 无错误

**Step 3: Commit**

```bash
git add internal/types/types.go
git commit -m "feat: add types module with AST definitions"
```

---

## Task 2: 实现 Comment 解析（TDD）

**Files:**
- Create: `internal/parser/parser.go`
- Create: `internal/parser/parser_test.go`

**Step 1: 写测试 - 解析普通注释**

```go
// internal/parser/parser_test.go

package parser

import (
	"testing"
)

func TestParseComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantText string
		wantTag  bool
	}{
		{
			name:     "normal comment",
			input:    "; 2026年3月交易",
			wantText: "2026年3月交易",
			wantTag:  false,
		},
		{
			name:     "separator line",
			input:    "; =======================================",
			wantText: "=======================================",
			wantTag:  false,
		},
		{
			name:     "tag comment",
			input:    ";  夜宵:",
			wantText: "夜宵:",
			wantTag:  true,
		},
		{
			name:     "tag with extra spaces",
			input:    ";  F1:",
			wantText: "F1:",
			wantTag:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			journal, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if len(journal.Items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(journal.Items))
			}
			comment := journal.Items[0].(*Comment)
			if comment.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", comment.Text, tt.wantText)
			}
			if comment.IsTag != tt.wantTag {
				t.Errorf("IsTag = %v, want %v", comment.IsTag, tt.wantTag)
			}
		})
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/parser/... -v`
Expected: FAIL (parser not implemented)

**Step 3: 实现最小代码使测试通过**

```go
// internal/parser/parser.go

package parser

import (
	"regexp"
	"strings"

	"github.com/Lin-Jiong-HDU/journal-fmt/internal/types"
)

type Parser struct {
	lines []string
	pos   int
}

func NewParser(content string) *Parser {
	return &Parser{
		lines: strings.Split(content, "\n"),
		pos:   0,
	}
}

func (p *Parser) Parse() (*types.Journal, error) {
	journal := &types.Journal{}

	for p.pos < len(p.lines) {
		line := p.lines[p.pos]

		if strings.TrimSpace(line) == "" {
			journal.Items = append(journal.Items, &types.EmptyLine{})
			p.pos++
			continue
		}

		if strings.HasPrefix(strings.TrimSpace(line), ";") {
			comment := p.parseComment(line)
			journal.Items = append(journal.Items, comment)
			p.pos++
			continue
		}

		// TODO: handle other item types
		p.pos++
	}

	return journal, nil
}

func (p *Parser) parseComment(line string) *types.Comment {
	// Remove leading semicolon and spaces
	content := strings.TrimLeft(line, " \t")
	content = strings.TrimPrefix(content, ";")

	// Check if it's a tag (two spaces + text + colon)
	// Tag pattern: "  word:" where word ends with colon
	tagPattern := regexp.MustCompile(`^ {2}(\S+:)\s*$`)
	if matches := tagPattern.FindStringSubmatch(content); matches != nil {
		return &types.Comment{
			Text:  matches[1],
			IsTag: true,
		}
	}

	// Normal comment: trim leading space
	return &types.Comment{
		Text:  strings.TrimLeft(content, " "),
		IsTag: false,
	}
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/parser/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/parser/parser.go internal/parser/parser_test.go
git commit -m "feat(parser): implement comment parsing with tag detection"
```

---

## Task 3: 实现 PriceDecl 解析（TDD）

**Files:**
- Modify: `internal/parser/parser_test.go`
- Modify: `internal/parser/parser.go`

**Step 1: 写测试**

在 `parser_test.go` 添加：

```go
func TestParsePriceDecl(t *testing.T) {
	input := "P 2026/03/01 CNY 1.00 USD 7.20"
	p := NewParser(input)
	journal, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(journal.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(journal.Items))
	}
	priceDecl := journal.Items[0].(*types.PriceDecl)
	if priceDecl.Date != "2026-03-01" {
		t.Errorf("Date = %q, want %q", priceDecl.Date, "2026-03-01")
	}
	if priceDecl.Commodity != "CNY" {
		t.Errorf("Commodity = %q, want %q", priceDecl.Commodity, "CNY")
	}
	if priceDecl.Price != "1.00" {
		t.Errorf("Price = %q, want %q", priceDecl.Price, "1.00")
	}
	if priceDecl.TargetCommodity != "USD 7.20" {
		t.Errorf("TargetCommodity = %q, want %q", priceDecl.TargetCommodity, "USD 7.20")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/parser/... -v -run TestParsePriceDecl`
Expected: FAIL

**Step 3: 实现**

在 `parser.go` 的 `Parse()` 方法中，在处理注释的逻辑之后添加：

```go
// Check for price declaration
if strings.HasPrefix(strings.TrimSpace(line), "P ") {
	priceDecl, err := p.parsePriceDecl(line)
	if err != nil {
		return nil, err
	}
	journal.Items = append(journal.Items, priceDecl)
	p.pos++
	continue
}
```

添加 `parsePriceDecl` 方法：

```go
func (p *Parser) parsePriceDecl(line string) (*types.PriceDecl, error) {
	// Format: P DATE COMMODITY PRICE TARGET_COMMODITY TARGET_PRICE
	// Example: P 2026/03/01 CNY 1.00 USD 7.20
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "P "))
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid price declaration: %s", line)
	}

	date := p.normalizeDate(parts[0])

	return &types.PriceDecl{
		Date:            date,
		Commodity:       parts[1],
		Price:           parts[2],
		TargetCommodity: strings.Join(parts[3:], " "),
	}, nil
}

func (p *Parser) normalizeDate(date string) string {
	// Replace / and . with -
	return strings.ReplaceAll(strings.ReplaceAll(date, "/", "-"), ".", "-")
}
```

需要在文件顶部添加 `"fmt"` import。

**Step 4: 运行测试确认通过**

Run: `go test ./internal/parser/... -v -run TestParsePriceDecl`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/parser/parser.go internal/parser/parser_test.go
git commit -m "feat(parser): implement price declaration parsing with date normalization"
```

---

## Task 4: 实现 Transaction 解析（TDD）

**Files:**
- Modify: `internal/parser/parser_test.go`
- Modify: `internal/parser/parser.go`

**Step 1: 写测试 - 简单交易**

```go
func TestParseTransaction(t *testing.T) {
	input := `2026/03/02 * Apple iCloud+ 订阅
    expenses:subscription:icloud      21 CNY
    assets:wechat`
	p := NewParser(input)
	journal, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(journal.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(journal.Items))
	}
	tx := journal.Items[0].(*types.Transaction)
	if tx.Date != "2026-03-02" {
		t.Errorf("Date = %q, want %q", tx.Date, "2026-03-02")
	}
	if tx.Status != "*" {
		t.Errorf("Status = %q, want %q", tx.Status, "*")
	}
	if tx.Description != "Apple iCloud+ 订阅" {
		t.Errorf("Description = %q, want %q", tx.Description, "Apple iCloud+ 订阅")
	}
	if len(tx.Postings) != 2 {
		t.Fatalf("expected 2 postings, got %d", len(tx.Postings))
	}
	if tx.Postings[0].Account != "expenses:subscription:icloud" {
		t.Errorf("Account = %q", tx.Postings[0].Account)
	}
	if tx.Postings[0].Amount != "21" {
		t.Errorf("Amount = %q", tx.Postings[0].Amount)
	}
	if tx.Postings[0].Commodity != "CNY" {
		t.Errorf("Commodity = %q", tx.Postings[0].Commodity)
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/parser/... -v -run TestParseTransaction`
Expected: FAIL

**Step 3: 实现**

添加日期检测正则和交易解析逻辑到 `Parse()` 方法：

```go
// Date pattern: YYYY-MM-DD, YYYY/MM/DD, YYYY.MM.DD
datePattern := regexp.MustCompile(`^\d{4}[-/.]\d{2}[-/.]\d{2}`)
trimmedLine := strings.TrimSpace(line)

if datePattern.MatchString(trimmedLine) {
	tx, err := p.parseTransaction(line)
	if err != nil {
		return nil, err
	}
	journal.Items = append(journal.Items, tx)
	continue
}
```

添加 `parseTransaction` 和 `parsePosting` 方法：

```go
func (p *Parser) parseTransaction(line string) (*types.Transaction, error) {
	// Format: DATE [STATUS] DESCRIPTION [; COMMENT]
	// Example: 2026/03/02 * Apple iCloud+ 订阅
	// Example: 2026/03/02 * 夜宵（炒粉干） ;  夜宵:
	trimmedLine := strings.TrimSpace(line)

	// Extract date
	datePattern := regexp.MustCompile(`^(\d{4}[-/.]\d{2}[-/.]\d{2})`)
	dateMatch := datePattern.FindStringSubmatch(trimmedLine)
	if dateMatch == nil {
		return nil, fmt.Errorf("invalid transaction date: %s", line)
	}
	date := p.normalizeDate(dateMatch[1])

	rest := strings.TrimPrefix(trimmedLine, dateMatch[0])
	rest = strings.TrimSpace(rest)

	// Extract status (* or !)
	status := ""
	if strings.HasPrefix(rest, "*") || strings.HasPrefix(rest, "!") {
		status = string(rest[0])
		rest = strings.TrimSpace(rest[1:])
	}

	// Extract description and optional comment
	description := rest
	comment := ""
	if idx := strings.Index(rest, ";"); idx != -1 {
		description = strings.TrimSpace(rest[:idx])
		comment = strings.TrimSpace(rest[idx+1:])
	}

	tx := &types.Transaction{
		Date:        date,
		Status:      status,
		Description: description,
		Postings:    []types.Posting{},
		Comment:     comment,
	}

	// Parse postings (indented lines following the transaction)
	p.pos++
	for p.pos < len(p.lines) {
		postingLine := p.lines[p.pos]
		if !strings.HasPrefix(postingLine, " ") && !strings.HasPrefix(postingLine, "\t") {
			// Not a posting, stop
			p.pos-- // Will be incremented by the main loop
			break
		}
		posting := p.parsePosting(postingLine)
		tx.Postings = append(tx.Postings, posting)
		p.pos++
	}
	p.pos++ // Move past the last line we processed

	return tx, nil
}

func (p *Parser) parsePosting(line string) types.Posting {
	// Format: ACCOUNT [AMOUNT COMMODITY] [; COMMENT]
	// Example: expenses:subscription:icloud      21 CNY
	// Example: assets:wechat
	trimmedLine := strings.TrimSpace(line)

	// Extract comment if present
	comment := ""
	if idx := strings.Index(trimmedLine, ";"); idx != -1 {
		comment = strings.TrimSpace(trimmedLine[idx+1:])
		trimmedLine = strings.TrimSpace(trimmedLine[:idx])
	}

	// Parse account and amount
	// Account is the first word(s), amount and commodity follow
	// Pattern: account:parts [amount commodity]
	fields := strings.Fields(trimmedLine)
	if len(fields) == 0 {
		return types.Posting{}
	}

	// Find where account ends and amount begins
	// Amount is typically a number (possibly with decimal)
	accountEnd := len(fields)
	for i, field := range fields {
		// Check if this field looks like a number
		if _, err := strconv.ParseFloat(field, 64); err == nil {
			accountEnd = i
			break
		}
	}

	account := strings.Join(fields[:accountEnd], " ")

	posting := types.Posting{
		Account: account,
		Comment: comment,
	}

	// Extract amount and commodity
	if accountEnd < len(fields) {
		posting.Amount = fields[accountEnd]
		if accountEnd+1 < len(fields) {
			posting.Commodity = strings.Join(fields[accountEnd+1:], " ")
		}
	}

	return posting
}
```

需要添加 `"strconv"` import。

**Step 4: 运行测试确认通过**

Run: `go test ./internal/parser/... -v -run TestParseTransaction`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/parser/parser.go internal/parser/parser_test.go
git commit -m "feat(parser): implement transaction and posting parsing"
```

---

## Task 5: 写测试解析完整示例文件

**Files:**
- Modify: `internal/parser/parser_test.go`

**Step 1: 写集成测试**

```go
func TestParseFullFile(t *testing.T) {
	content := `; 2026年3月交易
; Monthly journal file

P 2026/03/01 CNY 1.00 USD 7.20

; =======================================
; Apple 订阅服务
; =======================================
2026/03/02 * Apple iCloud+ 订阅
    expenses:subscription:icloud      21 CNY
    assets:wechat

2026/03/02 * Apple Music 订阅
    expenses:subscription:music      6 CNY
    assets:wechat

; =======================================
; 日常开销
; =======================================
2026/03/02 * 夜宵（炒粉干） ;  夜宵:
    expenses:food:meal               21 CNY
    assets:wechat
`
	p := NewParser(content)
	journal, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Count item types
	comments := 0
	priceDecls := 0
	transactions := 0
	emptyLines := 0
	for _, item := range journal.Items {
		switch item.(type) {
		case *types.Comment:
			comments++
		case *types.PriceDecl:
			priceDecls++
		case *types.Transaction:
			transactions++
		case *types.EmptyLine:
			emptyLines++
		}
	}

	if comments != 6 {
		t.Errorf("expected 6 comments, got %d", comments)
	}
	if priceDecls != 1 {
		t.Errorf("expected 1 price declaration, got %d", priceDecls)
	}
	if transactions != 3 {
		t.Errorf("expected 3 transactions, got %d", transactions)
	}
	if emptyLines < 2 {
		t.Errorf("expected at least 2 empty lines, got %d", emptyLines)
	}
}
```

**Step 2: 运行测试**

Run: `go test ./internal/parser/... -v -run TestParseFullFile`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/parser/parser_test.go
git commit -m "test(parser): add integration test for full file parsing"
```

---

## Task 6: 实现 Comment 格式化（TDD）

**Files:**
- Create: `internal/formatter/formatter.go`
- Create: `internal/formatter/formatter_test.go`

**Step 1: 写测试**

```go
// internal/formatter/formatter_test.go

package formatter

import (
	"testing"

	"github.com/Lin-Jiong-HDU/journal-fmt/internal/types"
)

func TestFormatComment(t *testing.T) {
	f := NewFormatter()

	tests := []struct {
		name     string
		comment  *types.Comment
		want     string
	}{
		{
			name:    "normal comment",
			comment: &types.Comment{Text: "2026年3月交易", IsTag: false},
			want:    "; 2026年3月交易",
		},
		{
			name:    "separator line",
			comment: &types.Comment{Text: "=======================================", IsTag: false},
			want:    "; =======================================",
		},
		{
			name:    "tag comment",
			comment: &types.Comment{Text: "夜宵:", IsTag: true},
			want:    ";  夜宵:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.formatComment(tt.comment)
			if got != tt.want {
				t.Errorf("formatComment() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/formatter/... -v`
Expected: FAIL

**Step 3: 实现**

```go
// internal/formatter/formatter.go

package formatter

import (
	"github.com/Lin-Jiong-HDU/journal-fmt/internal/types"
)

type Formatter struct {
	AccountWidth   int
	AmountWidth    int
	CommodityWidth int
}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) Format(journal *types.Journal) string {
	// TODO: implement full formatting
	return ""
}

func (f *Formatter) formatComment(c *types.Comment) string {
	if c.IsTag {
		return ";  " + c.Text
	}
	return "; " + c.Text
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/formatter/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/formatter/formatter.go internal/formatter/formatter_test.go
git commit -m "feat(formatter): implement comment formatting"
```

---

## Task 7: 实现 PriceDecl 格式化（TDD）

**Files:**
- Modify: `internal/formatter/formatter_test.go`
- Modify: `internal/formatter/formatter.go`

**Step 1: 写测试**

```go
func TestFormatPriceDecl(t *testing.T) {
	f := NewFormatter()

	priceDecl := &types.PriceDecl{
		Date:             "2026-03-01",
		Commodity:        "CNY",
		Price:            "1.00",
		TargetCommodity:  "USD 7.20",
	}

	want := "P 2026-03-01 CNY 1.00 USD 7.20"
	got := f.formatPriceDecl(priceDecl)
	if got != want {
		t.Errorf("formatPriceDecl() = %q, want %q", got, want)
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/formatter/... -v -run TestFormatPriceDecl`
Expected: FAIL

**Step 3: 实现**

```go
func (f *Formatter) formatPriceDecl(p *types.PriceDecl) string {
	return fmt.Sprintf("P %s %s %s %s", p.Date, p.Commodity, p.Price, p.TargetCommodity)
}
```

需要添加 `"fmt"` import。

**Step 4: 运行测试确认通过**

Run: `go test ./internal/formatter/... -v -run TestFormatPriceDecl`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/formatter/formatter.go internal/formatter/formatter_test.go
git commit -m "feat(formatter): implement price declaration formatting"
```

---

## Task 8: 实现 Transaction 格式化（TDD）

**Files:**
- Modify: `internal/formatter/formatter_test.go`
- Modify: `internal/formatter/formatter.go`

**Step 1: 写测试 - 列宽计算**

```go
func TestCalculateWidths(t *testing.T) {
	journal := &types.Journal{
		Items: []types.Item{
			&types.Transaction{
				Postings: []types.Posting{
					{Account: "expenses:subscription:icloud", Amount: "21", Commodity: "CNY"},
					{Account: "assets:wechat"},
				},
			},
			&types.Transaction{
				Postings: []types.Posting{
					{Account: "expenses:electronics", Amount: "1719", Commodity: "CNY"},
					{Account: "assets:alipay"},
				},
			},
		},
	}

	f := NewFormatter()
	f.calculateWidths(journal)

	if f.AccountWidth != len("expenses:subscription:icloud") {
		t.Errorf("AccountWidth = %d, want %d", f.AccountWidth, len("expenses:subscription:icloud"))
	}
	if f.AmountWidth != len("1719") {
		t.Errorf("AmountWidth = %d, want %d", f.AmountWidth, len("1719"))
	}
	if f.CommodityWidth != len("CNY") {
		t.Errorf("CommodityWidth = %d, want %d", f.CommodityWidth, len("CNY"))
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/formatter/... -v -run TestCalculateWidths`
Expected: FAIL

**Step 3: 实现列宽计算**

```go
func (f *Formatter) calculateWidths(journal *types.Journal) {
	for _, item := range journal.Items {
		tx, ok := item.(*types.Transaction)
		if !ok {
			continue
		}
		for _, posting := range tx.Postings {
			if len(posting.Account) > f.AccountWidth {
				f.AccountWidth = len(posting.Account)
			}
			if len(posting.Amount) > f.AmountWidth {
				f.AmountWidth = len(posting.Amount)
			}
			if len(posting.Commodity) > f.CommodityWidth {
				f.CommodityWidth = len(posting.Commodity)
			}
		}
	}
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/formatter/... -v -run TestCalculateWidths`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/formatter/formatter.go internal/formatter/formatter_test.go
git commit -m "feat(formatter): implement column width calculation"
```

---

## Task 9: 实现 Posting 格式化（TDD）

**Files:**
- Modify: `internal/formatter/formatter_test.go`
- Modify: `internal/formatter/formatter.go`

**Step 1: 写测试**

```go
func TestFormatPosting(t *testing.T) {
	f := NewFormatter()
	f.AccountWidth = 30
	f.AmountWidth = 6
	f.CommodityWidth = 3

	tests := []struct {
		name    string
		posting types.Posting
		want    string
	}{
		{
			name: "posting with amount",
			posting: types.Posting{
				Account:   "expenses:subscription:icloud",
				Amount:    "21",
				Commodity: "CNY",
			},
			want: "    expenses:subscription:icloud        21 CNY",
		},
		{
			name: "posting without amount",
			posting: types.Posting{
				Account: "assets:wechat",
			},
			want: "    assets:wechat",
		},
		{
			name: "posting with larger amount",
			posting: types.Posting{
				Account:   "expenses:electronics",
				Amount:    "1719",
				Commodity: "CNY",
			},
			want: "    expenses:electronics              1719 CNY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.formatPosting(tt.posting)
			if got != tt.want {
				t.Errorf("formatPosting() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/formatter/... -v -run TestFormatPosting`
Expected: FAIL

**Step 3: 实现**

```go
func (f *Formatter) formatPosting(p types.Posting) string {
	// Account: left-aligned, padded to AccountWidth
	account := fmt.Sprintf("%-*s", f.AccountWidth, p.Account)

	if p.Amount == "" {
		return "    " + account
	}

	// Amount: right-aligned, padded to AmountWidth
	amount := fmt.Sprintf("%*s", f.AmountWidth, p.Amount)

	// Commodity: left-aligned, padded to CommodityWidth
	commodity := fmt.Sprintf("%-*s", f.CommodityWidth, p.Commodity)

	return fmt.Sprintf("    %s %s %s", account, amount, commodity)
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/formatter/... -v -run TestFormatPosting`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/formatter/formatter.go internal/formatter/formatter_test.go
git commit -m "feat(formatter): implement posting formatting with column alignment"
```

---

## Task 10: 实现 Transaction 格式化（TDD）

**Files:**
- Modify: `internal/formatter/formatter_test.go`
- Modify: `internal/formatter/formatter.go`

**Step 1: 写测试**

```go
func TestFormatTransaction(t *testing.T) {
	f := NewFormatter()
	f.AccountWidth = 30
	f.AmountWidth = 6
	f.CommodityWidth = 3

	tx := &types.Transaction{
		Date:        "2026-03-02",
		Status:      "*",
		Description: "Apple iCloud+ 订阅",
		Postings: []types.Posting{
			{Account: "expenses:subscription:icloud", Amount: "21", Commodity: "CNY"},
			{Account: "assets:wechat"},
		},
	}

	want := `2026-03-02 * Apple iCloud+ 订阅
    expenses:subscription:icloud        21 CNY
    assets:wechat`

	got := f.formatTransaction(tx)
	if got != want {
		t.Errorf("formatTransaction() = %q, want %q", got, want)
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/formatter/... -v -run TestFormatTransaction`
Expected: FAIL

**Step 3: 实现**

```go
func (f *Formatter) formatTransaction(tx *types.Transaction) string {
	var sb strings.Builder

	// Header line
	sb.WriteString(fmt.Sprintf("%s %s %s", tx.Date, tx.Status, tx.Description))

	// Add comment if present
	if tx.Comment != "" {
		// Determine if it's a tag or regular comment
		if strings.HasSuffix(tx.Comment, ":") {
			sb.WriteString(fmt.Sprintf(" ;  %s", tx.Comment))
		} else {
			sb.WriteString(fmt.Sprintf(" ; %s", tx.Comment))
		}
	}
	sb.WriteString("\n")

	// Postings
	for _, posting := range tx.Postings {
		sb.WriteString(f.formatPosting(posting))
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}
```

需要添加 `"strings"` import。

**Step 4: 运行测试确认通过**

Run: `go test ./internal/formatter/... -v -run TestFormatTransaction`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/formatter/formatter.go internal/formatter/formatter_test.go
git commit -m "feat(formatter): implement transaction formatting"
```

---

## Task 11: 实现完整的 Format 方法（TDD）

**Files:**
- Modify: `internal/formatter/formatter_test.go`
- Modify: `internal/formatter/formatter.go`

**Step 1: 写测试**

```go
func TestFormat(t *testing.T) {
	journal := &types.Journal{
		Items: []types.Item{
			&types.Comment{Text: "2026年3月交易", IsTag: false},
			&types.EmptyLine{},
			&types.Transaction{
				Date:        "2026-03-02",
				Status:      "*",
				Description: "Apple iCloud+ 订阅",
				Postings: []types.Posting{
					{Account: "expenses:subscription:icloud", Amount: "21", Commodity: "CNY"},
					{Account: "assets:wechat"},
				},
			},
			&types.EmptyLine{},
			&types.Transaction{
				Date:        "2026-03-02",
				Status:      "*",
				Description: "Apple Music 订阅",
				Postings: []types.Posting{
					{Account: "expenses:subscription:music", Amount: "6", Commodity: "CNY"},
					{Account: "assets:wechat"},
				},
			},
		},
	}

	f := NewFormatter()
	got := f.Format(journal)

	// Check that output contains expected elements
	if !strings.Contains(got, "; 2026年3月交易") {
		t.Error("missing comment")
	}
	if !strings.Contains(got, "2026-03-02 * Apple iCloud+ 订阅") {
		t.Error("missing first transaction")
	}
	if !strings.Contains(got, "2026-03-02 * Apple Music 订阅") {
		t.Error("missing second transaction")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/formatter/... -v -run TestFormat`
Expected: FAIL

**Step 3: 实现**

```go
func (f *Formatter) Format(journal *types.Journal) string {
	// First pass: calculate column widths
	f.calculateWidths(journal)

	var sb strings.Builder

	for _, item := range journal.Items {
		switch v := item.(type) {
		case *types.Comment:
			sb.WriteString(f.formatComment(v))
			sb.WriteString("\n")
		case *types.PriceDecl:
			sb.WriteString(f.formatPriceDecl(v))
			sb.WriteString("\n")
		case *types.Transaction:
			sb.WriteString(f.formatTransaction(v))
			sb.WriteString("\n")
		case *types.EmptyLine:
			sb.WriteString("\n")
		}
	}

	return strings.TrimSuffix(sb.String(), "\n") + "\n"
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/formatter/... -v -run TestFormat`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/formatter/formatter.go internal/formatter/formatter_test.go
git commit -m "feat(formatter): implement full Format method"
```

---

## Task 12: 实现交易按日期排序

**Files:**
- Modify: `internal/formatter/formatter_test.go`
- Modify: `internal/formatter/formatter.go`

**Step 1: 写测试**

```go
func TestSortTransactions(t *testing.T) {
	journal := &types.Journal{
		Items: []types.Item{
			&types.Transaction{
				Date:        "2026-03-04",
				Description: "Third",
				Postings:    []types.Posting{{Account: "a"}},
			},
			&types.Comment{Text: "comment"},
			&types.Transaction{
				Date:        "2026-03-02",
				Description: "First",
				Postings:    []types.Posting{{Account: "a"}},
			},
			&types.Transaction{
				Date:        "2026-03-03",
				Description: "Second",
				Postings:    []types.Posting{{Account: "a"}},
			},
		},
	}

	f := NewFormatter()
	sorted := f.sortTransactions(journal)

	tx1 := sorted.Items[0].(*types.Transaction)
	tx2 := sorted.Items[2].(*types.Transaction)
	tx3 := sorted.Items[3].(*types.Transaction)

	if tx1.Description != "First" {
		t.Errorf("first transaction = %q, want %q", tx1.Description, "First")
	}
	if tx2.Description != "Second" {
		t.Errorf("second transaction = %q, want %q", tx2.Description, "Second")
	}
	if tx3.Description != "Third" {
		t.Errorf("third transaction = %q, want %q", tx3.Description, "Third")
	}
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/formatter/... -v -run TestSortTransactions`
Expected: FAIL

**Step 3: 实现**

```go
import "sort"

func (f *Formatter) sortTransactions(journal *types.Journal) *types.Journal {
	// Extract transactions with their positions
	type txWithPos struct {
		tx  *types.Transaction
		pos int
	}
	var transactions []txWithPos

	for i, item := range journal.Items {
		if tx, ok := item.(*types.Transaction); ok {
			transactions = append(transactions, txWithPos{tx: tx, pos: i})
		}
	}

	// Sort by date
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].tx.Date < transactions[j].tx.Date
	})

	// Create new journal with sorted transactions
	newJournal := &types.Journal{Items: make([]types.Item, len(journal.Items))}
	copy(newJournal.Items, journal.Items)

	// Put sorted transactions back
	txIdx := 0
	for i, item := range journal.Items {
		if _, ok := item.(*types.Transaction); ok {
			newJournal.Items[i] = transactions[txIdx].tx
			txIdx++
		}
	}

	return newJournal
}
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/formatter/... -v -run TestSortTransactions`
Expected: PASS

**Step 5: 更新 Format 方法使用排序**

```go
func (f *Formatter) Format(journal *types.Journal) string {
	// Sort transactions by date
	journal = f.sortTransactions(journal)

	// First pass: calculate column widths
	f.calculateWidths(journal)

	// ... rest of the method unchanged
}
```

**Step 6: 运行所有 formatter 测试**

Run: `go test ./internal/formatter/... -v`
Expected: PASS

**Step 7: Commit**

```bash
git add internal/formatter/formatter.go internal/formatter/formatter_test.go
git commit -m "feat(formatter): implement transaction sorting by date"
```

---

## Task 13: 创建 CLI 入口

**Files:**
- Create: `cmd/jf/main.go`

**Step 1: 实现 CLI**

```go
// cmd/jf/main.go

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lin-Jiong-HDU/journal-fmt/internal/formatter"
	"github.com/Lin-Jiong-HDU/journal-fmt/internal/parser"
)

var writeFlag = flag.Bool("w", false, "write result to file instead of stdout")

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: jf [-w] <file.journal> or jf ./...")
		os.Exit(1)
	}

	for _, arg := range args {
		if arg == "./..." {
			if err := walkDir("."); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := processFile(arg); err != nil {
				fmt.Fprintf(os.Stderr, "error processing %s: %v\n", arg, err)
				os.Exit(1)
			}
		}
	}
}

func walkDir(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".journal") {
			return processFile(path)
		}
		return nil
	})
}

func processFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	p := parser.NewParser(string(content))
	journal, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	f := formatter.NewFormatter()
	output := f.Format(journal)

	if *writeFlag {
		return os.WriteFile(filename, []byte(output), 0644)
	}

	_, err = io.WriteString(os.Stdout, output)
	return err
}
```

**Step 2: 验证编译**

Run: `go build ./cmd/jf`
Expected: 无错误

**Step 3: Commit**

```bash
git add cmd/jf/main.go
git commit -m "feat(cli): implement jf command with -w and ./... support"
```

---

## Task 14: 端到端测试

**Step 1: 构建并运行**

```bash
go build -o jf ./cmd/jf
./jf tmp/03-March.journal
```

Expected: 格式化后的内容输出到 stdout

**Step 2: 验证格式化效果**

检查输出：
- 日期格式为 `YYYY-MM-DD`
- 注释格式正确（普通 `; `，tag `;  tag:`）
- 三列对齐
- 交易按日期排序

**Step 3: 测试 -w 参数**

```bash
cp tmp/03-March.journal tmp/test.journal
./jf -w tmp/test.journal
cat tmp/test.journal
```

**Step 4: 测试 ./... 参数**

```bash
./jf -w ./...
```

**Step 5: 清理测试文件**

```bash
rm tmp/test.journal
```

**Step 6: Commit**

```bash
git add -A
git commit -m "chore: verify end-to-end functionality"
```

---

## Task 15: 安装和文档

**Step 1: 安装到 PATH**

```bash
go install ./cmd/jf
```

**Step 2: 验证安装**

```bash
which jf
jf -h
```

**Step 3: 更新 README（可选）**

```markdown
# journal-fmt

A formatter for hledger journal files, similar to `gofmt`.

## Installation

```bash
go install ./cmd/jf
```

## Usage

```bash
# Output to stdout
jf file.journal

# Write back to file
jf -w file.journal

# Format all .journal files in current directory
jf -w ./...
```

## Formatting Rules

- Date format: `YYYY-MM-DD`
- Comments: `; text` for normal, `;  tag:` for tags
- Column alignment: account (left), amount (right), commodity (left)
- Transactions sorted by date
```

**Step 4: Final Commit**

```bash
git add -A
git commit -m "docs: add README and installation instructions"
```

---

## 完成

工具已完成，可以：
1. 运行 `go install ./cmd/jf` 安装
2. 在 Neovim 中配置自动格式化
