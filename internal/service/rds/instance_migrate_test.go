// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestInstanceStateUpgradeV0(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	testCases := []struct {
		Description   string
		InputState    map[string]interface{}
		ExpectedState map[string]interface{}
	}{
		{
			Description:   "missing state",
			InputState:    nil,
			ExpectedState: nil,
		},
		{
			Description: "adds delete_automated_backups",
			InputState: map[string]interface{}{
				names.AttrAllocatedStorage: 10,
				names.AttrEngine:           "mariadb",
				names.AttrIdentifier:       "my-test-instance",
				"instance_class":           "db.t2.micro",
				names.AttrPassword:         "avoid-plaintext-passwords",
				names.AttrUsername:         "tfacctest",
				names.AttrTags:             map[string]interface{}{acctest.CtKey1: acctest.CtValue1},
			},
			ExpectedState: map[string]interface{}{
				names.AttrAllocatedStorage: 10,
				"delete_automated_backups": true,
				names.AttrEngine:           "mariadb",
				names.AttrIdentifier:       "my-test-instance",
				"instance_class":           "db.t2.micro",
				names.AttrPassword:         "avoid-plaintext-passwords",
				names.AttrUsername:         "tfacctest",
				names.AttrTags:             map[string]interface{}{acctest.CtKey1: acctest.CtValue1},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			t.Parallel()

			got, err := tfrds.InstanceStateUpgradeV0(ctx, testCase.InputState, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			if !reflect.DeepEqual(testCase.ExpectedState, got) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", testCase.ExpectedState, got)
			}
		})
	}
}

func TestInstanceStateUpgradeV1(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	testCases := []struct {
		Description   string
		InputState    map[string]interface{}
		ExpectedState map[string]interface{}
	}{
		{
			Description:   "missing state",
			InputState:    nil,
			ExpectedState: nil,
		},
		{
			Description: "change id to resource id",
			InputState: map[string]interface{}{
				names.AttrAllocatedStorage: 10,
				names.AttrEngine:           "mariadb",
				names.AttrID:               "my-test-instance",
				names.AttrIdentifier:       "my-test-instance",
				"instance_class":           "db.t2.micro",
				names.AttrPassword:         "avoid-plaintext-passwords",
				names.AttrResourceID:       "db-cnuap2ilnbmok4eunzklfvwjca",
				names.AttrTags:             map[string]interface{}{acctest.CtKey1: acctest.CtValue1},
				names.AttrUsername:         "tfacctest",
			},
			ExpectedState: map[string]interface{}{
				names.AttrAllocatedStorage: 10,
				names.AttrEngine:           "mariadb",
				names.AttrID:               "db-cnuap2ilnbmok4eunzklfvwjca",
				names.AttrIdentifier:       "my-test-instance",
				"instance_class":           "db.t2.micro",
				names.AttrPassword:         "avoid-plaintext-passwords",
				names.AttrResourceID:       "db-cnuap2ilnbmok4eunzklfvwjca",
				names.AttrTags:             map[string]interface{}{acctest.CtKey1: acctest.CtValue1},
				names.AttrUsername:         "tfacctest",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			t.Parallel()

			got, err := tfrds.InstanceStateUpgradeV1(ctx, testCase.InputState, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			if !reflect.DeepEqual(testCase.ExpectedState, got) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", testCase.ExpectedState, got)
			}
		})
	}
}
