// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"slices"
	"testing"
)

func TestStringListSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		values []string
		want   []string
	}{
		"single":            {values: []string{"s3"}, want: []string{"s3"}},
		"comma-separated":   {values: []string{"s3,sqs"}, want: []string{"s3", "sqs"}},
		"spaces and blanks": {values: []string{" s3 , ,sqs,"}, want: []string{"s3", "sqs"}},
		"repeated":          {values: []string{"s3", "sqs,sns"}, want: []string{"s3", "sqs", "sns"}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var l stringList
			for _, v := range tc.values {
				if err := l.Set(v); err != nil {
					t.Fatal(err)
				}
			}
			if !slices.Equal(l, tc.want) {
				t.Errorf("got %v, want %v", l, tc.want)
			}
		})
	}
}
