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
