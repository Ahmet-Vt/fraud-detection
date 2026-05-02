package mcp

import (
	"fmt"
	"strings"
	"time"
)

type rangeBoundKind int

const (
	boundFrom rangeBoundKind = iota
	boundTo
)

const layoutMCPDate = "02.01.2006"

var rangeNow = time.Now

func parseRangeBound(raw string, kind rangeBoundKind) (time.Time, error) {
	s := strings.TrimSpace(raw)
	loc := time.Local

	if s == "" {
		if kind == boundTo {
			return rangeNow().In(loc), nil
		}
		return time.Time{}, fmt.Errorf("from is required")
	}

	t, err := time.ParseInLocation(layoutMCPDate, s, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("use DD.MM.YYYY (e.g. 02.05.2026); omit to for end=now")
	}

	return calendarDayBound(t, kind, loc), nil
}

func calendarDayBound(t time.Time, kind rangeBoundKind, loc *time.Location) time.Time {
	y, m, d := t.Date()
	if kind == boundFrom {
		return time.Date(y, m, d, 0, 0, 0, 0, loc)
	}
	return time.Date(y, m, d, 23, 59, 59, 999999999, loc)
}
