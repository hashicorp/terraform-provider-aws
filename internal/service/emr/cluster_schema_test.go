// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func monitoringConfigurationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"monitoring_configuration": tfemr.ResourceCluster().SchemaMap()["monitoring_configuration"],
	}
}

func TestClusterMonitoringConfigurationValidation(t *testing.T) {
	cloudWatch := []any{map[string]any{names.AttrEnabled: true}}
	s3 := []any{map[string]any{
		"log_type_upload_policy": []any{map[string]any{
			"log_type":      "system-logs",
			"upload_policy": "emr-managed",
		}},
	}}

	testCases := map[string]struct {
		configuration []any
		wantError     bool
	}{
		"omitted":           {},
		"empty":             {configuration: []any{map[string]any{}}, wantError: true},
		"CloudWatch only":   {configuration: []any{map[string]any{"cloud_watch_log_configuration": cloudWatch}}},
		"S3 only":           {configuration: []any{map[string]any{"s3_logging_configuration": s3}}},
		"CloudWatch and S3": {configuration: []any{map[string]any{"cloud_watch_log_configuration": cloudWatch, "s3_logging_configuration": s3}}},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			raw := map[string]any{}
			if testCase.configuration != nil {
				raw["monitoring_configuration"] = testCase.configuration
			}

			diags := schema.InternalMap(monitoringConfigurationSchema()).Validate(terraform.NewResourceConfigRaw(raw))

			if got := diags.HasError(); got != testCase.wantError {
				t.Fatalf("expected error %t, got diagnostics: %#v", testCase.wantError, diags)
			}
			if !testCase.wantError {
				return
			}

			for _, diag := range diags {
				if strings.Contains(diag.Detail, "cloud_watch_log_configuration") && strings.Contains(diag.Detail, "s3_logging_configuration") {
					return
				}
			}
			t.Fatalf("expected an error naming both configuration blocks, got diagnostics: %#v", diags)
		})
	}
}

// Amazon EMR models log types and upload policies as maps, so the key attribute
// of each block determines set identity.
func TestClusterMonitoringConfigurationSetIdentity(t *testing.T) {
	testCases := map[string]struct {
		configuration map[string]any
		child         string
		attribute     string
		key           string
		wantKeys      []string
	}{
		"log types": {
			configuration: map[string]any{
				"cloud_watch_log_configuration": []any{map[string]any{
					names.AttrEnabled: true,
					"log_types": []any{
						map[string]any{names.AttrName: "STEP_LOGS", names.AttrValues: []any{"STDOUT"}},
						map[string]any{names.AttrName: "STEP_LOGS", names.AttrValues: []any{"STDERR"}},
						map[string]any{names.AttrName: "SPARK_DRIVER", names.AttrValues: []any{"STDOUT"}},
					},
				}},
			},
			child:     "cloud_watch_log_configuration",
			attribute: "log_types",
			key:       names.AttrName,
			wantKeys:  []string{"SPARK_DRIVER", "STEP_LOGS"},
		},
		"upload policies": {
			configuration: map[string]any{
				"s3_logging_configuration": []any{map[string]any{
					"log_type_upload_policy": []any{
						map[string]any{"log_type": "system-logs", "upload_policy": "emr-managed"},
						map[string]any{"log_type": "system-logs", "upload_policy": "disabled"},
						map[string]any{"log_type": "application-logs", "upload_policy": "disabled"},
					},
				}},
			},
			child:     "s3_logging_configuration",
			attribute: "log_type_upload_policy",
			key:       "log_type",
			wantKeys:  []string{"application-logs", "system-logs"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, monitoringConfigurationSchema(), map[string]any{
				"monitoring_configuration": []any{testCase.configuration},
			})
			monitoring := data.Get("monitoring_configuration").([]any)[0].(map[string]any)
			set := monitoring[testCase.child].([]any)[0].(map[string]any)[testCase.attribute].(*schema.Set)

			got := make([]string, 0, set.Len())
			for _, raw := range set.List() {
				got = append(got, raw.(map[string]any)[testCase.key].(string))
			}
			slices.Sort(got)

			if !slices.Equal(got, testCase.wantKeys) {
				t.Errorf("expected identities %v, got %v", testCase.wantKeys, got)
			}
		})
	}
}
