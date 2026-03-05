// internal/formatter/formatter_test.go

package formatter

import (
	"strings"
	"testing"

	"github.com/Lin-Jiong-HDU/journal-fmt/internal/types"
)

func TestFormatComment(t *testing.T) {
	f := NewFormatter()

	tests := []struct {
		name    string
		comment *types.Comment
		want    string
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

func TestFormatPriceDecl(t *testing.T) {
	f := NewFormatter()

	priceDecl := &types.PriceDecl{
		Date:            "2026-03-01",
		Commodity:       "CNY",
		Price:           "1.00",
		TargetCommodity: "USD 7.20",
	}

	want := "P 2026-03-01 CNY 1.00 USD 7.20"
	got := f.formatPriceDecl(priceDecl)
	if got != want {
		t.Errorf("formatPriceDecl() = %q, want %q", got, want)
	}
}

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
			want: "    expenses:subscription:icloud       21 CNY",
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
			want: "    expenses:electronics             1719 CNY",
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
    expenses:subscription:icloud       21 CNY
    assets:wechat`

	got := f.formatTransaction(tx)
	if got != want {
		t.Errorf("formatTransaction() = %q, want %q", got, want)
	}
}

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
