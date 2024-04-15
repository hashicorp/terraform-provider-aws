// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

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
			got := ToSnakeCase(testCase.Input, "")

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestToHumanResName(t *testing.T) {
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
			Expected: "Cheese",
		},
		{
			TestName: "two word",
			Input:    "CityLights",
			Expected: "City Lights",
		},
		{
			TestName: "three word",
			Input:    "OpenEndResource",
			Expected: "Open End Resource",
		},
		{
			TestName: "initialism",
			Input:    "DBInstance",
			Expected: "DB Instance",
		},
		{
			TestName: "initialisms",
			Input:    "DBInstanceVPCEndpoint",
			Expected: "DB Instance VPC Endpoint",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			got := ToHumanResName(testCase.Input)

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}

func TestToLowercasePrefix(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"title", "HelloWorld", "helloWorld"},
		{"multi-upper prefix", "ARNParse", "arnParse"},
		{"all upper", "FOO", "foo"},
		{"all lower", "bar", "bar"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLowercasePrefix(tt.s); got != tt.want {
				t.Errorf("ToLowercasePrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
