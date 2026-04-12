package config

import (
	"os/exec"
	"strings"
)

type LinterResult struct {
	Passed bool
	Issues []string
	Output string
}

func RunLinter(path string) (*LinterResult, error) {
	cmd := exec.Command("golangci-lint", "run", "--new-from-rev=HEAD~", path)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
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
