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
