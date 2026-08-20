package migration_test

import (
	"testing"

	"bimbingan-backend/internal/migration"
)

func TestParseSteps(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"1", 1, false},
		{"5", 5, false},
		{"0", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run("input "+tt.input, func(t *testing.T) {
			got, err := migration.ParseSteps(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSteps(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ParseSteps(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunnerInvalidURL(t *testing.T) {
	runner := migration.NewRunner("")
	if err := runner.Up(); err == nil {
		t.Fatal("expected error when dbURL is empty")
	}
}
