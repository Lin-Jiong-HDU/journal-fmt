// internal/formatter/formatter_test.go

package formatter

import (
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
