package types

// Journal represents a complete hledger journal file.
type Journal struct {
	Items []Item
}

// Item represents an entry in the journal file.
type Item interface {
	isItem()
}

// Comment represents a comment line.
type Comment struct {
	Text  string // Comment text without semicolon
	IsTag bool   // True if tag format (;  tag:)
}

func (c *Comment) isItem() {}

// PriceDecl represents a price declaration (P directive).
type PriceDecl struct {
	Date             string
	Commodity        string
	Price            string
	TargetCommodity  string
}

func (p *PriceDecl) isItem() {}

// Transaction represents a transaction entry.
type Transaction struct {
	Date        string
	Status      string     // "*", "!", or empty
	Description string
	Postings    []Posting
	Comment     string
}

func (t *Transaction) isItem() {}

// Posting represents an account posting within a transaction.
type Posting struct {
	Account   string
	Amount    string
	Commodity string
	Comment   string
}

// EmptyLine represents a blank line.
type EmptyLine struct{}

func (e *EmptyLine) isItem() {}
