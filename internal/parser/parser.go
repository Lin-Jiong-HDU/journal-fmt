// internal/parser/parser.go

package parser

import (
	"fmt"
	"regexp"
	"strconv"
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

		// Check for transaction (starts with date)
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

		// Unknown item type, skip
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
