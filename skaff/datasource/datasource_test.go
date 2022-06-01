package datasource

import (
	"testing"
)

func TestToSnakeName(t *testing.T) {
	testCases := []struct {
		TestName string
		Input    string
		Expected string
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
		},
		{
			TestName: "simple",
			Input:    "Cheese",
			Expected: "cheese",
		},
		{
			TestName: "two word",
			Input:    "CityLights",
			Expected: "city_lights",
		},
		{
			TestName: "three word",
			Input:    "OpenEndResource",
			Expected: "open_end_resource",
		},
		{
			TestName: "initialism",
			Input:    "DBInstance",
			Expected: "db_instance",
		},
		{
			TestName: "initialisms",
			Input:    "DBInstanceVPCEndpoint",
			Expected: "db_instance_vpc_endpoint",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := toSnakeCase(testCase.Input, "")

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}
