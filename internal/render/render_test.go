package render

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestContextColorFunc(t *testing.T) {
	t.Parallel()

	colorFn := ContextColorFunc(80)

	tests := []struct {
		name string
		pct  int
		want string
	}{
		{name: "smart zone 0%", pct: 0, want: Green},
		{name: "smart zone 40%", pct: 40, want: Green},
		{name: "dumb zone 41%", pct: 41, want: Yellow},
		{name: "dumb zone 60%", pct: 60, want: Yellow},
		{name: "danger zone 61%", pct: 61, want: Orange},
		{name: "danger zone 79%", pct: 79, want: Orange},
		{name: "near compaction 80%", pct: 80, want: Red},
		{name: "near compaction 100%", pct: 100, want: Red},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := colorFn(tt.pct)
			if got != tt.want {
				t.Errorf("ContextColorFunc(80)(%d) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestRenderOutput(t *testing.T) {
	t.Parallel()

	subSep := Dim + " · " + Reset
	sep := Dim + " │ " + Reset
	ident := Cyan + "[Opus 4.6 | Pro]" + Reset
	identCwd := ident + sep + Yellow + "myproject" + Reset
	identBranch := ident + sep + Magenta + "feat/foo" + Reset
	identCwdBranch := ident + sep + Yellow + "myproject" + Reset + sep + Magenta + "feat/foo" + Reset

	costStr := "$0.04"

	tests := []struct {
		name            string
		identity        string
		contextBar      string
		usage5h         string
		usage7d         string
		cost            string
		usageExtra      string
		statusIndicator string
		updateIndicator string
		want            string
	}{
		// Minimal: identity + context only.
		{
			name:       "identity and context only",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			want:       ident + sep + "█░░░░ 23%",
		},
		// Identity variants with cwd/branch.
		{
			name:       "with cwd",
			identity:   identCwd,
			contextBar: "█░░░░ 23%",
			want:       identCwd + sep + "█░░░░ 23%",
		},
		{
			name:       "with git branch",
			identity:   identBranch,
			contextBar: "█░░░░ 23%",
			want:       identBranch + sep + "█░░░░ 23%",
		},
		{
			name:       "with cwd and git branch",
			identity:   identCwdBranch,
			contextBar: "█░░░░ 23%",
			want:       identCwdBranch + sep + "█░░░░ 23%",
		},
		// Usage bar combinations.
		{
			name:       "5h only",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			usage5h:    "░░░░░ 9% (13:00)",
			want:       ident + sep + "█░░░░ 23%" + sep + "░░░░░ 9% (13:00)",
		},
		{
			name:       "5h and 7d",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			usage5h:    "░░░░░ 9% (13:00)",
			usage7d:    "█░░░░ 31% (Sun 09:00)",
			want: ident + sep + "█░░░░ 23%" + sep + "░░░░░ 9% (13:00)" + sep +
				"█░░░░ 31% (Sun 09:00)",
		},
		{
			name:       "7d with sub-bars",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			usage5h:    "░░░░░ 9% (13:00)",
			usage7d:    "█░░░░ 31% (Sun 09:00)" + subSep + "░░░░░ 12% son (14:00)",
			want: ident + sep + "█░░░░ 23%" + sep + "░░░░░ 9% (13:00)" + sep +
				"█░░░░ 31% (Sun 09:00)" + subSep + "░░░░░ 12% son (14:00)",
		},
		// Extra usage variants.
		{
			name:       "with extra usage",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			usage5h:    "░░░░░ 9% (13:00)",
			usage7d:    "█░░░░ 31% (Sun 09:00)",
			usageExtra: "$40/$50",
			want: ident + sep + "█░░░░ 23%" + sep + "░░░░░ 9% (13:00)" + sep +
				"█░░░░ 31% (Sun 09:00)" + sep + "$40/$50",
		},
		{
			name:       "with sub-bars and extra usage",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			usage5h:    "░░░░░ 9% (13:00)",
			usage7d:    "█░░░░ 31% (Sun 09:00)" + subSep + "░░░░░ 12% son (14:00)",
			usageExtra: Red + "$45/$50" + Reset,
			want: ident + sep + "█░░░░ 23%" + sep + "░░░░░ 9% (13:00)" + sep +
				"█░░░░ 31% (Sun 09:00)" + subSep + "░░░░░ 12% son (14:00)" + sep +
				Red + "$45/$50" + Reset,
		},
		// Cost variants.
		{
			name:       "with cost only",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			cost:       costStr,
			want:       ident + sep + "█░░░░ 23%" + sep + costStr,
		},
		{
			name:       "cost before extra usage",
			identity:   ident,
			contextBar: "█░░░░ 23%",
			usage5h:    "░░░░░ 9% (13:00)",
			usage7d:    "█░░░░ 31% (Sun 09:00)",
			cost:       costStr,
			usageExtra: "$40/$50",
			want: ident + sep + "█░░░░ 23%" + sep + "░░░░░ 9% (13:00)" + sep +
				"█░░░░ 31% (Sun 09:00)" + sep + costStr + sep + "$40/$50",
		},
		// Full combination: cwd + branch + all bars + cost + extra.
		{
			name:       "all segments",
			identity:   identCwdBranch,
			contextBar: "██░░░ 42%",
			usage5h:    "███░░ 62% (15:00)",
			usage7d:    "█░░░░ 27% (Fri 09:00)" + subSep + "░░░░░ 1% son (Tue 08:00)",
			cost:       costStr,
			usageExtra: "$12/$50",
			want: identCwdBranch + sep + "██░░░ 42%" + sep + "███░░ 62% (15:00)" + sep +
				"█░░░░ 27% (Fri 09:00)" + subSep + "░░░░░ 1% son (Tue 08:00)" + sep + costStr + sep + "$12/$50",
		},
		// Status indicator variants.
		{
			name:            "with status indicator",
			identity:        ident,
			contextBar:      "█░░░░ 23%",
			statusIndicator: Orange + "🔥▂" + Reset,
			want:            ident + sep + "█░░░░ 23%" + sep + Orange + "🔥▂" + Reset,
		},
		{
			name:            "all segments with status indicator",
			identity:        ident,
			contextBar:      "█░░░░ 23%",
			usage5h:         "░░░░░ 9% (13:00)",
			usage7d:         "█░░░░ 31% (Sun 09:00)",
			statusIndicator: Orange + "🔥▆▄▂" + Reset,
			want: ident + sep + "█░░░░ 23%" + sep + "░░░░░ 9% (13:00)" + sep +
				"█░░░░ 31% (Sun 09:00)" + sep + Orange + "🔥▆▄▂" + Reset,
		},
		// Update indicator variants.
		{
			name:            "with update indicator",
			identity:        ident,
			contextBar:      "█░░░░ 23%",
			updateIndicator: Green + "↑" + Reset,
			want:            ident + sep + "█░░░░ 23%" + sep + Green + "↑" + Reset,
		},
		{
			name:            "with status and update indicators",
			identity:        ident,
			contextBar:      "█░░░░ 23%",
			statusIndicator: Orange + "🔥▂" + Reset,
			updateIndicator: Green + "↑" + Reset,
			want:            ident + sep + "█░░░░ 23%" + sep + Orange + "🔥▂" + Reset + sep + Green + "↑" + Reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Output(
				tt.identity,
				tt.contextBar,
				tt.usage5h,
				tt.usage7d,
				tt.cost,
				tt.usageExtra,
				tt.statusIndicator,
				tt.updateIndicator,
			)
			if got != tt.want {
				t.Errorf("Output() =\n  %q\nwant\n  %q", got, tt.want)
			}
		})
	}
}

func TestFormatResetTime(t *testing.T) {
	t.Parallel()

	// Use a fixed "now" for deterministic tests.
	now := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)

	// Test today: should NOT contain day name.
	todayResult := ResetTime("2026-03-09T13:00:00+00:00", now)
	if todayResult == "" {
		t.Fatal("ResetTime returned empty for valid today timestamp")
	}
	if strings.Contains(todayResult, "Mon") || strings.Contains(todayResult, "Sun") ||
		strings.Contains(todayResult, "Tue") {
		t.Errorf("today reset should not contain day name, got %q", todayResult)
	}

	// Test different day: should contain day name.
	futureResult := ResetTime("2026-03-15T08:00:00+00:00", now)
	if futureResult == "" {
		t.Fatal("ResetTime returned empty for valid future timestamp")
	}
	if !strings.Contains(futureResult, "Sun") {
		t.Errorf("future reset should contain day name 'Sun', got %q", futureResult)
	}

	// Test empty.
	emptyResult := ResetTime("", now)
	if emptyResult != "" {
		t.Errorf("ResetTime('') = %q, want empty", emptyResult)
	}
}

func TestStatusIndicator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		indicator string
		want      string
	}{
		{name: "none", indicator: "none", want: ""},
		{name: "empty", indicator: "", want: ""},
		{
			name:      "minor",
			indicator: "minor",
			want:      "\033]8;;https://status.claude.com\a" + Orange + "🔥▂" + Reset + "\033]8;;\a",
		},
		{
			name:      "major",
			indicator: "major",
			want:      "\033]8;;https://status.claude.com\a" + Orange + "🔥▄▂" + Reset + "\033]8;;\a",
		},
		{
			name:      "critical",
			indicator: "critical",
			want:      "\033]8;;https://status.claude.com\a" + Orange + "🔥▆▄▂" + Reset + "\033]8;;\a",
		},
		{name: "unknown", indicator: "maintenance", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := StatusIndicator(tt.indicator)
			if got != tt.want {
				t.Errorf("StatusIndicator(%q) = %q, want %q", tt.indicator, got, tt.want)
			}
		})
	}
}

func TestUpdateIndicator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tag  string
		want string
	}{
		{name: "empty", tag: "", want: ""},
		{
			name: "version tag",
			tag:  "v0.14.0",
			want: "\033]8;;https://github.com/fredrikaverpil/claudeline/releases/tag/v0.14.0\a" +
				Green + "↑" + Reset + "\033]8;;\a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := UpdateIndicator(tt.tag)
			if got != tt.want {
				t.Errorf("UpdateIndicator(%q) = %q, want %q", tt.tag, got, tt.want)
			}
		})
	}
}

func TestExtraUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		used, limit int
		want        string
	}{
		{name: "zero usage hidden", used: 0, limit: 50, want: ""},
		{name: "below threshold", used: 12, limit: 50, want: "$12/$50"},
		{name: "at 80% red", used: 40, limit: 50, want: Red + "$40/$50" + Reset},
		{name: "above 80% red", used: 45, limit: 50, want: Red + "$45/$50" + Reset},
		{name: "zero limit", used: 5, limit: 0, want: "$5/$0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ExtraUsage(tt.used, tt.limit)
			if got != tt.want {
				t.Errorf("ExtraUsage(%d, %d) = %q, want %q", tt.used, tt.limit, got, tt.want)
			}
		})
	}
}

func TestQuotaSubBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		pct       int
		label     string
		resetTime string
		wantPct   string
		wantLabel string
	}{
		{name: "sonnet 12%", pct: 12, label: "sonnet", resetTime: "13:00", wantPct: "12%", wantLabel: "sonnet"},
		{name: "opus 45%", pct: 45, label: "opus", resetTime: "14:00", wantPct: "45%", wantLabel: "opus"},
		{name: "no reset time", pct: 5, label: "cowork", resetTime: "", wantPct: "5%", wantLabel: "cowork"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := QuotaSubBar(tt.pct, tt.label, tt.resetTime)
			if !strings.Contains(got, tt.wantPct) {
				t.Errorf("QuotaSubBar() = %q, missing percentage %q", got, tt.wantPct)
			}
			if !strings.Contains(got, tt.wantLabel) {
				t.Errorf("QuotaSubBar() = %q, missing label %q", got, tt.wantLabel)
			}
			if tt.resetTime != "" && !strings.Contains(got, tt.resetTime) {
				t.Errorf("QuotaSubBar() = %q, missing reset time %q", got, tt.resetTime)
			}
		})
	}
}

func TestBar(t *testing.T) {
	t.Parallel()

	// Use a simple color function that returns a fixed color.
	colorFn := func(int) string { return BrightBlue }

	tests := []struct {
		name      string
		pct       int
		wantFull  int // expected filled block count
		wantEmpty int // expected empty block count
		wantPct   int // expected displayed percentage
	}{
		{name: "0%", pct: 0, wantFull: 0, wantEmpty: 5, wantPct: 0},
		{name: "20% boundary", pct: 20, wantFull: 1, wantEmpty: 4, wantPct: 20},
		{name: "50%", pct: 50, wantFull: 2, wantEmpty: 3, wantPct: 50},
		{name: "100%", pct: 100, wantFull: 5, wantEmpty: 0, wantPct: 100},
		{name: "negative clamped to 0", pct: -1, wantFull: 0, wantEmpty: 5, wantPct: 0},
		{name: "over 100 clamped", pct: 101, wantFull: 5, wantEmpty: 0, wantPct: 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Bar(tt.pct, colorFn)
			fullCount := strings.Count(got, "█")
			emptyCount := strings.Count(got, "░")
			if fullCount != tt.wantFull {
				t.Errorf("Bar(%d) filled blocks = %d, want %d", tt.pct, fullCount, tt.wantFull)
			}
			if emptyCount != tt.wantEmpty {
				t.Errorf("Bar(%d) empty blocks = %d, want %d", tt.pct, emptyCount, tt.wantEmpty)
			}
			if !strings.Contains(got, fmt.Sprintf("%d%%", tt.wantPct)) {
				t.Errorf("Bar(%d) = %q, missing %q", tt.pct, got, fmt.Sprintf("%d%%", tt.wantPct))
			}
		})
	}
}

func TestQuotaColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pct  int
		want string
	}{
		{name: "0% blue", pct: 0, want: BrightBlue},
		{name: "74% blue", pct: 74, want: BrightBlue},
		{name: "75% magenta", pct: 75, want: BrightMagenta},
		{name: "89% magenta", pct: 89, want: BrightMagenta},
		{name: "90% red", pct: 90, want: Red},
		{name: "100% red", pct: 100, want: Red},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := QuotaColor(tt.pct)
			if got != tt.want {
				t.Errorf("QuotaColor(%d) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestIdentity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		model, loginType string
		want             string
	}{
		{
			name:      "login type and model",
			model:     "Opus",
			loginType: "Pro",
			want:      Cyan + "Pro" + Reset + Dim + " │ " + Reset + Cyan + "Opus" + Reset,
		},
		{name: "model only", model: "Sonnet", loginType: "", want: Cyan + "Sonnet" + Reset},
		{name: "both empty", model: "", loginType: "", want: ""},
		{name: "plan only returns empty", model: "", loginType: "Pro", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Identity(tt.loginType, tt.model)
			if got != tt.want {
				t.Errorf("Identity(%q, %q) = %q, want %q", tt.loginType, tt.model, got, tt.want)
			}
		})
	}
}

func TestContextColorFunc_custom_warnPct(t *testing.T) {
	t.Parallel()

	colorFn := ContextColorFunc(85)

	tests := []struct {
		name string
		pct  int
		want string
	}{
		{name: "84% is orange (danger)", pct: 84, want: Orange},
		{name: "85% is red (near compaction)", pct: 85, want: Red},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := colorFn(tt.pct)
			if got != tt.want {
				t.Errorf("ContextColorFunc(85)(%d) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestContextWarnPct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		compactWindow      string
		contextWindowSize  int
		compactPctOverride string
		want               int
	}{
		{
			name:              "default",
			contextWindowSize: 1000000,
			want:              80,
		},
		{
			name:              "absolute compact window",
			compactWindow:     "400000",
			contextWindowSize: 1000000,
			want:              32,
		},
		{
			name:               "percentage override applies to absolute compact window",
			compactWindow:      "400000",
			contextWindowSize:  1000000,
			compactPctOverride: "75",
			want:               28,
		},
		{
			name:               "percentage override applies to full window by default",
			contextWindowSize:  1000000,
			compactPctOverride: "75",
			want:               70,
		},
		{
			name:              "absolute compact window rounds",
			compactWindow:     "333333",
			contextWindowSize: 1000000,
			want:              26,
		},
		{
			name:              "absolute compact window clamps high",
			compactWindow:     "2000000",
			contextWindowSize: 1000000,
			want:              80,
		},
		{
			name:              "absolute compact window clamps low",
			compactWindow:     "1",
			contextWindowSize: 1000000,
			want:              1,
		},
		{
			name:               "invalid absolute compact window uses full window",
			compactWindow:      "not-a-number",
			contextWindowSize:  1000000,
			compactPctOverride: "75",
			want:               70,
		},
		{
			name:               "missing context window size uses full window",
			compactWindow:      "400000",
			compactPctOverride: "75",
			want:               70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := contextWarnPct(tt.compactWindow, tt.contextWindowSize, tt.compactPctOverride)
			if got != tt.want {
				t.Errorf("contextWarnPct(%q, %d, %q) = %d, want %d",
					tt.compactWindow,
					tt.contextWindowSize,
					tt.compactPctOverride,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestCost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		usd  float64
		want string
	}{
		{name: "zero", usd: 0, want: "$0.00"},
		{name: "small", usd: 0.05, want: "$0.05"},
		{name: "typical", usd: 1.23, want: "$1.23"},
		{name: "large", usd: 42.5, want: "$42.50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Cost(tt.usd)
			if got != tt.want {
				t.Errorf("Cost(%v) = %q, want %q", tt.usd, got, tt.want)
			}
		})
	}
}

func TestResetTime_invalid(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	got := ResetTime("not-a-date", now)
	if got != "" {
		t.Errorf("ResetTime(invalid) = %q, want empty", got)
	}
}

func TestResetTimeUnix(t *testing.T) {
	t.Parallel()

	// Use local time to match ResetTimeUnix's .Local() conversion.
	now := time.Date(2026, 3, 9, 10, 0, 0, 0, time.Local)

	sameDayTS := float64(now.Add(time.Hour).Unix())
	diffDayTS := float64(now.Add(48 * time.Hour).Unix())
	wantSameDay := now.Add(time.Hour).Format("15:04")
	wantDiffDay := now.Add(48 * time.Hour).Format("Mon 15:04")

	tests := []struct {
		name string
		ts   *float64
		want string
	}{
		{name: "nil timestamp", ts: nil, want: ""},
		{name: "same day", ts: &sameDayTS, want: wantSameDay},
		{name: "different day", ts: &diffDayTS, want: wantDiffDay},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ResetTimeUnix(tt.ts, now)
			if got != tt.want {
				t.Errorf("ResetTimeUnix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompactName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short name unchanged",
			input:  "main",
			maxLen: 30,
			want:   "main",
		},
		{
			name:   "exactly at limit",
			input:  strings.Repeat("a", 30),
			maxLen: 30,
			want:   strings.Repeat("a", 30),
		},
		{
			name:   "truncated with ellipsis",
			input:  "backup/feat-support-claudeline-progress-tracker",
			maxLen: 30,
			want:   "backup/feat-su…rogress-tracker",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 30,
			want:   "",
		},
		{
			name:   "multibyte unicode",
			input:  "日本語テスト文字列",
			maxLen: 5,
			want:   "日本…字列",
		},
		{
			name:   "maxLen 3",
			input:  "abcdef",
			maxLen: 3,
			want:   "a…f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := compactName(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("compactName(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
			if len([]rune(got)) > tt.maxLen {
				t.Errorf("compactName(%q, %d) rune length = %d, exceeds maxLen", tt.input, tt.maxLen, len([]rune(got)))
			}
		})
	}
}

func TestBuild_HideIdentity(t *testing.T) {
	t.Parallel()

	pct := 25.0

	t.Run("hide subscription plan", func(t *testing.T) {
		t.Parallel()
		got := Build(Params{
			LoginType:      "Pro",
			Model:          "Opus",
			ContextUsedPct: &pct,
			ShowIdentity:   false,
		})
		if !strings.Contains(got, "Opus") {
			t.Errorf("Build() with ShowIdentity=false should contain model %q, got %q", "Opus", got)
		}
		if strings.Contains(got, "Pro") {
			t.Errorf("Build() with ShowIdentity=false should not contain plan %q, got %q", "Pro", got)
		}
	})

	t.Run("hide provider label", func(t *testing.T) {
		t.Parallel()
		got := Build(Params{
			LoginType:      "API",
			Model:          "Sonnet",
			ContextUsedPct: &pct,
			ShowIdentity:   false,
		})
		if !strings.Contains(got, "Sonnet") {
			t.Errorf("Build() with ShowIdentity=false should contain model %q, got %q", "Sonnet", got)
		}
		if strings.Contains(got, "API") {
			t.Errorf("Build() with ShowIdentity=false should not contain provider %q, got %q", "API", got)
		}
	})

	t.Run("show preserves both segments", func(t *testing.T) {
		t.Parallel()
		got := Build(Params{
			LoginType:      "Pro",
			Model:          "Opus",
			ContextUsedPct: &pct,
			ShowIdentity:   true,
		})
		if !strings.Contains(got, "Pro") {
			t.Errorf("Build() with ShowIdentity=true should contain plan %q, got %q", "Pro", got)
		}
		if !strings.Contains(got, "Opus") {
			t.Errorf("Build() with ShowIdentity=true should contain model %q, got %q", "Opus", got)
		}
	})
}

func TestBuild_CacheMiss(t *testing.T) {
	t.Parallel()

	pct := 25.0
	base := Params{
		LoginType:      "Pro",
		Model:          "Opus",
		ContextUsedPct: &pct,
	}

	t.Run("cache miss shows indicator", func(t *testing.T) {
		t.Parallel()
		p := base
		p.CacheMiss = true
		got := Build(p)
		if !strings.Contains(got, "🥊") {
			t.Errorf("Build() with CacheMiss=true should contain 🥊, got %q", got)
		}
	})

	t.Run("cache hit hides indicator", func(t *testing.T) {
		t.Parallel()
		p := base
		p.CacheMiss = false
		got := Build(p)
		// Replace NBSP back to space for comparison.
		normalized := strings.ReplaceAll(got, "\u00A0", " ")
		if strings.Contains(normalized, "🥊") {
			t.Errorf("Build() with CacheMiss=false should not contain 🥊, got %q", got)
		}
	})
}

func TestCwdName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		cwd    string
		maxLen int
		want   string
	}{
		{
			name:   "simple path",
			cwd:    "/Users/fredrik/code/public/claudeline",
			maxLen: 30,
			want:   "claudeline",
		},
		{
			name:   "root path",
			cwd:    "/",
			maxLen: 30,
			want:   "",
		},
		{
			name:   "empty cwd",
			cwd:    "",
			maxLen: 30,
			want:   "",
		},
		{
			name:   "trailing slash",
			cwd:    "/Users/fredrik/code/claudeline/",
			maxLen: 30,
			want:   "claudeline",
		},
		{
			name:   "long name truncated",
			cwd:    "/home/user/my-very-long-project-name-that-exceeds-limit",
			maxLen: 20,
			want:   "my-very-l…eeds-limit",
		},
		{
			name:   "windows path",
			cwd:    `C:\Users\oa\code\claudeline`,
			maxLen: 30,
			want:   "claudeline",
		},
		{
			name:   "home directory",
			cwd:    "/Users/fredrik",
			maxLen: 30,
			want:   "fredrik",
		},
		{
			name:   "windows root C:\\",
			cwd:    `C:\`,
			maxLen: 30,
			want:   "",
		},
		{
			name:   "windows root C:/",
			cwd:    "C:/",
			maxLen: 30,
			want:   "",
		},
		{
			name:   "bare windows drive letter",
			cwd:    "C:",
			maxLen: 30,
			want:   "",
		},
		{
			name:   "dot cwd",
			cwd:    ".",
			maxLen: 30,
			want:   "",
		},
		{
			name:   "backslash only",
			cwd:    `\`,
			maxLen: 30,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := cwdName(tt.cwd, tt.maxLen)
			if got != tt.want {
				t.Errorf("cwdName(%q, %d) = %q, want %q", tt.cwd, tt.maxLen, got, tt.want)
			}
		})
	}
}
