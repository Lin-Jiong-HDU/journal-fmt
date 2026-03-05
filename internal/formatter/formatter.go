// internal/formatter/formatter.go

package formatter

import (
	"fmt"

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
