package creds

import "testing"

func TestPlanName(t *testing.T) {
	tests := []struct {
		name    string
		subType string
		want    string
	}{
		{name: "max", subType: "claude_max_monthly", want: "Max"},
		{name: "pro", subType: "claude_pro_monthly", want: "Pro"},
		{name: "team", subType: "team_monthly", want: "Team"},
		{name: "enterprise", subType: "enterprise_annual", want: "Enterprise"},
		{name: "empty", subType: "", want: ""},
		{name: "unknown", subType: "free", want: ""},
		{name: "case_insensitive", subType: "CLAUDE_PRO_MONTHLY", want: "Pro"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PlanName(tt.subType)
			if got != tt.want {
				t.Errorf("PlanName(%q) = %q, want %q", tt.subType, got, tt.want)
			}
		})
	}
}
