package rds

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestExpandParameters(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":         "character_set_client",
			"value":        "utf8",
			"apply_method": "immediate",
		},
	}
	parameters := expandParameters(expanded)

	expected := &rds.Parameter{
		ParameterName:  aws.String("character_set_client"),
		ParameterValue: aws.String("utf8"),
		ApplyMethod:    aws.String("immediate"),
	}

	if !reflect.DeepEqual(parameters[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			parameters[0],
			expected)
	}
}

func TestFlattenParameters(t *testing.T) {
	cases := []struct {
		Input  []*rds.Parameter
		Output []map[string]interface{}
	}{
		{
			Input: []*rds.Parameter{
				{
					ParameterName:  aws.String("character_set_client"),
					ParameterValue: aws.String("utf8"),
				},
			},
			Output: []map[string]interface{}{
				{
					"name":  "character_set_client",
					"value": "utf8",
				},
			},
		},
		{
			Input: []*rds.Parameter{
				{
					ParameterName:  aws.String("character_set_client"),
					ParameterValue: aws.String("utf8"),
					ApplyMethod:    aws.String("immediate"),
				},
			},
			Output: []map[string]interface{}{
				{
					"name":         "character_set_client",
					"value":        "utf8",
					"apply_method": "immediate",
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenParameters(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

// TestExpandRdsClusterScalingConfiguration_serverless removed in v3.0.0
// as all engine_modes are treated equal when expanding scaling_configuration
// and an override of min_capacity is no longer needed
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/11698
func TestExpandClusterScalingConfiguration_basic(t *testing.T) {
	type testCase struct {
		EngineMode string
		Input      []interface{}
		ExpectNil  bool
	}
	cases := []testCase{}

	// RDS Cluster Scaling Configuration is only valid for serverless, but we're relying on AWS errors.
	// If Terraform adds whole-resource validation, we can do our own validation at plan time.
	for _, engineMode := range []string{"global", "multimaster", "parallelquery", "provisioned", "serverless"} {
		cases = append(cases, []testCase{
			{
				EngineMode: engineMode,
				Input: []interface{}{
					map[string]interface{}{
						"auto_pause":               false,
						"max_capacity":             32,
						"min_capacity":             4,
						"seconds_until_auto_pause": 600,
						"timeout_action":           "ForceApplyCapacityChange",
					},
				},
				ExpectNil: false,
			},
			{
				EngineMode: engineMode,
				Input:      []interface{}{},
				ExpectNil:  true,
			}, {
				EngineMode: engineMode,
				Input: []interface{}{
					nil,
				},
				ExpectNil: true,
			},
		}...)
	}

	for _, tc := range cases {
		output := expandClusterScalingConfiguration(tc.Input)
		if tc.ExpectNil != (output == nil) {
			t.Errorf("EngineMode %q: Expected nil: %t, Got: %v", tc.EngineMode, tc.ExpectNil, output)
		}
	}
}
