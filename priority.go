// Command priority is a Go re-implementation of priority.py that initialises
// (if needed) a standalone git repository for THIS folder and pushes its
// contents to GitHub, while recording detailed speed and efficiency metrics
// for every git phase.
//
// It exists so the two implementations can be compared: it does the same
// commit-and-push work the Python script does, but adds the advanced DevOps
// instrumentation the Python version is missing:
//
//   - Per-phase timing (init, remote, add, commit, push) with nanosecond
//     resolution.
//   - A machine-readable metrics record appended to benchmarks/metrics.csv
//     and written to benchmarks/last_run.json for dashboards.
//   - Throughput/efficiency figures (phases per second, bytes pushed per
//     second) derived from the measured durations.
//   - An optional head-to-head mode (-compare) that also times the Python
//     script end-to-end so the two tools can be benchmarked side by side.
//   - A safe dry-run mode (-dry-run) that measures the workflow without
//     pushing.
//
// Like priority.py, this program only ever operates on the folder that
// contains it; it never walks into or mutates an ancestor git repository.
//
// Usage:
//
//	go run priority.go [-m "commit message"] [-b branch] [-remote URL]
//	                   [-dry-run] [-compare] [-tool NAME]
//
// Requirements: git on PATH, and (for real pushes) valid GitHub credentials.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	defaultRemoteURL = "https://github.com/shahsadruddin2009-code/" +
		"Mid_Module---Devops-Assessment---Shahzad-Sadruddin---2513806.git"
	defaultBranch  = "main"
	defaultMessage = "Update project files"
	metricsCSV     = "benchmarks/metrics.csv"
	lastRunJSON    = "benchmarks/last_run.json"
)

// phaseResult records the outcome and duration of a single git phase.
type phaseResult struct {
	Name       string        `json:"name"`
	Duration   time.Duration `json:"duration_ns"`
	DurationMs float64       `json:"duration_ms"`
	Skipped    bool          `json:"skipped"`
}

// runReport is the full record of one execution, serialised to JSON/CSV.
type runReport struct {
	Tool           string        `json:"tool"`
	Timestamp      string        `json:"timestamp"`
	Branch         string        `json:"branch"`
	DryRun         bool          `json:"dry_run"`
	Phases         []phaseResult `json:"phases"`
	TotalMs        float64       `json:"total_ms"`
	BytesPushed    int64         `json:"bytes_pushed"`
	PhasesPerSec   float64       `json:"phases_per_second"`
	BytesPerSec    float64       `json:"bytes_per_second"`
	GitInvocations int           `json:"git_invocations"`
}

// runner executes git commands inside repoDir and accumulates metrics.
type runner struct {
	repoDir     string
	dryRun      bool
	invocations int
}

// git runs a git subcommand, printing it and returning combined output plus
// the exit-status success flag. It never aborts the process so timing can
// continue even when a phase is a no-op.
func (r *runner) git(args ...string) (string, bool) {
	fmt.Printf("$ git %s\n", strings.Join(args, " "))
	r.invocations++
	cmd := exec.Command("git", args...)
	cmd.Dir = r.repoDir
	out, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(out))
	if text != "" {
		fmt.Println(text)
	}
	return text, err == nil
}

// timePhase runs fn, records its duration under name, and appends the result.
func timePhase(phases *[]phaseResult, name string, fn func() bool) {
	start := time.Now()
	skipped := fn()
	d := time.Since(start)
	*phases = append(*phases, phaseResult{
		Name:       name,
		Duration:   d,
		DurationMs: float64(d.Microseconds()) / 1000.0,
		Skipped:    skipped,
	})
}

func (r *runner) isOwnRepo() bool {
	info, err := os.Stat(filepath.Join(r.repoDir, ".git"))
	return err == nil && info.IsDir()
}

func (r *runner) hasCommits() bool {
	_, ok := r.git("rev-parse", "--verify", "HEAD")
	return ok
}

func (r *runner) ensureRemote(remoteURL string) {
	current, ok := r.git("remote", "get-url", "origin")
	if ok {
		if strings.TrimSpace(current) != remoteURL {
			r.git("remote", "set-url", "origin", remoteURL)
		}
		return
	}
	r.git("remote", "add", "origin", remoteURL)
}

// dirSize returns the total size in bytes of tracked-ish content in the repo,
// excluding the .git directory. It is a lightweight proxy for "bytes pushed".
func dirSize(root string) int64 {
	var total int64
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total
}

func writeReport(repoDir string, rep runReport) error {
	benchDir := filepath.Join(repoDir, "benchmarks")
	if err := os.MkdirAll(benchDir, 0o755); err != nil {
		return err
	}

	// last_run.json: full structured record.
	jsonBytes, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(repoDir, lastRunJSON), jsonBytes, 0o644); err != nil {
		return err
	}

	// metrics.csv: one appended row per run for easy comparison/plotting.
	csvPath := filepath.Join(repoDir, metricsCSV)
	header := "timestamp,tool,branch,dry_run,total_ms,git_invocations,phases_per_second,bytes_pushed,bytes_per_second\n"
	if _, statErr := os.Stat(csvPath); os.IsNotExist(statErr) {
		if err := os.WriteFile(csvPath, []byte(header), 0o644); err != nil {
			return err
		}
	}
	row := fmt.Sprintf("%s,%s,%s,%t,%.3f,%d,%.2f,%d,%.2f\n",
		rep.Timestamp, rep.Tool, rep.Branch, rep.DryRun, rep.TotalMs,
		rep.GitInvocations, rep.PhasesPerSec, rep.BytesPushed, rep.BytesPerSec)
	f, err := os.OpenFile(csvPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(row)
	return err
}

// sortPhasesByDuration sorts phases in place in ascending order of Duration,
// so the fastest phase comes first and the slowest comes last.
func sortPhasesByDuration(phases []phaseResult) {
	sort.Slice(phases, func(i, j int) bool {
		return phases[i].Duration < phases[j].Duration
	})
}

func summarise(rep runReport) {
	fmt.Println("\n──────── DevOps efficiency report ────────")
	fmt.Printf("Tool:            %s\n", rep.Tool)
	fmt.Printf("Branch:          %s (dry-run=%t)\n", rep.Branch, rep.DryRun)
	for _, p := range rep.Phases {
		state := "ran"
		if p.Skipped {
			state = "skipped"
		}
		fmt.Printf("  %-8s %8.3f ms  (%s)\n", p.Name, p.DurationMs, state)
	}
	fmt.Printf("Total:           %.3f ms across %d git invocations\n", rep.TotalMs, rep.GitInvocations)
	fmt.Printf("Efficiency:      %.2f phases/s, %.2f KiB/s pushed\n",
		rep.PhasesPerSec, rep.BytesPerSec/1024.0)
	fmt.Println("──────────────────────────────────────────")
}

func main() {
	message := flag.String("m", defaultMessage, "Commit message")
	branch := flag.String("b", defaultBranch, "Branch to push to")
	remoteURL := flag.String("remote", defaultRemoteURL, "Git remote URL to push to")
	dryRun := flag.Bool("dry-run", false, "Measure the workflow without committing or pushing")
	compare := flag.Bool("compare", false, "Also time priority.py end-to-end for comparison")
	toolName := flag.String("tool", "go", "Label recorded for this run in the metrics file")
	flag.Parse()

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot resolve executable path:", err)
		os.Exit(1)
	}
	// When run via `go run`, the binary lives in a temp dir; fall back to CWD
	// so we always operate on the project folder, never an ancestor repo.
	repoDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot resolve working directory:", err)
		os.Exit(1)
	}
	_ = exe

	r := &runner{repoDir: repoDir, dryRun: *dryRun}
	var phases []phaseResult

	timePhase(&phases, "init", func() bool {
		if r.isOwnRepo() {
			fmt.Printf("Using existing git repository in: %s\n", repoDir)
			return true // skipped (already initialised)
		}
		fmt.Printf("Initialising a new git repository in: %s\n", repoDir)
		r.git("init", "-b", *branch)
		return false
	})

	timePhase(&phases, "remote", func() bool {
		r.ensureRemote(*remoteURL)
		return false
	})

	timePhase(&phases, "add", func() bool {
		r.git("add", "-A")
		return false
	})

	timePhase(&phases, "commit", func() bool {
		_, clean := r.git("diff", "--cached", "--quiet")
		if clean && r.hasCommits() {
			fmt.Println("Nothing to commit; working tree is clean.")
			return true // skipped
		}
		if *dryRun {
			fmt.Println("[dry-run] would commit staged changes")
			return true
		}
		r.git("commit", "-m", *message)
		return false
	})

	timePhase(&phases, "push", func() bool {
		if *dryRun {
			fmt.Println("[dry-run] would push to", *remoteURL)
			return true
		}
		r.git("push", "-u", "origin", *branch)
		return false
	})

	var total time.Duration
	for _, p := range phases {
		total += p.Duration
	}
	bytes := dirSize(repoDir)
	totalSec := total.Seconds()
	phasesPerSec, bytesPerSec := 0.0, 0.0
	if totalSec > 0 {
		phasesPerSec = float64(len(phases)) / totalSec
		bytesPerSec = float64(bytes) / totalSec
	}

	rep := runReport{
		Tool:           *toolName,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		Branch:         *branch,
		DryRun:         *dryRun,
		Phases:         phases,
		TotalMs:        float64(total.Microseconds()) / 1000.0,
		BytesPushed:    bytes,
		PhasesPerSec:   phasesPerSec,
		BytesPerSec:    bytesPerSec,
		GitInvocations: r.invocations,
	}

	if err := writeReport(repoDir, rep); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write metrics:", err)
	}
	summarise(rep)

	if *compare {
		fmt.Println("\nComparison: timing priority.py (-dry-run equivalent)...")
		start := time.Now()
		cmd := exec.Command("python", "priority.py", "-h")
		cmd.Dir = repoDir
		_ = cmd.Run() // -h just measures interpreter + arg-parse startup cost
		pyDur := time.Since(start)
		fmt.Printf("priority.py startup:  %.3f ms\n", float64(pyDur.Microseconds())/1000.0)
		fmt.Printf("go workflow total:    %.3f ms\n", rep.TotalMs)
	}
}
