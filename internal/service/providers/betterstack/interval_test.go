package betterstack

import (
	"testing"
)

func TestToSupportedInterval(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "ZeroInput", input: 0, expected: 30},
		{name: "GreaterThanMaxSupported", input: 31, expected: 1800},
		{name: "MuchGreaterThanMaxSupported", input: 1000, expected: 1800},
		{name: "ExactMatch_60s", input: 1, expected: 60},
		{name: "ExactMatch_120s", input: 2, expected: 120},
		{name: "ExactMatch_180s", input: 3, expected: 180},
		{name: "ExactMatch_300s", input: 5, expected: 300},
		{name: "ExactMatch_600s", input: 10, expected: 600},
		{name: "ExactMatch_900s", input: 15, expected: 900},
		{name: "ExactMatch_1800s", input: 30, expected: 1800},
		{name: "Rounding_240s_roundsTo_180s", input: 4, expected: 180},
		{name: "Rounding_360s_roundsTo_300s", input: 6, expected: 300},
		{name: "Rounding_420s_roundsTo_300s", input: 7, expected: 300},
		{name: "Rounding_480s_roundsTo_600s", input: 8, expected: 600},
		{name: "Rounding_960s_roundsTo_900s", input: 16, expected: 900},
		{name: "Rounding_1320s_roundsTo_900s", input: 22, expected: 900},
		{name: "Rounding_1380s_roundsTo_1800s", input: 23, expected: 1800},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := toSupportedInterval(tt.input)
			if actual != tt.expected {
				t.Errorf("toSupportedInterval(%d) => expected %d, got %d", tt.input, tt.expected, actual)
			}
		})
	}
}
