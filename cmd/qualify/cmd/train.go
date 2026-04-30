// SPDX-FileCopyrightText: 2026 Scott Friedman
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"github.com/provabl/qualify/internal/localaudit"
	"github.com/provabl/qualify/internal/training"
)

func init() {
	rootCmd.AddCommand(trainCmd())
}

// trainCmd is the parent command for training operations.
func trainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "train",
		Short: "Training module commands",
		Long:  "Manage and run compliance training modules.",
	}
	cmd.AddCommand(trainRequiredCmd())
	cmd.AddCommand(trainStartCmd())
	cmd.AddCommand(trainStatusCmd())
	cmd.AddCommand(trainListCmd())
	return cmd
}

// trainRequiredCmd returns the training modules required for the active attest frameworks.
func trainRequiredCmd() *cobra.Command {
	var attDir string
	var frameworkIDs []string

	cmd := &cobra.Command{
		Use:   "required",
		Short: "Show training modules required for your SRE's active frameworks",
		Long: `Reads .attest/sre.yaml (or --attest-dir) to determine which compliance
frameworks are active, then shows which qualify training modules are required.

Examples:
  # Show required training for current SRE
  qualify train required

  # Check requirements for specific frameworks
  qualify train required --framework nist-800-171-r2 --framework hipaa

  # Use a different attest directory
  qualify train required --attest-dir /path/to/.attest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrainRequired(attDir, frameworkIDs)
		},
	}
	cmd.Flags().StringVar(&attDir, "attest-dir", ".attest", "path to attest directory (.attest/sre.yaml)")
	cmd.Flags().StringArrayVar(&frameworkIDs, "framework", nil, "specific framework IDs (overrides sre.yaml)")
	return cmd
}

// frameworkModuleMap maps framework IDs to the training modules they require.
// This mirrors the database migration 000007 but is available offline (no DB needed).
var frameworkModuleMap = map[string][]string{
	"nist-800-171-r2":       {"cui-fundamentals", "security-awareness", "data-classification"},
	"cmmc-level-1":          {"security-awareness", "cui-fundamentals"},
	"cmmc-level-2":          {"cui-fundamentals", "security-awareness", "data-classification"},
	"cmmc-level-3":          {"cui-fundamentals", "security-awareness", "data-classification"},
	"hipaa":                 {"hipaa-privacy-security", "security-awareness", "data-classification"},
	"ferpa":                 {"ferpa-basics", "security-awareness"},
	"itar":                  {"itar-export-control", "countries-of-concern-awareness"},
	"nih-gds":               {"nih-research-security", "countries-of-concern-awareness", "cui-fundamentals"},
	"nih-research-security": {"nih-research-security"},
	"fedramp-moderate":      {"security-awareness", "data-classification"},
	"fedramp-high":          {"security-awareness", "data-classification"},
	"gdpr":                  {"security-awareness", "data-classification"},
	"fisma-moderate":        {"security-awareness", "data-classification"},
	"cjis":                  {"security-awareness", "data-classification"},
}

// moduleDescriptions provides human-readable names for training modules.
var moduleDescriptions = map[string]string{
	"cui-fundamentals":               "Controlled Unclassified Information (CUI)",
	"hipaa-privacy-security":         "HIPAA Security & Privacy",
	"security-awareness":             "Security Awareness (annual)",
	"ferpa-basics":                   "FERPA",
	"itar-export-control":            "ITAR / Export Controls",
	"data-classification":            "Data Classification",
	"nih-research-security":          "NIH Research Security (NOT-OD-26-017)",
	"countries-of-concern-awareness": "Countries-of-Concern Awareness (NOT-OD-25-083)",
}

// sreYAML is a minimal struct to read active frameworks from .attest/sre.yaml.
type sreYAML struct {
	OrgID      string `yaml:"org_id"`
	Name       string `yaml:"name"`
	Frameworks []struct {
		ID string `yaml:"id"`
	} `yaml:"frameworks"`
}

func runTrainRequired(attDir string, overrideFrameworks []string) error {
	var activeFrameworks []string

	if len(overrideFrameworks) > 0 {
		activeFrameworks = overrideFrameworks
	} else {
		// Read from .attest/sre.yaml
		srePath := filepath.Join(attDir, "sre.yaml")
		data, err := os.ReadFile(srePath) // #nosec G304 — operator-controlled path
		if os.IsNotExist(err) {
			fmt.Printf("No .attest/sre.yaml found at %s\n\n", srePath)
			fmt.Println("To check requirements for specific frameworks:")
			fmt.Println("  qualify train required --framework nist-800-171-r2 --framework hipaa")
			fmt.Println()
			fmt.Println("To initialize attest first:")
			fmt.Println("  attest init --region us-east-1")
			return nil
		}
		if err != nil {
			return fmt.Errorf("read sre.yaml: %w", err)
		}

		var sre sreYAML
		if err := yaml.Unmarshal(data, &sre); err != nil {
			return fmt.Errorf("parse sre.yaml: %w", err)
		}
		for _, fw := range sre.Frameworks {
			activeFrameworks = append(activeFrameworks, fw.ID)
		}

		if len(activeFrameworks) == 0 {
			fmt.Println("No frameworks active in .attest/sre.yaml")
			fmt.Println("Run 'attest frameworks add <id>' to activate frameworks.")
			return nil
		}
		if sre.Name != "" {
			fmt.Printf("SRE: %s\n", sre.Name)
		}
	}

	fmt.Printf("Active frameworks: %s\n\n", strings.Join(activeFrameworks, ", "))

	// Collect required modules (dedup).
	required := make(map[string]bool)
	fwForModule := make(map[string][]string) // module → frameworks that require it
	for _, fwID := range activeFrameworks {
		modules, ok := frameworkModuleMap[fwID]
		if !ok {
			continue
		}
		for _, m := range modules {
			required[m] = true
			fwForModule[m] = append(fwForModule[m], fwID)
		}
	}

	if len(required) == 0 {
		fmt.Println("No training modules required for active frameworks.")
		fmt.Println("(Framework-specific requirements may not be mapped yet.)")
		return nil
	}

	fmt.Printf("Required training modules (%d):\n\n", len(required))

	// Display in logical order.
	order := []string{
		"security-awareness",
		"data-classification",
		"cui-fundamentals",
		"hipaa-privacy-security",
		"ferpa-basics",
		"itar-export-control",
		"nih-research-security",
		"countries-of-concern-awareness",
	}
	for _, m := range order {
		if !required[m] {
			continue
		}
		desc := moduleDescriptions[m]
		if desc == "" {
			desc = m
		}
		fws := strings.Join(fwForModule[m], ", ")
		fmt.Printf("  %-38s  required by: %s\n", desc, fws)
	}

	fmt.Println()
	fmt.Println("Complete training:")
	fmt.Println("  qualify train start <module-id>")
	fmt.Println()
	fmt.Println("Check your current status:")
	fmt.Println("  qualify train status")

	return nil
}

// trainStartCmd starts or resumes an interactive training module.
func trainStartCmd() *cobra.Command {
	var userID string
	var restart bool

	cmd := &cobra.Command{
		Use:   "start <module-id>",
		Short: "Start or resume a training module",
		Long: `Run a training module interactively in the terminal.
Presents each section, then runs the quiz. Completion writes IAM tags
automatically so Cedar PDP grants access on the next request.

Progress is saved automatically — you can quit and resume later.

Examples:
  qualify train start security-awareness
  qualify train start cui-fundamentals
  qualify train start nih-research-security`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrainStart(args[0], userID, restart)
		},
	}
	cmd.Flags().StringVar(&userID, "user", os.Getenv("USER"), "user ID (defaults to $USER)")
	cmd.Flags().BoolVar(&restart, "restart", false, "restart from the beginning even if progress exists")
	return cmd
}

// moduleContent is the parsed JSON structure of training_modules.content.
type moduleContent struct {
	Sections     []contentSection `json:"sections"`
	Quiz         []quizQuestion   `json:"quiz"`
	PassingScore int              `json:"passing_score"`
}

type contentSection struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type quizQuestion struct {
	ID          string   `json:"id"`
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Correct     int      `json:"correct"`
	Explanation string   `json:"explanation"`
}

// trainProgress is saved to ~/.qualify/progress/<module-id>.json
type trainProgress struct {
	ModuleID      string `json:"module_id"`
	SectionIndex  int    `json:"section_index"` // next section to show
	SectionsTotal int    `json:"sections_total"`
}

func runTrainStart(moduleID, userID string, restart bool) error {
	db, err := openDB()
	if err != nil {
		return fmt.Errorf("database not available: %w\n  Start the qualify backend first, or set DB_HOST/DB_USER/DB_PASSWORD", err)
	}
	defer db.Close()

	svc := training.NewService(db)
	ctx := context.Background()

	// Local audit log — always available, never blocks on DB.
	alog, _ := localaudit.New() // error only if home dir unresolvable; treat as nil-safe

	mod, err := svc.GetModule(ctx, moduleID)
	if err != nil {
		return fmt.Errorf("module %q not found: %w\n  Run 'qualify train list' to see available modules", moduleID, err)
	}

	if len(mod.Content) == 0 {
		return fmt.Errorf("module %q has no content — run migrations to populate training content", moduleID)
	}

	var mc moduleContent
	if jsonErr := json.Unmarshal(mod.Content, &mc); jsonErr != nil {
		return fmt.Errorf("parse module content: %w", jsonErr)
	}
	if mc.PassingScore == 0 {
		mc.PassingScore = 80
	}

	// Load or initialise progress.
	progressPath := trainProgressPath(moduleID)
	progress := trainProgress{ModuleID: moduleID, SectionsTotal: len(mc.Sections)}
	if !restart {
		if saved, loadErr := loadProgress(progressPath); loadErr == nil && saved.SectionIndex > 0 {
			fmt.Printf("  Resuming %s from section %d of %d.\n",
				mod.Title, saved.SectionIndex+1, len(mc.Sections))
			fmt.Printf("  (Run 'qualify train start %s --restart' to begin again)\n\n", moduleID)
			progress = *saved
		}
	}

	reader := bufio.NewReader(os.Stdin)
	sepWidth := 70
	if isTTY {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 40 {
			sepWidth = min(w-2, 80)
		}
	}
	sep := strings.Repeat("─", sepWidth)

	// Audit: module started (or resumed).
	if alog != nil {
		alog.ModuleStarted(userID, moduleID)
	}

	// ── Sections ────────────────────────────────────────────────────────────
	if progress.SectionIndex < len(mc.Sections) {
		for i := progress.SectionIndex; i < len(mc.Sections); i++ {
			sec := mc.Sections[i]
			fmt.Printf("\n%s\n", sep)
			fmt.Printf("  %s  —  Section %d of %d: %s\n", mod.Title, i+1, len(mc.Sections), sec.Title)
			fmt.Printf("%s\n\n", sep)

			fmt.Println(renderText(sec.Content))
			fmt.Println()

			// Save progress after each section.
			progress.SectionIndex = i + 1
			_ = saveProgress(progressPath, &progress)
			if alog != nil {
				alog.SectionCompleted(userID, moduleID, i+1, len(mc.Sections))
			}

			if i < len(mc.Sections)-1 {
				fmt.Printf("  [Press Enter to continue — 'q' to save and quit] ")
				line, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(line)) == "q" {
					fmt.Printf("\n  Progress saved. Resume with: qualify train start %s\n", moduleID)
					return nil
				}
			}
		}
	}

	// ── Quiz ─────────────────────────────────────────────────────────────────
	if len(mc.Quiz) == 0 {
		fmt.Printf("\n%s\n  No quiz for this module — marking complete.\n%s\n", sep, sep)
	} else {
		for attempt := 1; attempt <= 2; attempt++ {
			if attempt > 1 {
				fmt.Printf("\n  Let's try again. Review the sections above if needed.\n")
			}

			fmt.Printf("\n%s\n", sep)
			fmt.Printf("  Quiz: %d questions, %d%% to pass\n", len(mc.Quiz), mc.PassingScore)
			fmt.Printf("%s\n", sep)

			correct := 0
			var wrongAnswers []int
			for qi, q := range mc.Quiz {
				fmt.Printf("\n  Q%d: %s\n\n", qi+1, wrapText(q.Question, 68))
				for oi, opt := range q.Options {
					fmt.Printf("    %d) %s\n", oi+1, opt)
				}
				fmt.Printf("\n  Answer [1-%d]: ", len(q.Options))
				line, _ := reader.ReadString('\n')
				ans := parseAnswer(strings.TrimSpace(line), len(q.Options))

				if ans == q.Correct {
					fmt.Printf("  ✓ Correct.\n")
					if q.Explanation != "" {
						fmt.Printf("    %s\n", wrapText(q.Explanation, 68))
					}
					correct++
				} else {
					fmt.Printf("  ✗ Not quite. The answer was %d) %s\n", q.Correct+1, q.Options[q.Correct])
					if q.Explanation != "" {
						fmt.Printf("    %s\n", wrapText(q.Explanation, 68))
					}
					wrongAnswers = append(wrongAnswers, qi+1)
				}
			}

			score := (correct * 100) / len(mc.Quiz)
			passed := score >= mc.PassingScore
			if alog != nil {
				alog.QuizAttempt(userID, moduleID, attempt, score, mc.PassingScore, passed)
			}
			fmt.Printf("\n%s\n", sep)

			if passed {
				fmt.Printf("  Score: %d/%d (%d%%) — PASSED ✓\n", correct, len(mc.Quiz), score)
				fmt.Printf("%s\n\n", sep)

				// Mark complete and write IAM tags.
				if completeErr := svc.CompleteModule(ctx, userID, moduleID, score); completeErr != nil {
					fmt.Printf("  ⚠ Could not record completion: %v\n", completeErr)
					fmt.Printf("    Run 'qualify lab register-role --user %s --role-arn <arn>' to enable IAM tag writes.\n", userID)
				} else {
					if alog != nil {
						alog.ModuleCompleted(userID, moduleID, score)
					}
					tagKey, hasTag := moduleTagMap[moduleID]
					if hasTag {
						if alog != nil {
							alog.IAMTagWritten(userID, moduleID, tagKey)
						}
						fmt.Printf("  IAM tags written:\n")
						fmt.Printf("    %-40s = true\n", tagKey)
						fmt.Printf("    %-40s = <1 year from now>\n", tagKey+"-expiry")
						fmt.Printf("\n  Cedar PDP will grant access on the next request.\n")
					}
				}

				// Remove progress file — module complete.
				_ = os.Remove(progressPath)

				// Suggest next module.
				if next := nextRequired(moduleID); next != "" {
					fmt.Printf("\n  Next up: qualify train start %s\n", next)
				}
				return nil
			}

			fmt.Printf("  Score: %d/%d (%d%%) — need %d%% to pass\n", correct, len(mc.Quiz), score, mc.PassingScore)
			if len(wrongAnswers) > 0 {
				fmt.Printf("  Review sections above, then try again (attempt %d of 2).\n", attempt)
			}
			fmt.Printf("%s\n", sep)

			if attempt == 2 {
				if alog != nil {
					alog.ModuleFailed(userID, moduleID, score, mc.PassingScore)
				}
				fmt.Printf("\n  Module not completed. Review the material and run:\n")
				fmt.Printf("    qualify train start %s --restart\n", moduleID)
				return nil
			}
		}
	}
	return nil
}

// --- helpers -----------------------------------------------------------------

// ANSI terminal codes — applied only when stdout is a TTY.
const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiDim   = "\x1b[2m"
	ansiUnder = "\x1b[4m"
)

// isTTY returns true when stdout is a real terminal (not a pipe or CI log).
// When false, renderText degrades to plain text.
var isTTY = term.IsTerminal(int(os.Stdout.Fd()))

var (
	mdBold    = regexp.MustCompile(`\*\*([^*\n]+)\*\*`)
	mdItalic  = regexp.MustCompile(`\*([^*\n]+)\*`)
	mdCode    = regexp.MustCompile("`([^`\n]+)`")
	mdH1      = regexp.MustCompile(`(?m)^# (.+)$`)
	mdH2      = regexp.MustCompile(`(?m)^## (.+)$`)
	mdH3      = regexp.MustCompile(`(?m)^### (.+)$`)
	mdNumList = regexp.MustCompile(`(?m)^(\d+)\. (.+)$`)
	mdQuote   = regexp.MustCompile(`(?m)^> (.+)$`)
)

// renderText converts markdown-lite content to formatted terminal text.
// Uses ANSI codes when stdout is a TTY; degrades to plain text otherwise.
func renderText(s string) string {
	// Unescape SQL-escaped single quotes.
	s = strings.ReplaceAll(s, "''", "'")

	// Determine terminal width for wrapping (default 78 if not a TTY).
	wrapWidth := 78
	if isTTY {
		if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 40 {
			wrapWidth = min(w-4, 90) // indent of 2 + some margin
		}
	}

	if isTTY {
		// Headers.
		s = mdH1.ReplaceAllString(s, ansiBold+ansiUnder+"$1"+ansiReset)
		s = mdH2.ReplaceAllString(s, ansiBold+"$1"+ansiReset)
		s = mdH3.ReplaceAllString(s, ansiBold+ansiDim+"$1"+ansiReset)
		// Inline formatting.
		s = mdBold.ReplaceAllString(s, ansiBold+"$1"+ansiReset)
		s = mdItalic.ReplaceAllString(s, "$1") // italic unreliable in terminals
		s = mdCode.ReplaceAllString(s, ansiDim+"`$1`"+ansiReset)
		// Blockquote.
		s = mdQuote.ReplaceAllString(s, ansiDim+"  │ $1"+ansiReset)
	} else {
		// Plain text: strip all markers.
		s = mdH1.ReplaceAllString(s, "$1")
		s = mdH2.ReplaceAllString(s, "$1")
		s = mdH3.ReplaceAllString(s, "$1")
		s = mdBold.ReplaceAllString(s, "$1")
		s = mdItalic.ReplaceAllString(s, "$1")
		s = mdCode.ReplaceAllString(s, "`$1`")
		s = mdQuote.ReplaceAllString(s, "  $1")
	}

	// Lists — apply after inline formatting.
	s = mdNumList.ReplaceAllString(s, "  $1. $2")
	s = strings.ReplaceAll(s, "\n- ", "\n  • ")
	s = strings.ReplaceAll(s, "\n  - ", "\n    • ") // nested lists

	// Wrap paragraphs (split on blank lines, preserve list blocks).
	var out strings.Builder
	for _, para := range strings.Split(s, "\n\n") {
		trimmed := strings.TrimSpace(para)
		if trimmed == "" {
			continue
		}
		// Don't re-wrap list items or blockquotes — they may have ANSI codes.
		if strings.Contains(trimmed, "  •") || strings.Contains(trimmed, "  │") ||
			strings.Contains(trimmed, "  1.") || strings.Contains(trimmed, "  2.") {
			out.WriteString(trimmed)
		} else {
			out.WriteString(wrapText(trimmed, wrapWidth))
		}
		out.WriteString("\n\n")
	}
	return strings.TrimRight(out.String(), "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// wrapText wraps text at max width, preserving existing newlines.
func wrapText(s string, width int) string {
	var out strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if utf8.RuneCountInString(line) <= width {
			out.WriteString(line)
			out.WriteRune('\n')
			continue
		}
		words := strings.Fields(line)
		col := 0
		for i, w := range words {
			wl := utf8.RuneCountInString(w)
			if col > 0 && col+1+wl > width {
				out.WriteRune('\n')
				col = 0
			}
			if i > 0 && col > 0 {
				out.WriteRune(' ')
				col++
			}
			out.WriteString(w)
			col += wl
		}
		out.WriteRune('\n')
	}
	return strings.TrimRight(out.String(), "\n")
}

// parseAnswer parses "1"-"4" input to 0-based index. Returns -1 on invalid.
func parseAnswer(s string, max int) int {
	if len(s) != 1 || s[0] < '1' || s[0] > byte('0'+max) {
		return -1
	}
	return int(s[0]-'0') - 1
}

// nextRequired suggests the next required module after completing moduleID.
func nextRequired(moduleID string) string {
	order := []string{
		"security-awareness", "data-classification", "cui-fundamentals",
		"hipaa-privacy-security", "ferpa-basics", "itar-export-control",
		"nih-research-security", "countries-of-concern-awareness",
	}
	for i, m := range order {
		if m == moduleID && i+1 < len(order) {
			return order[i+1]
		}
	}
	return ""
}

// --- progress persistence ----------------------------------------------------

func trainProgressPath(moduleID string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".qualify", "progress", moduleID+".json")
}

func saveProgress(path string, p *trainProgress) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o640)
}

func loadProgress(path string) (*trainProgress, error) {
	data, err := os.ReadFile(path) // #nosec G304 — user's own home dir
	if err != nil {
		return nil, err
	}
	var p trainProgress
	return &p, json.Unmarshal(data, &p)
}

// moduleTagMap is a local copy so trainStartCmd can show which tag was written.
// The authoritative map is in internal/training/service.go.
var moduleTagMap = map[string]string{
	"cui-fundamentals":               "attest:cui-training",
	"hipaa-privacy-security":         "attest:hipaa-training",
	"security-awareness":             "attest:awareness-training",
	"ferpa-basics":                   "attest:ferpa-training",
	"itar-export-control":            "attest:itar-training",
	"data-classification":            "attest:data-class-training",
	"nih-research-security":          "attest:research-security-training",
	"countries-of-concern-awareness": "attest:coc-check-current",
}

// trainStatusCmd shows the current training completion status.
func trainStatusCmd() *cobra.Command {
	var userID string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show training completion status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrainStatus(userID)
		},
	}
	cmd.Flags().StringVar(&userID, "user", "", "user ID or email (admin: check any user)")
	return cmd
}

func runTrainStatus(userID string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	ctx := context.Background()
	query := `
		SELECT m.name, m.title, p.status, p.score, p.completed_at, p.expires_at
		FROM training_modules m
		LEFT JOIN training_progress p ON p.module_id = m.name
		WHERE ($1 = '' OR p.user_id = $1)
		  AND m.required_for_frameworks != '[]'
		ORDER BY m.name`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("query training status: %w", err)
	}
	defer rows.Close()

	fmt.Printf("Training status%s:\n\n", func() string {
		if userID != "" {
			return " for " + userID
		}
		return ""
	}())

	found := false
	for rows.Next() {
		found = true
		var name, title string
		var status, score, completedAt, expiresAt *string
		if scanErr := rows.Scan(&name, &title, &status, &score, &completedAt, &expiresAt); scanErr != nil {
			continue
		}
		if status == nil || *status == "" {
			fmt.Printf("  ✗ %-40s  not started\n", title)
			if unlocks := moduleUnlocks(name); unlocks != "" {
				fmt.Printf("    Unlocks: %s\n", unlocks)
			}
			fmt.Printf("    qualify train start %s\n\n", name)
		} else if *status == "completed" {
			expiry := ""
			if expiresAt != nil {
				expiry = "  (expires " + *expiresAt + ")"
			}
			fmt.Printf("  ✓ %-40s  complete%s\n", title, expiry)
		} else {
			fmt.Printf("  … %-40s  in progress\n", title)
			fmt.Printf("    qualify train start %s\n\n", name)
		}
	}
	if !found {
		fmt.Println("  No training records found.")
		fmt.Println("  Run 'qualify train required' to see what's needed.")
	}
	return rows.Err()
}

// moduleUnlocks returns a human-readable description of what completing a module unlocks.
func moduleUnlocks(moduleID string) string {
	unlocks := map[string]string{
		"security-awareness":             "basic AWS access in all environments",
		"data-classification":            "data-classified S3 buckets and EC2",
		"cui-fundamentals":               "CUI S3 buckets, CUI research OU access",
		"hipaa-privacy-security":         "PHI environments, HIPAA research OU",
		"ferpa-basics":                   "student records environments",
		"itar-export-control":            "ITAR research environments, export-controlled data",
		"nih-research-security":          "NIH controlled-access data (with active DUA)",
		"countries-of-concern-awareness": "NIH controlled-access data (countries-of-concern gate)",
	}
	return unlocks[moduleID]
}

// trainListCmd lists available training modules.
func trainListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available training modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrainList()
		},
	}
}

func runTrainList() error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.QueryContext(context.Background(),
		`SELECT name, title, category, difficulty, estimated_minutes FROM training_modules ORDER BY category, name`)
	if err != nil {
		return fmt.Errorf("list modules: %w", err)
	}
	defer rows.Close()

	fmt.Println("Available training modules:")

	lastCategory := ""
	for rows.Next() {
		var name, title, category, difficulty string
		var mins int
		if err := rows.Scan(&name, &title, &category, &difficulty, &mins); err != nil {
			continue
		}
		if category != lastCategory {
			fmt.Printf("  %s\n", strings.ToUpper(category))
			lastCategory = category
		}
		fmt.Printf("    %-38s  %s (%d min)\n", name, difficulty, mins)
	}
	return rows.Err()
}
