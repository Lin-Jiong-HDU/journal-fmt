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

		if strings.HasPrefix(strings.TrimSpace(line), "P ") {
			priceDecl, err := p.parsePriceDecl(line)
			if err != nil {
				return nil, err
			}
			journal.Items = append(journal.Items, priceDecl)
			p.pos++
			continue
		}

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

		// Preserve unrecognized lines (account, commodity, include directives, etc.)
		journal.Items = append(journal.Items, &types.RawLine{Content: line})
		p.pos++
	}

	return journal, nil
}

func (p *Parser) parseComment(line string) *types.Comment {
	content := strings.TrimLeft(line, " \t")
	content = strings.TrimPrefix(content, ";")

	// Tag pattern: two spaces followed by text ending with colon
	tagPattern := regexp.MustCompile(`^ {2}(\S+:)\s*$`)
	if matches := tagPattern.FindStringSubmatch(content); matches != nil {
		return &types.Comment{
			Text:  matches[1],
			IsTag: true,
		}
	}

	return &types.Comment{
		Text:  strings.TrimLeft(content, " "),
		IsTag: false,
	}
}

func (p *Parser) parsePriceDecl(line string) (*types.PriceDecl, error) {
	parts := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "P "))
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid price declaration: %s", line)
	}

	return &types.PriceDecl{
		Date:            p.normalizeDate(parts[0]),
		Commodity:       parts[1],
		Price:           parts[2],
		TargetCommodity: strings.Join(parts[3:], " "),
	}, nil
}

func (p *Parser) normalizeDate(date string) string {
	return strings.ReplaceAll(strings.ReplaceAll(date, "/", "-"), ".", "-")
}

func (p *Parser) parseTransaction(line string) (*types.Transaction, error) {
	trimmedLine := strings.TrimSpace(line)

	datePattern := regexp.MustCompile(`^(\d{4}[-/.]\d{2}[-/.]\d{2})`)
	dateMatch := datePattern.FindStringSubmatch(trimmedLine)
	if dateMatch == nil {
		return nil, fmt.Errorf("invalid transaction date: %s", line)
	}
	date := p.normalizeDate(dateMatch[1])

	rest := strings.TrimPrefix(trimmedLine, dateMatch[0])
	rest = strings.TrimSpace(rest)

	status := ""
	if strings.HasPrefix(rest, "*") || strings.HasPrefix(rest, "!") {
		status = string(rest[0])
		rest = strings.TrimSpace(rest[1:])
	}

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

	p.pos++
	for p.pos < len(p.lines) {
		postingLine := p.lines[p.pos]
		if !strings.HasPrefix(postingLine, " ") && !strings.HasPrefix(postingLine, "\t") {
			p.pos--
			break
		}
		posting := p.parsePosting(postingLine)
		tx.Postings = append(tx.Postings, posting)
		p.pos++
	}
	p.pos++

	return tx, nil
}

func (p *Parser) parsePosting(line string) types.Posting {
	trimmedLine := strings.TrimSpace(line)

	comment := ""
	if idx := strings.Index(trimmedLine, ";"); idx != -1 {
		comment = strings.TrimSpace(trimmedLine[idx+1:])
		trimmedLine = strings.TrimSpace(trimmedLine[:idx])
	}

	fields := strings.Fields(trimmedLine)
	if len(fields) == 0 {
		return types.Posting{}
	}

	// Find where account ends and amount begins
	accountEnd := len(fields)
	for i, field := range fields {
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

	if accountEnd < len(fields) {
		posting.Amount = fields[accountEnd]
		if accountEnd+1 < len(fields) {
			posting.Commodity = strings.Join(fields[accountEnd+1:], " ")
		}
	}

	return posting
}
