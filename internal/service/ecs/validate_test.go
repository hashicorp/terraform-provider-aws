// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidPlacementConstraint(t *testing.T) {
	t.Parallel()

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
			constExpr: names.AttrExpression,
			Err:       false,
		},
		{
			constType: "memberOf",
			constExpr: names.AttrExpression,
			Err:       false,
		},
	}

	for _, tc := range cases {
		if err := validPlacementConstraint(tc.constType, tc.constExpr); err != nil && !tc.Err {
			t.Fatalf("Unexpected validation error for \"%s:%s\": %s",
				tc.constType, tc.constExpr, err)
		}
	}
}

func TestValidPlacementStrategy(t *testing.T) {
	t.Parallel()

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
		if err := validPlacementStrategy(tc.stratType, tc.stratField); err != nil && !tc.Err {
			t.Fatalf("Unexpected validation error for \"%s:%s\": %s",
				tc.stratType, tc.stratField, err)
		}
	}
}
