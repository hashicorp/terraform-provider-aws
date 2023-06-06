package naming_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/tools/tfsdk2fw/naming"
)

func TestToCamelCase(t *testing.T) {
	testCases := []struct {
		TestName      string
		Value         string
		ExpectedValue string
	}{
		{
			TestName:      "empty string",
			Value:         "",
			ExpectedValue: "",
		},
		{
			TestName:      "whitespace string",
			Value:         "  ",
			ExpectedValue: "",
		},
		{
			TestName:      "single word",
			Value:         "description",
			ExpectedValue: "Description",
		},
		{
			TestName:      "multiple words",
			Value:         "health_check_config",
			ExpectedValue: "HealthCheckConfig",
		},
		{
			TestName:      "ID",
			Value:         "id",
			ExpectedValue: "ID",
		},
		{
			TestName:      "something ID",
			Value:         "something_id",
			ExpectedValue: "SomethingID",
		},
		{
			TestName:      "ARN",
			Value:         "arn",
			ExpectedValue: "ARN",
		},
		{
			TestName:      "something ARN",
			Value:         "something_arn",
			ExpectedValue: "SomethingARN",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := naming.ToCamelCase(testCase.Value)

			if got != testCase.ExpectedValue {
				t.Errorf("expected: %s, got: %s", testCase.ExpectedValue, got)
			}
		})
	}
}
