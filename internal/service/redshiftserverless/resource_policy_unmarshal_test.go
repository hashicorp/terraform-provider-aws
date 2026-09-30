// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"encoding/json"
	"testing"

	tfredshiftserverless "github.com/hashicorp/terraform-provider-aws/internal/service/redshiftserverless"
)

// TestResourcePolicyDocUnmarshal verifies that the resource policy document
// unmarshals a Statement provided either as a JSON array (the shape the AWS
// API returns, and the shape documented for the resource) or as a single
// JSON object. Prior to the fix, an array Statement failed to unmarshal with
// "cannot unmarshal array into Go struct field resourcePolicyDoc.Statement".
func TestResourcePolicyDocUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		policy             string
		wantStatementCount int
	}{
		"statement as array": {
			policy: `{
				"Version": "2012-10-17",
				"Statement": [{
					"Effect": "Allow",
					"Principal": {"AWS": ["123456789012"]},
					"Action": ["redshift-serverless:RestoreFromSnapshot"],
					"Sid": ""
				}]
			}`,
			wantStatementCount: 1,
		},
		"statement as array with multiple statements": {
			policy: `{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {"Service": "redshift.amazonaws.com"},
						"Action": "redshift:AuthorizeInboundIntegration",
						"Sid": "Authorize"
					},
					{
						"Effect": "Allow",
						"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
						"Action": "redshift:CreateInboundIntegration",
						"Sid": "Create"
					}
				]
			}`,
			wantStatementCount: 2,
		},
		"statement as single object": {
			policy: `{
				"Version": "2012-10-17",
				"Statement": {
					"Effect": "Allow",
					"Principal": {"AWS": ["123456789012"]},
					"Action": ["redshift-serverless:RestoreFromSnapshot"],
					"Sid": ""
				}
			}`,
			wantStatementCount: 1,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var doc tfredshiftserverless.ResourcePolicyDoc
			if err := json.Unmarshal([]byte(testCase.policy), &doc); err != nil {
				t.Fatalf("unexpected error unmarshaling policy: %s", err)
			}

			if got, want := len(doc.Statement), testCase.wantStatementCount; got != want {
				t.Errorf("statement count = %d, want %d", got, want)
			}
		})
	}
}
