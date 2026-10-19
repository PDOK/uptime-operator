package betterstack

import (
	"testing"

	"github.com/PDOK/uptime-operator/internal/model"
)

func TestCheckToMonitorConfirmationPeriod(t *testing.T) {
	tests := []struct {
		name                                 string
		input                                int
		betterStackConfirmationPeriodEnabled bool
		expected                             int
	}{
		{name: "ZeroInput", input: 0, betterStackConfirmationPeriodEnabled: true, expected: 60},
		{name: "NegativeInput", input: -1, betterStackConfirmationPeriodEnabled: true, expected: 60},
		{name: "GreaterThanMaxSupported", input: 31, betterStackConfirmationPeriodEnabled: true, expected: 1830},
		{name: "MuchGreaterThanMaxSupported", input: 1000, betterStackConfirmationPeriodEnabled: true, expected: 1830},
		{name: "ExactMatch_60s", input: 1, betterStackConfirmationPeriodEnabled: true, expected: 90},
		{name: "ExactMatch_120s", input: 2, betterStackConfirmationPeriodEnabled: true, expected: 150},
		{name: "ExactMatch_180s", input: 3, betterStackConfirmationPeriodEnabled: true, expected: 210},
		{name: "ExactMatch_300s", input: 5, betterStackConfirmationPeriodEnabled: true, expected: 330},
		{name: "ExactMatch_600s", input: 10, betterStackConfirmationPeriodEnabled: true, expected: 630},
		{name: "ExactMatch_900s", input: 15, betterStackConfirmationPeriodEnabled: true, expected: 930},
		{name: "ExactMatch_1800s", input: 30, betterStackConfirmationPeriodEnabled: true, expected: 1830},
		{name: "Rounding_240s_roundsTo_180s", input: 4, betterStackConfirmationPeriodEnabled: true, expected: 210},
		{name: "Rounding_360s_roundsTo_300s", input: 6, betterStackConfirmationPeriodEnabled: true, expected: 330},
		{name: "Rounding_420s_roundsTo_300s", input: 7, betterStackConfirmationPeriodEnabled: true, expected: 330},
		{name: "Rounding_480s_roundsTo_600s", input: 8, betterStackConfirmationPeriodEnabled: true, expected: 630},
		{name: "Rounding_960s_roundsTo_900s", input: 16, betterStackConfirmationPeriodEnabled: true, expected: 930},
		{name: "Rounding_1320s_roundsTo_900s", input: 22, betterStackConfirmationPeriodEnabled: true, expected: 930},
		{name: "Rounding_1380s_roundsTo_1800s", input: 23, betterStackConfirmationPeriodEnabled: true, expected: 1830},
		{name: "ZeroInputConfirmationPeriodDisabled", input: 0, betterStackConfirmationPeriodEnabled: false, expected: 0},
		{name: "ExactMatch_60sConfirmationPeriodDisabled", input: 1, betterStackConfirmationPeriodEnabled: false, expected: 0},
		{name: "Rounding_240s_roundsTo_180sConfirmationPeriodDisabled", input: 4, betterStackConfirmationPeriodEnabled: false, expected: 0},
		{name: "GreaterThanMaxSupportedConfirmationPeriodDisabled", input: 31, betterStackConfirmationPeriodEnabled: false, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := checkToMonitor(model.UptimeCheck{
				Interval: tt.input,
			}, tt.betterStackConfirmationPeriodEnabled).ConfirmationPeriod
			if actual != tt.expected {
				t.Errorf("checkToMonitor(Interval: %d).ConfirmationPeriod => expected %d, got %d", tt.input, tt.expected, actual)
			}
		})
	}
}
