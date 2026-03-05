// internal/parser/parser_test.go

package parser

import (
	"testing"

	"github.com/Lin-Jiong-HDU/journal-fmt/internal/types"
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
			comment := journal.Items[0].(*types.Comment)
			if comment.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", comment.Text, tt.wantText)
			}
			if comment.IsTag != tt.wantTag {
				t.Errorf("IsTag = %v, want %v", comment.IsTag, tt.wantTag)
			}
		})
	}
}

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
