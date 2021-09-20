package ecs

import (
	"testing"
)

func TestValidPlacementConstraint(t *testing.T) {
	cases := []struct {
		constType string
		constExpr string
		Err       bool
	}{
		{
			constType: "distinctInstance",
			constExpr: "",
			Err:       false,
		},
		{
			constType: "memberOf",
			constExpr: "",
			Err:       true,
		},
		{
			constType: "distinctInstance",
			constExpr: "expression",
			Err:       false,
		},
		{
			constType: "memberOf",
			constExpr: "expression",
			Err:       false,
		},
	}

	for _, tc := range cases {
		if err := validateAwsECSPlacementConstraint(tc.constType, tc.constExpr); err != nil && !tc.Err {
			t.Fatalf("Unexpected validation error for \"%s:%s\": %s",
				tc.constType, tc.constExpr, err)
		}

	}
}

func TestValidPlacementStrategy(t *testing.T) {
	cases := []struct {
		stratType  string
		stratField string
		Err        bool
	}{
		{
			stratType:  "random",
			stratField: "",
			Err:        false,
		},
		{
			stratType:  "spread",
			stratField: "instanceID",
			Err:        false,
		},
		{
			stratType:  "binpack",
			stratField: "cpu",
			Err:        false,
		},
		{
			stratType:  "binpack",
			stratField: "memory",
			Err:        false,
		},
		{
			stratType:  "binpack",
			stratField: "disk",
			Err:        true,
		},
		{
			stratType:  "fakeType",
			stratField: "",
			Err:        true,
		},
	}

	for _, tc := range cases {
		if err := validateAwsECSPlacementStrategy(tc.stratType, tc.stratField); err != nil && !tc.Err {
			t.Fatalf("Unexpected validation error for \"%s:%s\": %s",
				tc.stratType, tc.stratField, err)
		}
	}
}
