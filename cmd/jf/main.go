// cmd/jf/main.go

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lin-Jiong-HDU/journal-fmt/internal/formatter"
	"github.com/Lin-Jiong-HDU/journal-fmt/internal/parser"
)

var writeFlag = flag.Bool("w", false, "write result to file instead of stdout")

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: jf [-w] <file.journal> or jf ./...")
		os.Exit(1)
	}

	for _, arg := range args {
		if arg == "./..." {
			if err := walkDir("."); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := processFile(arg); err != nil {
				fmt.Fprintf(os.Stderr, "error processing %s: %v\n", arg, err)
				os.Exit(1)
			}
		}
	}
}

func walkDir(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".journal") {
			return processFile(path)
		}
		return nil
	})
}

func processFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	p := parser.NewParser(string(content))
	journal, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	f := formatter.NewFormatter()
	output := f.Format(journal)

	if *writeFlag {
		return os.WriteFile(filename, []byte(output), 0644)
	}

	_, err = io.WriteString(os.Stdout, output)
	return err
}
