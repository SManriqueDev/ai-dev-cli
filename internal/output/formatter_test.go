package output

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatter_FormatMarkdown(t *testing.T) {
	f := NewFormatter(FormatMarkdown, ThemeDark)
	data := OutputData{
		Content:  "# Test\n\nHello World",
		Original: "# Old\n\nHello",
	}

	result, err := f.Format(data)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestFormatter_FormatPlain(t *testing.T) {
	f := NewFormatter(FormatPlain, ThemeDark)
	data := OutputData{
		Content:  "Hello World",
		Original: "Hello",
	}

	result, err := f.Format(data)
	require.NoError(t, err)
	require.Equal(t, "Hello World", result)
}

func TestFormatter_FormatJSON(t *testing.T) {
	f := NewFormatter(FormatJSON, ThemeDark)
	data := OutputData{
		Content:  "improved code",
		Original: "original code",
	}

	result, err := f.Format(data)
	require.NoError(t, err)
	require.Contains(t, result, "content")
	require.Contains(t, result, "original code")
	require.Contains(t, result, "format")
}

func TestFormatter_FormatYAML(t *testing.T) {
	f := NewFormatter(FormatYAML, ThemeDark)
	data := OutputData{
		Content:  "improved code",
		Original: "original code",
	}

	result, err := f.Format(data)
	require.NoError(t, err)
	require.Contains(t, result, "content:")
	require.Contains(t, result, "original:")
}

func TestFormatDiff_Colored(t *testing.T) {
	original := "func Hello() {\n    return \"Hello\"\n}"
	changed := "func Hello() string {\n    return \"Hello, World\"\n}"

	result := FormatDiff(original, changed, true)
	require.NotEmpty(t, result)
}

func TestFormatDiff_Unified(t *testing.T) {
	original := "line1\nline2\nline3"
	changed := "line1\nmodified\nline3"

	result := FormatDiff(original, changed, false)
	require.NotEmpty(t, result)
	require.Contains(t, result, "modified")
}

func TestComputeChangeSummary(t *testing.T) {
	original := "line1\nline2\nline3"
	changed := "line1\nmodified\nline3\nline4"

	summary := ComputeChangeSummary(original, changed)
	require.Greater(t, summary.LinesAdded, 0)
	require.GreaterOrEqual(t, summary.LinesChanged, 0)
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty", "", 0},
		{"single", "hello", 1},
		{"two lines", "hello\nworld", 2},
		{"three lines", "a\nb\nc", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.input)
			require.Equal(t, tt.expected, len(result))
		})
	}
}
