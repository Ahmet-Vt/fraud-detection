package mcp

import (
	"testing"
	"time"
)

func TestParseRangeBound_DDMMYYYY_from(t *testing.T) {
	loc := time.Local
	got, err := parseRangeBound("01.05.2026", boundFrom)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 5, 1, 0, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("from: got %v want %v", got, want)
	}
}

func TestParseRangeBound_DDMMYYYY_toEndOfDay(t *testing.T) {
	loc := time.Local
	got, err := parseRangeBound("15.06.2026", boundTo)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 6, 15, 23, 59, 59, 999999999, loc)
	if !got.Equal(want) {
		t.Fatalf("to: got %v want %v", got, want)
	}
}

func TestParseRangeBound_toEmptyMeansFixedNow(t *testing.T) {
	fixed := time.Date(2026, 5, 2, 12, 30, 0, 0, time.Local)
	prev := rangeNow
	rangeNow = func() time.Time { return fixed }
	defer func() { rangeNow = prev }()

	got, err := parseRangeBound("", boundTo)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(fixed) {
		t.Fatalf("empty to: got %v want %v", got, fixed)
	}
}

func TestParseRangeBound_fromEmptyErrors(t *testing.T) {
	_, err := parseRangeBound("", boundFrom)
	if err == nil {
		t.Fatal("expected error for empty from")
	}
}

func TestParseRangeBound_rejectsNonPadded(t *testing.T) {
	_, err := parseRangeBound("1.05.2026", boundFrom)
	if err == nil {
		t.Fatal("expected error for non-DD.MM.YYYY")
	}
}

func TestParseRangeBound_rejectsISO(t *testing.T) {
	_, err := parseRangeBound("2026-05-01", boundFrom)
	if err == nil {
		t.Fatal("expected error for YYYY-MM-DD")
	}
}

func TestParseRangeBound_rejectsSlash(t *testing.T) {
	_, err := parseRangeBound("01/05/2026", boundFrom)
	if err == nil {
		t.Fatal("expected error for slash-separated date")
	}
}
