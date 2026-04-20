package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"charm.land/glamour/v2"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/term"
)

type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatPlain    Format = "plain"
	FormatJSON     Format = "json"
	FormatYAML     Format = "yaml"
)

type Theme string

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
	ThemeAuto  Theme = "auto"
)

type OutputData struct {
	Content  string `json:"content"`
	Original string `json:"original,omitempty"`
	Format   string `json:"format"`
}

type DiffData struct {
	Original string
	New      string
}

type Formatter interface {
	Format(data OutputData) (string, error)
}

type formatter struct {
	format Format
	theme  Theme
}

func NewFormatter(format Format, theme Theme) Formatter {
	return &formatter{
		format: format,
		theme:  theme,
	}
}

func (f *formatter) Format(data OutputData) (string, error) {
	fmt.Printf("Formatting output with format: %s and theme: %s\n", f.format, f.theme)
	switch f.format {
	case FormatMarkdown:
		return f.formatMarkdown(data.Content)
	case FormatPlain:
		return data.Content, nil
	case FormatJSON:
		return f.formatJSON(data)
	case FormatYAML:
		return f.formatYAML(data)
	default:
		return f.formatMarkdown(data.Content)
	}
}

func (f *formatter) formatMarkdown(content string) (string, error) {
	themeName := string(f.theme)
	if f.theme == ThemeAuto {
		themeName = detectTheme()
	}
	return glamour.Render(content, themeName)
}

func (f *formatter) formatJSON(data OutputData) (string, error) {
	output := OutputData{
		Content:  data.Content,
		Original: data.Original,
		Format:   "json",
	}
	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	return string(jsonBytes), nil
}

func (f *formatter) formatYAML(data OutputData) (string, error) {
	yaml := "---\n"
	yaml += "content: |\n"
	for _, line := range splitLines(data.Content) {
		yaml += "  " + line + "\n"
	}
	if data.Original != "" {
		yaml += "original: |\n"
		for _, line := range splitLines(data.Original) {
			yaml += "  " + line + "\n"
		}
	}
	yaml += "format: yaml\n"
	return yaml, nil
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return splitLinesSimple(s)
}

func splitLinesSimple(s string) []string {
	result := []string{}
	current := ""
	for _, c := range s {
		if c == '\n' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func FormatToPlain(content string) string {
	return content
}

func FormatDiff(original, changed string, colored bool) string {
	d := diffmatchpatch.New()
	diffs := d.DiffMain(original, changed, true)
	diffs = d.DiffCleanupSemantic(diffs)

	var result string
	lines := d.DiffPrettyText(diffs)

	if colored {
		for _, line := range splitLinesSimple(lines) {
			switch {
			case len(line) > 0 && line[0] == '-':
				result += "\033[31m" + line + "\033[0m\n"
			case len(line) > 0 && line[0] == '+':
				result += "\033[32m" + line + "\033[0m\n"
			default:
				result += line + "\n"
			}
		}
	} else {
		result = lines
	}

	return result
}

func detectTheme() string {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return "plain"
	}
	return "dark"
}

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

type ChangeSummary struct {
	LinesAdded   int
	LinesRemoved int
	LinesChanged int
}

func ComputeChangeSummary(original, changed string) ChangeSummary {
	d := diffmatchpatch.New()
	diffs := d.DiffMain(original, changed, false)

	added := 0
	removed := 0

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			added += len(splitLinesSimple(diff.Text))
		case diffmatchpatch.DiffDelete:
			removed += len(splitLinesSimple(diff.Text))
		}
	}

	return ChangeSummary{
		LinesAdded:   added,
		LinesRemoved: removed,
		LinesChanged: added + removed,
	}
}

func ExtractCodeFromMarkdown(markdown string) string {
	var result string
	lines := splitLinesSimple(markdown)
	inCodeBlock := false
	codeStarted := false

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				inCodeBlock = false
				continue
			}
			inCodeBlock = true
			continue
		}
		if inCodeBlock {
			if !codeStarted {
				result = line
				codeStarted = true
			} else {
				result += "\n" + line
			}
		}
	}

	if result == "" {
		return markdown
	}
	return result
}
