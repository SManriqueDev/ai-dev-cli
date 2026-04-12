package config

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

type LinterResult struct {
	Passed bool
	Issues []string
	Output string
}

func RunLinter(path string) (*LinterResult, error) {
	cleanPath := filepath.Clean(path)
	// #nosec G204
	cmd := exec.CommandContext(context.Background(), "golangci-lint", "run", "--new-from-rev=HEAD~", "--", cleanPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &LinterResult{
				Passed: exitErr.ExitCode() == 0,
				Issues: parseLinterOutput(string(output)),
				Output: string(output),
			}, nil
		}
		return nil, err
	}

	return &LinterResult{
		Passed: true,
		Issues: []string{},
		Output: string(output),
	}, nil
}

func parseLinterOutput(output string) []string {
	var issues []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, ".go:") {
			issues = append(issues, line)
		}
	}
	return issues
}
