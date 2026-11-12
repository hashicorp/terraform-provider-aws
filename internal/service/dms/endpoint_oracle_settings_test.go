// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/databasemigrationservice/types"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
)

func TestExpandOracleSettings_trimSpaceInChar(t *testing.T) {
	t.Parallel()

	// trim_space_in_char is a source-only Oracle setting. It must be sent to the
	// API for source endpoints (regression test for #49468, where it was missing
	// from the source case of expandOracleSettings and silently dropped), and it
	// must NOT be sent for target endpoints (it is not valid there).
	testCases := map[string]struct {
		endpointType awstypes.ReplicationEndpointTypeValue
		want         *bool
	}{
		"source endpoint sends the configured value": {
			endpointType: awstypes.ReplicationEndpointTypeValueSource,
			want:         aws.Bool(false),
		},
		"target endpoint does not send it": {
			endpointType: awstypes.ReplicationEndpointTypeValueTarget,
			want:         nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tfMap := map[string]any{
				"trim_space_in_char": false,
			}

			got := tfdms.ExpandOracleSettings(tfMap, tc.endpointType)
			if got == nil {
				t.Fatal("expandOracleSettings returned nil")
			}

			switch {
			case tc.want == nil && got.TrimSpaceInChar != nil:
				t.Fatalf("TrimSpaceInChar = %v, want nil (not valid for %s)", aws.ToBool(got.TrimSpaceInChar), tc.endpointType)
			case tc.want != nil && got.TrimSpaceInChar == nil:
				t.Fatalf("TrimSpaceInChar = nil, want %v for %s", aws.ToBool(tc.want), tc.endpointType)
			case tc.want != nil && aws.ToBool(got.TrimSpaceInChar) != aws.ToBool(tc.want):
				t.Fatalf("TrimSpaceInChar = %v, want %v", aws.ToBool(got.TrimSpaceInChar), aws.ToBool(tc.want))
			}
		})
	}
}
