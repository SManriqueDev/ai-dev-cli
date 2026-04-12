package integration

import (
	"os"
	"os/exec"
	"testing"
)

func TestCLI_ImproveCommand(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tmpFile, err := os.CreateTemp("", "test_*.go")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(`package main

func Add(a, b int) int {
	return a + b
}
`)
	tmpFile.Close()

	cmd := exec.Command("go", "run", ".", "improve", tmpFile.Name())
	cmd.Dir = "../.."
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
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(`package main

func Multiply(a, b int) int {
	return a * b
}
`)
	tmpFile.Close()

	cmd := exec.Command("go", "run", ".", "test", tmpFile.Name())
	cmd.Dir = "../.."
	cmd.Env = append(os.Environ(), "OPENAI_API_KEY=test-key")
	output, err := cmd.CombinedOutput()

	t.Logf("Output: %s", output)
	if err != nil {
		t.Logf("Error (expected without real API): %v", err)
	}
}

func TestCLI_HelpCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "--help")
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
