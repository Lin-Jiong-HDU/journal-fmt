// internal/parser/parser.go

package parser

import (
	"fmt"
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
