// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testTrue = strconv.FormatBool(true)

func TestConfiguredConfigParameterKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rawConfig cty.Value
		want      map[string]struct{}
		wantOK    bool
	}{
		{
			name:      "missing config_parameter",
			rawConfig: cty.EmptyObjectVal,
			wantOK:    false,
		},
		{
			name: "null config_parameter",
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"config_parameter": cty.NullVal(cty.Set(cty.Object(map[string]cty.Type{
					"parameter_key":   cty.String,
					"parameter_value": cty.String,
				}))),
			}),
			wantOK: false,
		},
		{
			name: "configured parameters",
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"config_parameter": cty.SetVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"parameter_key":   cty.StringVal("auto_mv"),
						"parameter_value": cty.StringVal(testTrue),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"parameter_key":   cty.StringVal("enable_large_strings_opt_in"),
						"parameter_value": cty.StringVal("No"),
					}),
				}),
			}),
			want: map[string]struct{}{
				"auto_mv":                     {},
				"enable_large_strings_opt_in": {},
			},
			wantOK: true,
		},
		{
			name: "explicitly empty configured parameters",
			rawConfig: cty.ObjectVal(map[string]cty.Value{
				"config_parameter": cty.SetValEmpty(cty.Object(map[string]cty.Type{
					"parameter_key":   cty.String,
					"parameter_value": cty.String,
				})),
			}),
			want:   map[string]struct{}{},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, gotOK := configuredConfigParameterKeys(tt.rawConfig)
			if gotOK != tt.wantOK {
				t.Fatalf("configuredConfigParameterKeys() ok = %t, want %t", gotOK, tt.wantOK)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("configuredConfigParameterKeys() len = %d, want %d", len(got), len(tt.want))
			}

			for key := range tt.want {
				if _, ok := got[key]; !ok {
					t.Fatalf("configuredConfigParameterKeys() missing key %q", key)
				}
			}
		})
	}
}

func TestFlattenConfigParametersForState(t *testing.T) {
	t.Parallel()

	apiObjects := []awstypes.ConfigParameter{
		{
			ParameterKey:   aws.String("auto_mv"),
			ParameterValue: aws.String(testTrue),
		},
		{
			ParameterKey:   aws.String("enable_large_strings_opt_in"),
			ParameterValue: aws.String("No"),
		},
	}

	t.Run("filters server-only parameters", func(t *testing.T) {
		t.Parallel()

		got := flattenConfigParametersForState(apiObjects, configuredConfigParameterKeysFromSet(schema.NewSet(
			schema.HashResource(workgroupConfigParameterSchema()),
			[]any{
				map[string]any{
					"parameter_key":   "auto_mv",
					"parameter_value": testTrue,
				},
			},
		)))

		if len(got) != 1 {
			t.Fatalf("flattenConfigParametersForState() len = %d, want 1", len(got))
		}

		gotMap, ok := got[0].(map[string]any)
		if !ok {
			t.Fatalf("flattenConfigParametersForState() element type = %T, want map[string]any", got[0])
		}

		if gotMap["parameter_key"] != "auto_mv" {
			t.Fatalf("flattenConfigParametersForState() parameter_key = %v, want auto_mv", gotMap["parameter_key"])
		}
	})

	t.Run("returns empty when nothing is configured", func(t *testing.T) {
		t.Parallel()

		got := flattenConfigParametersForState(apiObjects, map[string]struct{}{})
		if got != nil {
			t.Fatalf("flattenConfigParametersForState() = %#v, want nil", got)
		}
	})
}

func TestLegacyServerOnlyConfigParameters(t *testing.T) {
	t.Parallel()

	rawState := cty.ObjectVal(map[string]cty.Value{
		"config_parameter": cty.SetVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"parameter_key":   cty.StringVal("require_ssl"),
				"parameter_value": cty.StringVal(testTrue),
			}),
			cty.ObjectVal(map[string]cty.Value{
				"parameter_key":   cty.StringVal("enable_large_strings_opt_in"),
				"parameter_value": cty.StringVal("No"),
			}),
		}),
	})

	t.Run("preserves only legacy unconfigurable parameters", func(t *testing.T) {
		t.Parallel()

		got := legacyServerOnlyConfigParameters(rawState, map[string]struct{}{
			"require_ssl": {},
		})

		if len(got) != 1 {
			t.Fatalf("legacyServerOnlyConfigParameters() len = %d, want 1", len(got))
		}

		gotMap, ok := got[0].(map[string]any)
		if !ok {
			t.Fatalf("legacyServerOnlyConfigParameters() element type = %T, want map[string]any", got[0])
		}

		if gotMap["parameter_key"] != "enable_large_strings_opt_in" {
			t.Fatalf("legacyServerOnlyConfigParameters() parameter_key = %v, want enable_large_strings_opt_in", gotMap["parameter_key"])
		}
	})

	t.Run("does not preserve historical user-managed removals", func(t *testing.T) {
		t.Parallel()

		got := legacyServerOnlyConfigParameters(rawState, map[string]struct{}{})
		if len(got) != 1 {
			t.Fatalf("legacyServerOnlyConfigParameters() len = %d, want 1", len(got))
		}

		gotMap, ok := got[0].(map[string]any)
		if !ok {
			t.Fatalf("legacyServerOnlyConfigParameters() element type = %T, want map[string]any", got[0])
		}

		if gotMap["parameter_key"] != "enable_large_strings_opt_in" {
			t.Fatalf("legacyServerOnlyConfigParameters() parameter_key = %v, want enable_large_strings_opt_in", gotMap["parameter_key"])
		}
	})
}
