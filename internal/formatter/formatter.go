package formatter

import (
	"fmt"
	"sort"
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
	journal = f.sortTransactions(journal)
	f.calculateWidths(journal)

	var sb strings.Builder

	for _, item := range journal.Items {
		switch v := item.(type) {
		case *types.Comment:
			sb.WriteString(f.formatComment(v))
			sb.WriteString("\n")
		case *types.PriceDecl:
			sb.WriteString(f.formatPriceDecl(v))
			sb.WriteString("\n")
		case *types.Transaction:
			sb.WriteString(f.formatTransaction(v))
			sb.WriteString("\n")
		case *types.EmptyLine:
			sb.WriteString("\n")
		case *types.RawLine:
			sb.WriteString(v.Content)
			sb.WriteString("\n")
		}
	}

	return strings.TrimSuffix(sb.String(), "\n") + "\n"
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
	account := fmt.Sprintf("%-*s", f.AccountWidth, p.Account)

	if p.Amount == "" {
		result := "    " + strings.TrimRight(account, " ")
		if p.Comment != "" {
			result += " ; " + p.Comment
		}
		return result
	}

	amount := fmt.Sprintf("%*s", f.AmountWidth, p.Amount)
	commodity := fmt.Sprintf("%-*s", f.CommodityWidth, p.Commodity)

	result := fmt.Sprintf("    %s  %s %s", account, amount, commodity)
	if p.Comment != "" {
		result += " ; " + p.Comment
	}
	return result
}

func (f *Formatter) formatTransaction(tx *types.Transaction) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s %s", tx.Date, tx.Status, tx.Description))

	if tx.Comment != "" {
		if strings.HasSuffix(tx.Comment, ":") {
			sb.WriteString(fmt.Sprintf(" ;  %s", tx.Comment))
		} else {
			sb.WriteString(fmt.Sprintf(" ; %s", tx.Comment))
		}
	}
	sb.WriteString("\n")

	for _, posting := range tx.Postings {
		sb.WriteString(f.formatPosting(posting))
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

func (f *Formatter) sortTransactions(journal *types.Journal) *types.Journal {
	type txWithPos struct {
		tx  *types.Transaction
		pos int
	}
	var transactions []txWithPos

	for i, item := range journal.Items {
		if tx, ok := item.(*types.Transaction); ok {
			transactions = append(transactions, txWithPos{tx: tx, pos: i})
		}
	}

	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].tx.Date < transactions[j].tx.Date
	})

	newJournal := &types.Journal{Items: make([]types.Item, len(journal.Items))}
	copy(newJournal.Items, journal.Items)

	txIdx := 0
	for i, item := range journal.Items {
		if _, ok := item.(*types.Transaction); ok {
			newJournal.Items[i] = transactions[txIdx].tx
			txIdx++
		}
	}

	return newJournal
}
