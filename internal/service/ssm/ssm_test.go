// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func TestExtractPatchBaselineIDFromARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:     "baseline ID only",
			input:    "pb-0123456789abcdef0",
			expected: "pb-0123456789abcdef0",
		},
		{
			name:     "full ARN",
			input:    "arn:aws:ssm:us-west-2:280605243866:patchbaseline/pb-0123def04827e4e93",
			expected: "pb-0123def04827e4e93",
		},
		{
			name:     "ARN with different region",
			input:    "arn:aws:ssm:eu-west-1:123456789012:patchbaseline/pb-abcdef1234567890",
			expected: "pb-abcdef1234567890",
		},
		{
			name:     "ARN with aws-us-gov partition",
			input:    "arn:aws-us-gov:ssm:us-gov-west-1:123456789012:patchbaseline/pb-fedcba0987654321",
			expected: "pb-fedcba0987654321",
		},
		{
			name:      "empty string",
			input:     "",
			expectErr: true,
		},
		{
			name:      "malformed ARN without slash",
			input:     "arn:aws:ssm:us-west-2:280605243866:patchbaseline",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfssm.ExtractPatchBaselineIDFromARN(tc.input)
			if err != nil && !tc.expectErr {
				t.Fatalf("extractBaselineIDFromARN(%q) returned unexpected error: %v", tc.input, err)
			}
			if result != tc.expected {
				t.Errorf("extractBaselineIDFromARN(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

// These tests affect regional defaults, so they needs to be serialized
func TestAccSSM_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DefaultPatchBaseline": {
			acctest.CtBasic:        testAccSSMDefaultPatchBaseline_basic,
			acctest.CtDisappears:   testAccSSMDefaultPatchBaseline_disappears,
			"otherOperatingSystem": testAccSSMDefaultPatchBaseline_otherOperatingSystem,
			"patchBaselineARN":     testAccSSMDefaultPatchBaseline_patchBaselineARN,
			"systemDefault":        testAccSSMDefaultPatchBaseline_systemDefault,
			"update":               testAccSSMDefaultPatchBaseline_update,
			"deleteDefault":        testAccSSMPatchBaseline_deleteDefault,
			"multiRegion":          testAccSSMDefaultPatchBaseline_multiRegion,
			"wrongOperatingSystem": testAccSSMDefaultPatchBaseline_wrongOperatingSystem,
		},
		"PatchBaseline": {
			"deleteDefault": testAccSSMPatchBaseline_deleteDefault,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
