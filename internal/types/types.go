// internal/types/types.go

package types

// Journal 表示整个账本文件
type Journal struct {
	Items []Item
}

// Item 表示文件中的一个条目
type Item interface {
	isItem()
}

// Comment 注释行
type Comment struct {
	Text  string // 注释内容（不含分号）
	IsTag bool   // 是否是 tag（;  tag: 格式）
}

func (c *Comment) isItem() {}

// PriceDecl 价格声明 (P 2026/03/01 CNY 1.00 USD 7.20)
type PriceDecl struct {
	Date            string
	Commodity       string
	Price           string
	TargetCommodity string
}

func (p *PriceDecl) isItem() {}

// Transaction 交易
type Transaction struct {
	Date        string    // 日期
	Status      string    // "*", "!", 或空
	Description string    // 描述
	Postings    []Posting // 分录
	Comment     string    // 行尾注释（可选）
}

func (t *Transaction) isItem() {}

// Posting 分录（账户 + 金额）
type Posting struct {
	Account   string // 账户名
	Amount    string // 金额数值
	Commodity string // 货币单位
	Comment   string // 行尾注释（可选）
}

// EmptyLine 空行
type EmptyLine struct{}

func (e *EmptyLine) isItem() {}
