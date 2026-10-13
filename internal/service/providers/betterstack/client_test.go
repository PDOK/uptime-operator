package betterstack

import (
	"testing"

	"github.com/PDOK/uptime-operator/internal/model"
)

func TestCheckToMonitorConfirmationPeriod(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "ZeroInput", input: 0, expected: 31},
		{name: "NegativeInput", input: -1, expected: 31},
		{name: "GreaterThanMaxSupported", input: 31, expected: 1801},
		{name: "MuchGreaterThanMaxSupported", input: 1000, expected: 1801},
		{name: "ExactMatch_60s", input: 1, expected: 61},
		{name: "ExactMatch_120s", input: 2, expected: 121},
		{name: "ExactMatch_180s", input: 3, expected: 181},
		{name: "ExactMatch_300s", input: 5, expected: 301},
		{name: "ExactMatch_600s", input: 10, expected: 601},
		{name: "ExactMatch_900s", input: 15, expected: 901},
		{name: "ExactMatch_1800s", input: 30, expected: 1801},
		{name: "Rounding_240s_roundsTo_180s", input: 4, expected: 181},
		{name: "Rounding_360s_roundsTo_300s", input: 6, expected: 301},
		{name: "Rounding_420s_roundsTo_300s", input: 7, expected: 301},
		{name: "Rounding_480s_roundsTo_600s", input: 8, expected: 601},
		{name: "Rounding_960s_roundsTo_900s", input: 16, expected: 901},
		{name: "Rounding_1320s_roundsTo_900s", input: 22, expected: 901},
		{name: "Rounding_1380s_roundsTo_1800s", input: 23, expected: 1801},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := checkToMonitor(model.UptimeCheck{Interval: tt.input}).ConfirmationPeriod
			if actual != tt.expected {
				t.Errorf("checkToMonitor(Interval: %d).ConfirmationPeriod => expected %d, got %d", tt.input, tt.expected, actual)
			}
		})
	}
}
