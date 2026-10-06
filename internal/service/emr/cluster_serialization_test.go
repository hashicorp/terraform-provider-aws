// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
)

func TestSerializeAutoScalingPolicy(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"nil members": {input: `{}`},
		"all fields": {
			input: `{
				"Constraints":{"MaxCapacity":10,"MinCapacity":1},
				"Rules":[
					{
						"Action":{"Market":"ON_DEMAND","SimpleScalingPolicyConfiguration":{"AdjustmentType":"CHANGE_IN_CAPACITY","CoolDown":300,"ScalingAdjustment":2}},
						"Description":"scale \"up\" <&>\n","Name":"z-first",
						"Trigger":{"CloudWatchAlarmDefinition":{
							"ComparisonOperator":"GREATER_THAN","Dimensions":[{"Key":"z-first","Value":"one"},{"Key":"a-second","Value":"two"}],
							"EvaluationPeriods":3,"MetricName":"YARNMemoryAvailablePercentage","Namespace":"AWS/ElasticMapReduce",
							"Period":300,"Statistic":"AVERAGE","Threshold":12.5,"Unit":"PERCENT"
						}}
					},
					{"Action":{"Market":"SPOT","SimpleScalingPolicyConfiguration":{"AdjustmentType":"EXACT_CAPACITY","ScalingAdjustment":-2}},"Name":"a-second"}
				]
			}`,
		},
		"explicit zero values": {
			input: `{
				"Constraints":{"MaxCapacity":0,"MinCapacity":0},
				"Rules":[{
					"Action":{"SimpleScalingPolicyConfiguration":{"CoolDown":0,"ScalingAdjustment":0}},
					"Description":"","Name":"",
					"Trigger":{"CloudWatchAlarmDefinition":{"Dimensions":[{"Key":"","Value":""}],"EvaluationPeriods":0,"MetricName":"","Namespace":"","Period":0,"Threshold":0}}
				}]
			}`,
		},
		"empty collections and objects": {
			input: `{"Constraints":{},"Rules":[{},{"Action":{},"Trigger":{}},{"Action":{"SimpleScalingPolicyConfiguration":{}},"Trigger":{"CloudWatchAlarmDefinition":{}}},{"Trigger":{"CloudWatchAlarmDefinition":{"Dimensions":[]}}},{"Trigger":{"CloudWatchAlarmDefinition":{"Dimensions":[{},{"Key":"key"},{"Value":"value"}]}}}]}`,
		},
		"empty rules": {input: `{"Rules":[]}`},
		"null members omitted": {
			input: `{"Constraints":null,"Rules":null}`,
			want:  `{}`,
		},
		"nested null members omitted": {
			input: `{"Constraints":{"MaxCapacity":null,"MinCapacity":null},"Rules":[{"Action":null,"Description":null,"Name":null,"Trigger":null},{"Action":{"SimpleScalingPolicyConfiguration":null},"Trigger":{"CloudWatchAlarmDefinition":null}},{"Action":{"SimpleScalingPolicyConfiguration":{"CoolDown":null,"ScalingAdjustment":null}},"Trigger":{"CloudWatchAlarmDefinition":{"Dimensions":null,"EvaluationPeriods":null,"MetricName":null,"Namespace":null,"Period":null,"Threshold":null}}}]}`,
			want:  `{"Constraints":{},"Rules":[{},{"Action":{},"Trigger":{}},{"Action":{"SimpleScalingPolicyConfiguration":{}},"Trigger":{"CloudWatchAlarmDefinition":{}}}]}`,
		},
		"empty enums omitted": {
			input: `{"Rules":[{"Action":{"Market":"","SimpleScalingPolicyConfiguration":{"AdjustmentType":""}},"Trigger":{"CloudWatchAlarmDefinition":{"ComparisonOperator":"","Statistic":"","Unit":""}}}]}`,
			want:  `{"Rules":[{"Action":{"SimpleScalingPolicyConfiguration":{}},"Trigger":{"CloudWatchAlarmDefinition":{}}}]}`,
		},
		"unknown enums retained": {
			input: `{"Rules":[{"Action":{"Market":"FUTURE","SimpleScalingPolicyConfiguration":{"AdjustmentType":"FUTURE"}},"Trigger":{"CloudWatchAlarmDefinition":{"ComparisonOperator":"FUTURE","Statistic":"FUTURE","Unit":"FUTURE"}}}]}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input awstypes.AutoScalingPolicy
			if err := tfjson.DecodeFromString(tc.input, &input); err != nil {
				t.Fatal(err)
			}
			encoder := smithyjson.NewEncoder()
			if err := tfemr.SerializeAutoScalingPolicy(&input, encoder.Value); err != nil {
				t.Fatal(err)
			}
			got := encoder.String()
			want := tc.want
			if want == "" {
				want = tc.input
			}
			if !json.Valid([]byte(got)) {
				t.Fatalf("invalid JSON: %s", got)
			}
			if diff := jsoncmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected JSON (+got, -want): %s", diff)
			}
		})
	}
}

func TestSerializeAutoScalingPolicyThreshold(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input float64
		want  string
	}{
		"NaN":               {input: math.NaN(), want: `"NaN"`},
		"positive infinity": {input: math.Inf(1), want: `"Infinity"`},
		"negative infinity": {input: math.Inf(-1), want: `"-Infinity"`},
		"negative zero":     {input: math.Copysign(0, -1), want: `-0`},
		"smallest positive": {input: math.SmallestNonzeroFloat64, want: `5e-324`},
		"largest positive":  {input: math.MaxFloat64, want: `1.7976931348623157e+308`},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			input := awstypes.AutoScalingPolicy{
				Rules: []awstypes.ScalingRule{{
					Trigger: &awstypes.ScalingTrigger{
						CloudWatchAlarmDefinition: &awstypes.CloudWatchAlarmDefinition{
							Threshold: aws.Float64(tc.input),
						},
					},
				}},
			}
			encoder := smithyjson.NewEncoder()
			if err := tfemr.SerializeAutoScalingPolicy(&input, encoder.Value); err != nil {
				t.Fatal(err)
			}
			got := encoder.String()
			want := `{"Rules":[{"Trigger":{"CloudWatchAlarmDefinition":{"Threshold":` + tc.want + `}}}]}`
			if !json.Valid([]byte(got)) {
				t.Fatalf("invalid JSON: %s", got)
			}
			if diff := jsoncmp.Diff(want, got); diff != "" {
				t.Errorf("unexpected JSON (+got, -want): %s", diff)
			}
			// JSON comparisons do not distinguish negative zero or numeric formatting.
			if got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

func TestSerializeConfigurations(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"nil":         {input: `null`, want: `[]`},
		"empty":       {input: `[]`},
		"nil members": {input: `[{}]`},
		"null members omitted": {
			input: `[{"Classification":null,"Configurations":null,"Properties":null}]`,
			want:  `[{}]`,
		},
		"empty values retained": {
			input: `[{"Classification":"","Configurations":[],"Properties":{}},{"Properties":{"":""}}]`,
		},
		"recursive configurations": {
			input: `[
				{
					"Classification":"z-first",
					"Configurations":[
						{"Classification":"z-child","Configurations":[{},{"Classification":"leaf","Configurations":[],"Properties":{"key":"value"}}],"Properties":{}},
						{"Classification":"a-child","Properties":{"MixedCase.Key":"<&>","escaped\"key":"line\nbreak","empty":""}}
					],
					"Properties":{"top-level":"value"}
				},
				{"Classification":"a-second","Configurations":[{"Configurations":[{"Configurations":[{}]}]}]}
			]`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input []awstypes.Configuration
			if err := tfjson.DecodeFromString(tc.input, &input); err != nil {
				t.Fatal(err)
			}
			encoder := smithyjson.NewEncoder()
			if err := tfemr.SerializeConfigurations(input, encoder.Value); err != nil {
				t.Fatal(err)
			}
			got := encoder.String()
			want := tc.want
			if want == "" {
				want = tc.input
			}
			if !json.Valid([]byte(got)) {
				t.Fatalf("invalid JSON: %s", got)
			}
			// jsoncmp compares object roots; wrapping preserves array order checks.
			if diff := jsoncmp.Diff(`{"Configurations":`+want+`}`, `{"Configurations":`+got+`}`); diff != "" {
				t.Errorf("unexpected JSON (+got, -want): %s", diff)
			}
		})
	}
}
