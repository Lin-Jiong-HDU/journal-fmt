# journal-fmt 设计文档

## 概述

类似 `gofmt` 的 hledger 记账文件格式化工具。

## 功能需求

| 功能 | 规则 |
|------|------|
| 金额对齐 | 账户左对齐、金额右对齐、货币左对齐（跨交易） |
| 注释格式 | 普通注释 `; comment`，Tag `;  tag:` |
| 日期格式 | 统一为 `YYYY-MM-DD` |
| 交易排序 | 按日期升序 |

## 项目结构

```
journal-fmt/
├── cmd/
│   └── jf/
│       └── main.go          # CLI 入口
├── internal/
│   ├── parser/
│   │   └── parser.go        # 解析器：.journal → AST
│   ├── formatter/
│   │   └── formatter.go     # 格式化器：AST → 格式化文本
│   └── types/
│       └── types.go         # AST 数据结构定义
├── go.mod
└── go.sum
```

## 数据结构（AST）

```go
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
    Text string   // 注释内容（不含分号）
    IsTag bool    // 是否是 tag（;  tag: 格式）
}

// PriceDecl 价格声明
type PriceDecl struct {
    Date      string
    Commodity string
    Price     string
    TargetCommodity string
}

// Transaction 交易
type Transaction struct {
    Date        string
    Status      string    // "*", "!", 或空
    Description string
    Postings    []Posting
    Comment     string
}

// Posting 分录
type Posting struct {
    Account   string
    Amount    string
    Commodity string
    Comment   string
}
```

## 解析器

```go
type Parser struct {
    lines  []string
    pos    int
}

func NewParser(content string) *Parser
func (p *Parser) Parse() (*types.Journal, error)
```

**解析逻辑：**

```
按行读取
├── 空行 → 跳过，但记录分隔
├── 以 ; 开头 → 解析为 Comment
├── 以 P 开头 → 解析为 PriceDecl
├── 以日期开头 → 解析为 Transaction
│   └── 后续缩进行 → 解析为 Posting
└── 其他 → 保留原样或报错
```

**日期标准化：**
- `2026/03/02` → `2026-03-02`
- `2026.03.02` → `2026-03-02`

## 格式化器

```go
type Formatter struct {
    AccountWidth   int
    AmountWidth    int
    CommodityWidth int
}

func NewFormatter() *Formatter
func (f *Formatter) Format(journal *types.Journal) string
```

**格式化流程：**

```
1. 第一遍扫描：计算每列最大宽度
2. 第二遍扫描：按宽度格式化
   - 账户：左对齐
   - 金额：右对齐
   - 货币：左对齐
3. 按日期排序交易
```

## CLI

```bash
# 输出到 stdout
jf tmp/03-March.journal

# 覆盖原文件
jf -w tmp/03-March.journal

# 格式化当前目录所有 .journal 文件
jf -w ./...
```

**退出码：**
- `0` — 成功
- `1` — 解析错误或文件读取错误

## Neovim 集成

```lua
-- conform.nvim 配置
formatters_by_ft = {
  journal = { "journal-fmt" },
}

formatters = {
  ["journal-fmt"] = {
    command = "jf",
    args = { "-w" },
    stdin = false,
  },
}
```

**安装：**
```bash
go install ./cmd/jf
```
