# journal-fmt

A formatter for [hledger](https://hledger.org/) journal files, similar to `gofmt`.

## Features

- **Column alignment** - Aligns account, amount, and commodity columns across all transactions
- **Date normalization** - Converts dates to `YYYY-MM-DD` format
- **Comment formatting** - Normal comments (`; text`) and tags (`;  tag:`)
- **Transaction sorting** - Sorts transactions by date
- **Batch processing** - Format all `.journal` files in a directory

## Installation

```bash
go install github.com/Lin-Jiong-HDU/journal-fmt/cmd/jf@latest
```

Or build from source:

```bash
git clone https://github.com/Lin-Jiong-HDU/journal-fmt.git
cd journal-fmt
go install ./cmd/jf
```

## Usage

```bash
# Output to stdout
jf file.journal

# Write back to file
jf -w file.journal

# Format all .journal files in current directory
jf -w ./...
```

## Formatting Rules

| Rule | Before | After |
|------|--------|-------|
| Date format | `2026/03/02` | `2026-03-02` |
| Normal comment | `;  comment` | `; comment` |
| Tag comment | `; tag:` | `;  tag:` |
| Amount alignment | Inconsistent | Right-aligned |

### Example

**Before:**

```journal
2026/03/02 * Apple iCloud+ 订阅
    expenses:subscription:icloud      21 CNY
    assets:wechat

2026/03/04 * AirPods Pro 3
    expenses:electronics            1719 CNY
    assets:alipay
```

**After:**

```journal
2026-03-02 * Apple iCloud+ 订阅
    expenses:subscription:icloud      21 CNY
    assets:wechat

2026-03-04 * AirPods Pro 3
    expenses:electronics            1719 CNY
    assets:alipay
```

## Editor Integration

### Neovim (LazyVim / conform.nvim)

Create `~/.config/nvim/lua/plugins/journal-fmt.lua`:

```lua
-- Journal file type recognition
vim.filetype.add({
  extension = {
    journal = "journal",
  },
})

return {
  {
    "stevearc/conform.nvim",
    opts = {
      formatters_by_ft = {
        journal = { "journal-fmt" },
      },
      formatters = {
        ["journal-fmt"] = {
          command = "jf",
          args = { "-w", "$FILENAME" },
          stdin = false,
        },
      },
    },
  },
}
```

After saving, your `.journal` files will be automatically formatted.

### Other Editors

You can configure your editor to run `jf -w <file>` on save.

## License

MIT
