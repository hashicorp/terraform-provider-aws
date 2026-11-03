// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"strings"
	"testing"
)

func TestValidateTXTRecordStringSegments(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "short value",
			value:   "v=spf1 include:example.com ~all",
			wantErr: false,
		},
		{
			name:    "exactly 255 characters",
			value:   strings.Repeat("a", 255),
			wantErr: false,
		},
		{
			name:    "256 characters",
			value:   strings.Repeat("a", 256),
			wantErr: true,
		},
		{
			name:    "long value split into valid segments",
			value:   strings.Repeat("a", 255) + "\" \"" + strings.Repeat("b", 255),
			wantErr: false,
		},
		{
			name:    "split with first segment too long",
			value:   strings.Repeat("a", 256) + "\" \"" + strings.Repeat("b", 100),
			wantErr: true,
		},
		{
			name:    "split with second segment too long",
			value:   strings.Repeat("a", 100) + "\" \"" + strings.Repeat("b", 256),
			wantErr: true,
		},
		{
			name:    "three valid segments",
			value:   strings.Repeat("a", 200) + "\" \"" + strings.Repeat("b", 200) + "\" \"" + strings.Repeat("c", 200),
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: false,
		},
		{
			name:    "typical DKIM record with valid split",
			value:   "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAiajKNMp\" \"/A12roF4p3MBm9QxQu6GDsBlWUWFx8EaS8TCo3Qe8Cj0kTag1JMjzCC1s6oM0a43JhO6mp6z/",
			wantErr: false,
		},
		{
			name:    "single character",
			value:   "a",
			wantErr: false,
		},
		{
			name:    "exactly at boundary with split",
			value:   strings.Repeat("x", 255) + "\" \"" + strings.Repeat("y", 1),
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateTXTRecordStringSegments(tc.value)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got: %s", err)
			}
		})
	}
}
