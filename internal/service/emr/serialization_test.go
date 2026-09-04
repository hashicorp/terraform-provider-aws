// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestFlattenAutoScalingPolicyDescription(t *testing.T) {
	t.Parallel()

	const want = `
{
  "Constraints": {
    "MaxCapacity": 10,
    "MinCapacity": 1
  },
  "Rules": [
    {
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "CoolDown": 60,
          "ScalingAdjustment": 2
        }
      },
      "Description": "Scale out when utilization is high",
      "Name": "scale-out",
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "GREATER_THAN",
          "Dimensions": [
            {"Key": "ClusterId", "Value": "cluster-123"}
          ],
          "EvaluationPeriods": 2,
          "MetricName": "CPUUtilization",
          "Namespace": "AWS/EMR",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 75,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}`

	var policy awstypes.AutoScalingPolicy
	if err := tfjson.DecodeFromString(want, &policy); err != nil {
		t.Fatal(err)
	}

	got, err := flattenAutoScalingPolicyDescription(&awstypes.AutoScalingPolicyDescription{
		Constraints: policy.Constraints,
		Rules:       policy.Rules,
	})
	if err != nil {
		t.Fatal(err)
	}

	assertSerializationJSONEqual(t, want, got)
}

func TestFlattenConfigurationJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input string
		want  string
	}{
		"empty": {
			input: `[]`,
			want:  `[]`,
		},
		"nested configurations and properties": {
			input: `
[
  {
    "Classification": "spark-defaults",
    "Properties": {
      "spark.executor.memory": "4g",
      "spark.executor.cores": "2"
    },
    "Configurations": [
      {
        "Classification": "spark-env",
        "Configurations": [
          {
            "Classification": "export",
            "Properties": {"JAVA_HOME": "/usr/lib/jvm"}
          }
        ],
        "Properties": {"PYSPARK_PYTHON": "python3"}
      }
    ]
  }
]`,
			want: `
[
  {
    "Classification": "spark-defaults",
    "Properties": {
      "spark.executor.memory": "4g",
      "spark.executor.cores": "2"
    },
    "Configurations": [
      {
        "Classification": "spark-env",
        "Configurations": [
          {
            "Classification": "export",
            "Properties": {"JAVA_HOME": "/usr/lib/jvm"}
          }
        ],
        "Properties": {"PYSPARK_PYTHON": "python3"}
      }
    ]
  }
]`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var input []awstypes.Configuration
			if err := tfjson.DecodeFromString(testCase.input, &input); err != nil {
				t.Fatal(err)
			}

			got, err := flattenConfigurationJSON(input)
			if err != nil {
				t.Fatal(err)
			}

			assertSerializationJSONEqual(t, testCase.want, got)
		})
	}
}

func assertSerializationJSONEqual(t *testing.T, want, got string) {
	t.Helper()

	if diff := jsoncmp.Diff(`{"value":`+want+`}`, `{"value":`+got+`}`); diff != "" {
		t.Errorf("unexpected JSON diff (+want, -got):\n%s", diff)
	}
}
