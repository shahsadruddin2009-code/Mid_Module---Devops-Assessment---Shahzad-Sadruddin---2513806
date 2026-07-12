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
