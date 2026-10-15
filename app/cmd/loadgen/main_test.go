package main

import (
	"os"
	"testing"
)

func TestChooseScenario(t *testing.T) {
	scenarios := []scenario{
		{"success", "/success", 90},
		{"failure", "/failure", 10},
	}

	for i := 0; i < 1000; i++ {
		selected := chooseScenario(scenarios)

		if selected.name != "success" &&
			selected.name != "failure" {
			t.Fatalf("unexpected scenario: %s", selected.name)
		}
	}
}

func TestEnvInt(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int
	}{
		{"valid", "5", 5},
		{"invalid", "abc", 1},
		{"zero", "0", 1},
		{"negative", "-2", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_RPS", tt.env)

			got := envInt("TEST_RPS", 1)
			if got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}

	os.Unsetenv("TEST_RPS")
}
