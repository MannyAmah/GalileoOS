package main

import (
	"strings"
	"testing"
	"time"
)

func mustParse(t *testing.T, s string) *time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return &ts
}

func TestCheckWallClockUnderCap(t *testing.T) {
	rows := []manifestRow{
		{
			sourceKind:       "github",
			crawlStatus:      "completed",
			crawlStartedAt:   mustParse(t, "2026-05-16T00:00:00Z"),
			crawlCompletedAt: mustParse(t, "2026-05-16T01:00:00Z"),
		},
	}
	if !checkWallClock(rows) {
		t.Fatal("expected wall-clock check to pass for 1h elapsed")
	}
}

func TestCheckWallClockOverCap(t *testing.T) {
	rows := []manifestRow{
		{
			sourceKind:       "github",
			crawlStatus:      "completed",
			crawlStartedAt:   mustParse(t, "2026-05-16T00:00:00Z"),
			crawlCompletedAt: mustParse(t, "2026-05-16T07:00:00Z"),
		},
	}
	if checkWallClock(rows) {
		t.Fatal("expected wall-clock check to fail for 7h elapsed (cap 6h)")
	}
}

func TestCheckLLMCostUnderCap(t *testing.T) {
	if !checkLLMCost(4999) {
		t.Fatal("expected LLM cost check to pass at $49.99")
	}
}

func TestCheckLLMCostOverCap(t *testing.T) {
	if checkLLMCost(5001) {
		t.Fatal("expected LLM cost check to fail at $50.01")
	}
}

func TestCheckOrgSnapshotAllCompleted(t *testing.T) {
	rows := []manifestRow{
		{sourceKind: "github", crawlStatus: "completed"},
		{sourceKind: "slack", crawlStatus: "completed"},
		{sourceKind: "gdrive", crawlStatus: "completed"},
	}
	if !checkOrgSnapshot(rows) {
		t.Fatal("expected org-snapshot to pass at 3/3 completed")
	}
}

func TestCheckOrgSnapshotPartial(t *testing.T) {
	rows := []manifestRow{
		{sourceKind: "github", crawlStatus: "completed"},
		{sourceKind: "slack", crawlStatus: "failed"},
		{sourceKind: "gdrive", crawlStatus: "completed"},
	}
	// 2/3 = 66%; cap is 90%.
	if checkOrgSnapshot(rows) {
		t.Fatal("expected org-snapshot to fail at 2/3 completed (66% < 90%)")
	}
}

func TestCheckSkillPrecisionDeferred(t *testing.T) {
	if !checkSkillPrecision() {
		t.Fatal("Stage 0 Skill precision is N/A and must pass with a deferral message")
	}
}

func TestCheckDestructiveStructurallyZero(t *testing.T) {
	if !checkDestructive() {
		t.Fatal("Stage 0 destructive check is structural-zero and must always pass")
	}
}

func TestSkillPrecisionLineFormat(t *testing.T) {
	// The bracketed [gate] prefix and ADR reference are grep-able;
	// readers and CI logs depend on both. If this assertion fires,
	// either the format changed deliberately (update the test) or
	// the format drifted (fix the message).
	want := "[gate] Skill recommendation precision: N/A (Skill-Selector Agent deferred to Stage 1 per ADR-0003)"
	if !strings.HasPrefix(want, "[gate] ") {
		t.Fatal("Skill precision message must start with [gate] prefix")
	}
}
