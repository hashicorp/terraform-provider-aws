// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"testing"
)

// The legality table below is taken verbatim from the AgentCore developer
// guide's rate limit dimensions page, for dimensionKeys
// ["targetName", "toolName", "$.context.jwt.sub"]. Keeping the documented cases
// as the test matrix means this fails loudly if the rule is ever misread.
func TestValidateEntryDimensions_wildcardPositions(t *testing.T) {
	t.Parallel()

	dimensionKeys := []string{"targetName", "toolName", "$.context.jwt.sub"}

	testCases := map[string]struct {
		values     map[string]string
		wantErrors bool
	}{
		"all positions specific": {
			values: map[string]string{"targetName": "target1", "toolName": "readData", "$.context.jwt.sub": "alice"},
		},
		"only last position wildcard": {
			values: map[string]string{"targetName": "target1", "toolName": "readData", "$.context.jwt.sub": "*"},
		},
		"trailing two wildcards": {
			values: map[string]string{"targetName": "target1", "toolName": "*", "$.context.jwt.sub": "*"},
		},
		"all wildcards": {
			values: map[string]string{"targetName": "*", "toolName": "*", "$.context.jwt.sub": "*"},
		},
		"wildcard first, then specific values": {
			values:     map[string]string{"targetName": "*", "toolName": "readData", "$.context.jwt.sub": "alice"},
			wantErrors: true,
		},
		"wildcards then specific value": {
			values:     map[string]string{"targetName": "*", "toolName": "*", "$.context.jwt.sub": "alice"},
			wantErrors: true,
		},
		"wildcard in middle": {
			values:     map[string]string{"targetName": "target1", "toolName": "*", "$.context.jwt.sub": "alice"},
			wantErrors: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := validateEntryDimensions(dimensionKeys, testCase.values)
			if gotErrors := len(got) > 0; gotErrors != testCase.wantErrors {
				t.Errorf("validateEntryDimensions() errors = %v, want errors = %t: %+v", gotErrors, testCase.wantErrors, got)
			}
		})
	}
}

func TestValidateEntryDimensions_keySet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		dimensionKeys []string
		values        map[string]string
		wantSummary   string
	}{
		"exact match": {
			dimensionKeys: []string{"targetName"},
			values:        map[string]string{"targetName": "t1"},
		},
		"missing key": {
			dimensionKeys: []string{"targetName"},
			values:        map[string]string{"toolName": "readData"},
			wantSummary:   "Missing Dimension Key",
		},
		"extra key": {
			dimensionKeys: []string{"targetName"},
			values:        map[string]string{"targetName": "t1", "toolName": "readData"},
			wantSummary:   "Unexpected Dimension Key",
		},
		"missing one of several": {
			dimensionKeys: []string{"targetName", "toolName"},
			values:        map[string]string{"targetName": "t1"},
			wantSummary:   "Missing Dimension Key",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := validateEntryDimensions(testCase.dimensionKeys, testCase.values)

			if testCase.wantSummary == "" {
				if len(got) > 0 {
					t.Fatalf("validateEntryDimensions() = %+v, want no violations", got)
				}
				return
			}

			if len(got) == 0 {
				t.Fatalf("validateEntryDimensions() = no violations, want %q", testCase.wantSummary)
			}
			if got[0].summary != testCase.wantSummary {
				t.Errorf("validateEntryDimensions() summary = %q, want %q", got[0].summary, testCase.wantSummary)
			}
		})
	}
}

// A mismatched key set makes the positional wildcard check meaningless, so it
// must not run and produce a second, confusing diagnostic.
func TestValidateEntryDimensions_keyErrorsSuppressWildcardCheck(t *testing.T) {
	t.Parallel()

	got := validateEntryDimensions(
		[]string{"targetName", "toolName"},
		map[string]string{"targetName": "*", "qualifiedModelId": "m1"},
	)

	for _, violation := range got {
		if violation.summary == "Non-Trailing Wildcard" {
			t.Errorf("validateEntryDimensions() reported a wildcard violation alongside key errors: %+v", got)
		}
	}
}

func TestDimensionKeyPattern(t *testing.T) {
	t.Parallel()

	testCases := map[string]bool{
		// Valid — the fixed arms.
		"targetName":                   true,
		"toolName":                     true,
		"qualifiedModelId":             true,
		"$.context.iam.principal":      true,
		"$.context.iam.sourceIdentity": true,
		// Valid — the JWT arm is open-ended by design.
		"$.context.jwt.sub":            true,
		"$.context.jwt.role":           true,
		"$.context.jwt.custom_claim-1": true,
		// Invalid — issue #49344 guessed these screaming-snake names, none of
		// which the API accepts.
		"IAM_PRINCIPAL_ARN":   false,
		"IAM_SOURCE_IDENTITY": false,
		"OAUTH_CLAIM":         false,
		"MODEL":               false,
		"TARGET":              false,
		"TOOL":                false,
		// Invalid — near misses.
		"targetname":                 false,
		"$.context.jwt.":             false,
		"$.context.iam.principalArn": false,
		"":                           false,
	}

	for input, want := range testCases {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			if got := dimensionKeyPattern.MatchString(input); got != want {
				t.Errorf("dimensionKeyPattern.MatchString(%q) = %t, want %t", input, got, want)
			}
		})
	}
}

func TestGatewayRateLimitImportIDParse(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		id        string
		wantError bool
		wantGW    string
		wantRL    string
	}{
		"authored id": {
			id:     "my-gw-abc1234567,per-caller",
			wantGW: "my-gw-abc1234567",
			wantRL: "per-caller",
		},
		"generated id": {
			id:     "my-gw-abc1234567,rl-0a1b2c3d",
			wantGW: "my-gw-abc1234567",
			wantRL: "rl-0a1b2c3d",
		},
		"missing separator": {
			id:        "my-gw-abc1234567",
			wantError: true,
		},
		"empty second part": {
			id:        "my-gw-abc1234567,",
			wantError: true,
		},
		"empty": {
			id:        "",
			wantError: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, attrs, err := gatewayRateLimitImportID{}.Parse(testCase.id)

			if testCase.wantError {
				if err == nil {
					t.Fatalf("Parse(%q) = %v, want error", testCase.id, attrs)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %s", testCase.id, err)
			}
			if got := attrs["gateway_identifier"]; got != testCase.wantGW {
				t.Errorf("gateway_identifier = %v, want %q", got, testCase.wantGW)
			}
			if got := attrs["rate_limit_id"]; got != testCase.wantRL {
				t.Errorf("rate_limit_id = %v, want %q", got, testCase.wantRL)
			}
		})
	}
}
