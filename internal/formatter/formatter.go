// internal/formatter/formatter.go

package formatter

import (
	"fmt"
	"strings"

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

func (f *Formatter) formatPriceDecl(p *types.PriceDecl) string {
	return fmt.Sprintf("P %s %s %s %s", p.Date, p.Commodity, p.Price, p.TargetCommodity)
}

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

func (f *Formatter) formatPosting(p types.Posting) string {
	// Account: left-aligned, padded to AccountWidth
	account := fmt.Sprintf("%-*s", f.AccountWidth, p.Account)

	if p.Amount == "" {
		return "    " + strings.TrimRight(account, " ")
	}

	// Amount: right-aligned, padded to AmountWidth
	amount := fmt.Sprintf("%*s", f.AmountWidth, p.Amount)

	// Commodity: left-aligned, padded to CommodityWidth
	commodity := fmt.Sprintf("%-*s", f.CommodityWidth, p.Commodity)

	return fmt.Sprintf("    %s %s %s", account, amount, commodity)
}
