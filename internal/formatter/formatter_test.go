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
