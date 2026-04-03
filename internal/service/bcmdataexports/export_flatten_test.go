// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdataexports

// Unit tests for flatten behaviour of the export resource.
// These tests run without AWS credentials and guard against regressions where
// the autoflex flatten incorrectly replaces a non-empty outer map with nil
// when its inner maps happen to be empty (e.g. CARBON_EMISSIONS returns
// {"CARBON_EMISSIONS": {}}).

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bcmdataexports/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestFlattenDataQueryTableConfigurations(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := map[string]struct {
		input            *awstypes.DataQuery
		wantNull         bool
		wantOuterKeys    []string
		wantInnerLengths map[string]int
	}{
		// The bug case: AWS returns a non-nil outer map whose only inner map is
		// empty (CARBON_EMISSIONS has no table-configuration keys). Before the
		// fix, autoflex omitempty logic replaced the outer map with nil, causing
		// a perpetual plan diff on re-read.
		"outer map with empty inner map (CARBON_EMISSIONS)": {
			input: &awstypes.DataQuery{
				QueryStatement: aws.String("SELECT usage_account_id FROM CARBON_EMISSIONS"),
				TableConfigurations: map[string]map[string]string{
					"CARBON_EMISSIONS": {},
				},
			},
			wantNull:         false,
			wantOuterKeys:    []string{"CARBON_EMISSIONS"},
			wantInnerLengths: map[string]int{"CARBON_EMISSIONS": 0},
		},
		// Sanity check: a nil TableConfigurations must produce a null map value.
		"nil TableConfigurations": {
			input: &awstypes.DataQuery{
				QueryStatement:      aws.String("SELECT usage_account_id FROM CARBON_EMISSIONS"),
				TableConfigurations: nil,
			},
			wantNull: true,
		},
		// Standard case: non-empty inner map must be flattened normally.
		"non-empty inner map (COST_AND_USAGE_REPORT)": {
			input: &awstypes.DataQuery{
				QueryStatement: aws.String("SELECT identity_line_item_id FROM COST_AND_USAGE_REPORT"),
				TableConfigurations: map[string]map[string]string{
					"COST_AND_USAGE_REPORT": {
						"TIME_GRANULARITY": "HOURLY",
					},
				},
			},
			wantNull:         false,
			wantOuterKeys:    []string{"COST_AND_USAGE_REPORT"},
			wantInnerLengths: map[string]int{"COST_AND_USAGE_REPORT": 1},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var target dataQueryData
			diags := flex.Flatten(ctx, tc.input, &target)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %s", diags)
			}

			if tc.wantNull {
				if !target.TableConfigurations.IsNull() {
					t.Errorf("expected TableConfigurations to be null, got: %s", target.TableConfigurations)
				}
				return
			}

			if target.TableConfigurations.IsNull() {
				t.Fatalf("expected TableConfigurations to be non-null (outer map: %v)", tc.input.TableConfigurations)
			}

			outerElems := target.TableConfigurations.Elements()
			if got, want := len(outerElems), len(tc.wantOuterKeys); got != want {
				t.Errorf("outer map length: got %d, want %d", got, want)
			}

			for _, key := range tc.wantOuterKeys {
				rawInner, ok := outerElems[key]
				if !ok {
					t.Errorf("outer map missing expected key %q", key)
					continue
				}

				innerMap, ok := rawInner.(fwtypes.MapOfString)
				if !ok {
					t.Errorf("inner value for key %q has unexpected type %T", key, rawInner)
					continue
				}

				wantLen := tc.wantInnerLengths[key]
				if got := len(innerMap.Elements()); got != wantLen {
					t.Errorf("inner map for key %q: got %d elements, want %d", key, got, wantLen)
				}
			}
		})
	}
}
