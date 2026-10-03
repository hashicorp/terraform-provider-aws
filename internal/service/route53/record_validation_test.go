// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"strings"
	"testing"
)

func TestValidateTXTStringParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		records []string
		wantErr bool
	}{
		{
			name:    "valid short record",
			records: []string{"v=spf1 include:example.com ~all"},
			wantErr: false,
		},
		{
			name:    "valid record at 255 characters",
			records: []string{strings.Repeat("a", 255)},
			wantErr: false,
		},
		{
			name:    "invalid record at 256 characters",
			records: []string{strings.Repeat("a", 256)},
			wantErr: true,
		},
		{
			name:    "valid split record",
			records: []string{strings.Repeat("a", 255) + "\"\"" + strings.Repeat("b", 255)},
			wantErr: false,
		},
		{
			name:    "invalid first part of split record",
			records: []string{strings.Repeat("a", 256) + "\"\"" + strings.Repeat("b", 100)},
			wantErr: true,
		},
		{
			name:    "invalid second part of split record",
			records: []string{strings.Repeat("a", 100) + "\"\"" + strings.Repeat("b", 256)},
			wantErr: true,
		},
		{
			name:    "multiple valid records",
			records: []string{"short record", strings.Repeat("x", 255)},
			wantErr: false,
		},
		{
			name:    "one valid one invalid record",
			records: []string{"short record", strings.Repeat("x", 300)},
			wantErr: true,
		},
		{
			name:    "empty records",
			records: []string{},
			wantErr: false,
		},
		{
			name:    "DKIM-like long record without split",
			records: []string{"v=DKIM1; k=rsa; p=" + strings.Repeat("A", 300)},
			wantErr: true,
		},
		{
			name:    "DKIM-like record properly split",
			records: []string{"v=DKIM1; k=rsa; p=" + strings.Repeat("A", 237) + "\"\"" + strings.Repeat("A", 63)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateTXTStringParts(tt.records)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTXTStringParts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
