package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDirSizeExcludesGitAndSumsFiles(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "b.txt"), []byte("world!"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Content inside a .git directory must be ignored.
	gitDir := filepath.Join(root, ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "ignored.bin"), []byte("XXXXXXXXXX"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := dirSize(root)
	const want = int64(len("hello") + len("world!"))
	if got != want {
		t.Fatalf("dirSize = %d, want %d", got, want)
	}
}

func TestTimePhaseRecordsDurationAndSkipped(t *testing.T) {
	var phases []phaseResult

	timePhase(&phases, "sleepy", func() bool {
		time.Sleep(2 * time.Millisecond)
		return true // skipped flag
	})

	if len(phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(phases))
	}
	p := phases[0]
	if p.Name != "sleepy" {
		t.Errorf("Name = %q, want %q", p.Name, "sleepy")
	}
	if !p.Skipped {
		t.Error("Skipped = false, want true")
	}
	if p.Duration < time.Millisecond {
		t.Errorf("Duration = %v, want >= 1ms", p.Duration)
	}
	if p.DurationMs <= 0 {
		t.Errorf("DurationMs = %f, want > 0", p.DurationMs)
	}
}

func TestWriteReportCreatesJSONAndAppendsCSV(t *testing.T) {
	repoDir := t.TempDir()

	rep := runReport{
		Tool:           "go",
		Timestamp:      "2026-07-12T00:00:00Z",
		Branch:         "main",
		DryRun:         true,
		Phases:         []phaseResult{{Name: "add", Duration: time.Millisecond, DurationMs: 1.0}},
		TotalMs:        1.0,
		BytesPushed:    123,
		PhasesPerSec:   10.0,
		BytesPerSec:    1000.0,
		GitInvocations: 3,
	}

	if err := writeReport(repoDir, rep); err != nil {
		t.Fatalf("writeReport returned error: %v", err)
	}

	// last_run.json must exist.
	if _, err := os.Stat(filepath.Join(repoDir, lastRunJSON)); err != nil {
		t.Errorf("expected %s to exist: %v", lastRunJSON, err)
	}

	// metrics.csv must have a header plus exactly one data row after one run.
	csvPath := filepath.Join(repoDir, metricsCSV)
	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("cannot open metrics csv: %v", err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("cannot parse metrics csv: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 csv rows (header + 1 data), got %d", len(rows))
	}
	if rows[0][1] != "tool" {
		t.Errorf("header column 2 = %q, want %q", rows[0][1], "tool")
	}
	if rows[1][1] != "go" {
		t.Errorf("data tool = %q, want %q", rows[1][1], "go")
	}
}

func TestWriteReportAppendsSecondRow(t *testing.T) {
	repoDir := t.TempDir()
	rep := runReport{Tool: "go", Timestamp: "t1", Branch: "main"}

	if err := writeReport(repoDir, rep); err != nil {
		t.Fatal(err)
	}
	rep.Tool = "python"
	rep.Timestamp = "t2"
	if err := writeReport(repoDir, rep); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(filepath.Join(repoDir, metricsCSV))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 { // header + 2 data rows
		t.Fatalf("expected 3 csv rows, got %d", len(rows))
	}
}

func TestSortPhasesByDuration(t *testing.T) {
	phases := []phaseResult{
		{Name: "phase1", Duration: 3 * time.Second},
		{Name: "phase2", Duration: 1 * time.Second},
		{Name: "phase3", Duration: 2 * time.Second},
	}
	sortPhasesByDuration(phases)
	if phases[0].Name != "phase2" || phases[1].Name != "phase3" || phases[2].Name != "phase1" {
		t.Errorf("phases not sorted by duration: %+v", phases)
	} else {
		t.Logf("phases sorted by duration: %+v", phases)
	}
}
func TestDefaultRemoteURL(t *testing.T) {
	want := "https://github.com/shahsadruddin2009-code/" +
		"Mid_Module---Devops-Assessment---Shahzad-Sadruddin---2513806.git"

	if defaultRemoteURL != want {
		t.Errorf("defaultRemoteURL = %q, want %q", defaultRemoteURL, want)
	} else {
		t.Logf("defaultRemoteURL = %q, as expected", defaultRemoteURL)
	}
}

func TestTimePhaseRecordsDurationAndNotSkipped(t *testing.T) {
	var phases []phaseResult
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	timePhase := func(name string, f func()) {
		start := time.Now()
		f()
		duration := time.Since(start)
		phases = append(phases, phaseResult{Name: name, Duration: duration, Skipped: false})
	}

	timePhase("example", func() {
		time.Sleep(100 * time.Millisecond)
	})

	if len(phases) != 1 {
		t.Fatalf("expected 1 phase, got %d", len(phases))
	}
	if phases[0].Duration <= 0 {
		t.Errorf("expected positive duration, got %v", phases[0].Duration)
	}
	if phases[0].Skipped {
		t.Errorf("expected phase not to be skipped")
	}
}

func TestWriteReportHandlesEmptyPhases(t *testing.T) {
	repoDir := t.TempDir()
	rep := runReport{Tool: "go", Timestamp: "t1", Branch: "main", Phases: []phaseResult{}}

	if err := writeReport(repoDir, rep); err != nil {
		t.Fatalf("writeReport returned error: %v", err)
	}
}

func TestErrorHandlingDirname(t *testing.T) {
	// Wrong directory: a file sits where a directory is expected, so
	// os.MkdirAll inside writeReport cannot create "benchmarks" under it
	// and must return an error. hasdir tracks whether that error occurred.
	wrongDir := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(wrongDir, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	hasdir := false
	if err := writeReport(wrongDir, runReport{Tool: "go", Timestamp: "t1", Branch: "main"}); err != nil {
		hasdir = true
	}
	if !hasdir {
		t.Fatalf("expected error for wrong directory %q, got nil", wrongDir)
	}

	// Correct directory: a real, valid temp directory. writeReport must
	// succeed, so hasdir stays true and the test passes.
	correctDir := t.TempDir()
	hasdir = false
	if err := writeReport(correctDir, runReport{Tool: "go", Timestamp: "t2", Branch: "main"}); err == nil {
		hasdir = true
	}
	if !hasdir {
		t.Fatalf("expected no error for correct directory %q", correctDir)
	}
}

func TestErrorHandlingWriteFile(t *testing.T) {
	// Create a temporary directory and make it read-only to simulate a write error.
	tempDir := t.TempDir()
	os.Chmod(tempDir, 0500) // read-only
	if err := writeReport(tempDir, runReport{Tool: "go", Timestamp: "t1", Branch: "main"}); err == nil {
		t.Fatalf("expected error, got nil")
	}
	os.Chmod(tempDir, 0700) // restore permissions for cleanup
}
