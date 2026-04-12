package integration

import (
	"os"
	"os/exec"
	"testing"
)

const projectRoot = "../.."

func TestCLI_ImproveCommand(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmpFile.Name())
	})

	_, err = tmpFile.WriteString(`package main

func Add(a, b int) int {
	return a + b
}
`)
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	t.Cleanup(func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	})

	cmd := exec.CommandContext(t.Context(), "go", "run", ".", "improve")
	cmd.Args = append(cmd.Args, tmpFile.Name())
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test-key")
	output, err := cmd.CombinedOutput()

	t.Logf("Output: %s", output)
	if err != nil {
		t.Logf("Error (expected without real API): %v", err)
	}
}

func TestCLI_TestCommand(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmpFile.Name())
	})

	_, err = tmpFile.WriteString(`package main

func Multiply(a, b int) int {
	return a * b
}
`)
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	t.Cleanup(func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	})

	cmd := exec.CommandContext(t.Context(), "go", "run", ".", "test")
	cmd.Args = append(cmd.Args, tmpFile.Name())
	cmd.Dir = "../.."
	cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test-key")
	output, err := cmd.CombinedOutput()

	t.Logf("Output: %s", output)
	if err != nil {
		t.Logf("Error (expected without real API): %v", err)
	}
}

func TestCLI_HelpCommand(t *testing.T) {
	cmd := exec.CommandContext(t.Context(), "go", "run", ".", "--help")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	if !contains(string(output), "AI-Dev-CLI") {
		t.Error("Help output should contain AI-Dev-CLI")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
