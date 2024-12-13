// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testResourceTargetStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		names.AttrARN:  "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		names.AttrRule: "testrule",
		"target_id":    "testtargetid",
	}
}

func testResourceTargetStateDataV0EventBusName() map[string]interface{} {
	return map[string]interface{}{
		names.AttrARN:    "arn:aws:test:us-east-1:123456789012:test", //lintignore:AWSAT003,AWSAT005
		"event_bus_name": "testbus",
		names.AttrRule:   "testrule",
		"target_id":      "testtargetid",
	}
}

func testResourceTargetStateDataV1() map[string]interface{} {
	v0 := testResourceTargetStateDataV0()
	return map[string]interface{}{
		names.AttrARN:    v0[names.AttrARN],
		"event_bus_name": "default",
		names.AttrRule:   v0[names.AttrRule],
		"target_id":      v0["target_id"],
	}
}

func testResourceTargetStateDataV1EventBusName() map[string]interface{} {
	v0 := testResourceTargetStateDataV0EventBusName()
	return map[string]interface{}{
		names.AttrARN:    v0[names.AttrARN],
		"event_bus_name": v0["event_bus_name"],
		names.AttrRule:   v0[names.AttrRule],
		"target_id":      v0["target_id"],
	}
}

func TestTargetStateUpgradeV0(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	expected := testResourceTargetStateDataV1()
	actual, err := tfevents.TargetStateUpgradeV0(ctx, testResourceTargetStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}

func TestTargetStateUpgradeV0EventBusName(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	expected := testResourceTargetStateDataV1EventBusName()
	actual, err := tfevents.TargetStateUpgradeV0(ctx, testResourceTargetStateDataV0EventBusName(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
