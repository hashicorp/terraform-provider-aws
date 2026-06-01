// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func TestExpandUserJourneyImportID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        string
		wantArn   string
		wantId    string
		wantError bool
	}{
		{
			name:    "valid",
			id:      "arn:aws:resiliencehub:us-west-2:123456789012:system/my-system:abc123,uj-12345", //lintignore:AWSAT003,AWSAT005
			wantArn: "arn:aws:resiliencehub:us-west-2:123456789012:system/my-system:abc123",          //lintignore:AWSAT003,AWSAT005
			wantId:  "uj-12345",
		},
		{
			name:      "missing part",
			id:        "arn:aws:resiliencehub:us-west-2:123456789012:system/my-system:abc123", //lintignore:AWSAT003,AWSAT005
			wantError: true,
		},
		{
			name:      "too many parts",
			id:        "arn,system,extra",
			wantError: true,
		},
		{
			name:      "empty string",
			id:        "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parts, err := flex.ExpandResourceId(tt.id, 2, false)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if parts[0] != tt.wantArn {
				t.Errorf("ARN: got %q, want %q", parts[0], tt.wantArn)
			}
			if parts[1] != tt.wantId {
				t.Errorf("ID: got %q, want %q", parts[1], tt.wantId)
			}
		})
	}
}

func TestExpandServiceFunctionImportID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        string
		wantArn   string
		wantId    string
		wantError bool
	}{
		{
			name:    "valid with slash in ARN",
			id:      "arn:aws:resiliencehub:us-west-2:123456789012:service/my-service:xyz789,sf-99999", //lintignore:AWSAT003,AWSAT005
			wantArn: "arn:aws:resiliencehub:us-west-2:123456789012:service/my-service:xyz789",          //lintignore:AWSAT003,AWSAT005
			wantId:  "sf-99999",
		},
		{
			name:      "using slash delimiter fails",
			id:        "arn:aws:resiliencehub:us-west-2:123456789012:service/my-service:xyz789/sf-99999", //lintignore:AWSAT003,AWSAT005
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parts, err := flex.ExpandResourceId(tt.id, 2, false)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if parts[0] != tt.wantArn {
				t.Errorf("ARN: got %q, want %q", parts[0], tt.wantArn)
			}
			if parts[1] != tt.wantId {
				t.Errorf("ID: got %q, want %q", parts[1], tt.wantId)
			}
		})
	}
}
