// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestamp

import (
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
)

// Timestamp is a timestamp string type
type Timestamp string

func New(t string) Timestamp {
	return Timestamp(t)
}

func (t Timestamp) String() string {
	return string(t)
}

// ValidateOnceADayWindowFormat validates once a day window format
func (t Timestamp) ValidateOnceADayWindowFormat() error {
	// valid time format is "hh24:mi"
	validTimeFormat := "([0-1][0-9]|2[0-3]):([0-5][0-9])"
	validTimeFormatConsolidated := "^(" + validTimeFormat + "-" + validTimeFormat + "|)$"

	if !regexache.MustCompile(validTimeFormatConsolidated).MatchString(t.String()) {
		return fmt.Errorf("(%s) must satisfy the format of \"hh24:mi-hh24:mi\"", t.String())
	}

	return nil
}

// ValidateOnceAWeekWindowFormat validates once a week window date format
func (t Timestamp) ValidateOnceAWeekWindowFormat() error {
	// valid time format is "ddd:hh24:mi"
	validTimeFormat := "(sun|mon|tue|wed|thu|fri|sat):([0-1][0-9]|2[0-3]):([0-5][0-9])"
	validTimeFormatConsolidated := "^(" + validTimeFormat + "-" + validTimeFormat + "|)$"

	val := strings.ToLower(t.String())
	if !regexache.MustCompile(validTimeFormatConsolidated).MatchString(val) {
		return fmt.Errorf("(%s) must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\"", val)
	}

	return nil
}

// ValidateUTCFormat parses timestamp in RFC3339 format
func (t Timestamp) ValidateUTCFormat() error {
	_, err := time.Parse(time.RFC3339, t.String())
	if err != nil {
		return fmt.Errorf("must be in RFC3339 time format %q. Example: %s", time.RFC3339, err)
	}

	return nil
}
