package redshift

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
)

func TestExpandParameters(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":  "character_set_client",
			"value": "utf8",
		},
	}
	parameters := expandParameters(expanded)

	expected := &redshift.Parameter{
		ParameterName:  aws.String("character_set_client"),
		ParameterValue: aws.String("utf8"),
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
		Input  []*redshift.Parameter
		Output []map[string]interface{}
	}{
		{
			Input: []*redshift.Parameter{
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
	}

	for _, tc := range cases {
		output := flattenParameters(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}
