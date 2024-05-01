// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package duration

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
)

var ErrSyntax = errors.New("invalid syntax")

const (
	pattern = `^(?i)P((?P<years>\d+)Y)?((?P<months>\d+)M)?((?P<days>\d+)D)?$`
)

// Durations supports the year-month-day subset of an RFC 3339 duration
// https://www.rfc-editor.org/rfc/rfc3339
type Duration struct {
	years  int
	months int
	days   int
}

func Parse(s string) (Duration, error) {
	if s == "" || s == "P" {
		return Duration{}, ErrSyntax
	}

	re := regexache.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if match == nil {
		return Duration{}, ErrSyntax
	}

	var duration Duration

	for i, name := range re.SubexpNames() {
		value := match[i]
		if i == 0 || name == "" || value == "" {
			continue
		}

		v, err := strconv.Atoi(value)
		if err != nil {
			return Duration{}, err
		}

		switch name {
		case "years":
			duration.years = v
		case "months":
			duration.months = v
		case "days":
			duration.days = v
		}
	}

	return duration, nil
}

func (d Duration) String() string {
	var b strings.Builder
	b.WriteString("P")
	if d.years > 0 {
		fmt.Fprintf(&b, "%dY", d.years)
	}
	if d.months > 0 {
		fmt.Fprintf(&b, "%dM", d.months)
	}
	if d.days > 0 {
		fmt.Fprintf(&b, "%dD", d.days)
	}
	return b.String()
}

func (d Duration) IsZero() bool {
	return d.years == 0 && d.months == 0 && d.days == 0
}

func (d Duration) equal(o Duration) bool {
	if d.years != o.years || d.months != o.months || d.days != o.days {
		return false
	}
	return true
}

func Sub(t time.Time, d Duration) time.Time {
	return t.AddDate(-d.years, -d.months, -d.days)
}
