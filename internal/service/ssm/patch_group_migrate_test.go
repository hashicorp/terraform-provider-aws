// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func testResourcePatchGroupStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		names.AttrID:  "testgroup",
		"baseline_id": "pb-0c4e592064EXAMPLE",
		"patch_group": "testgroup",
	}
}

func testResourcePatchGroupStateDataV1() map[string]interface{} {
	v0 := testResourcePatchGroupStateDataV0()
	return map[string]interface{}{
		names.AttrID:  fmt.Sprintf("%s,%s", v0["patch_group"], v0["baseline_id"]),
		"baseline_id": v0["baseline_id"],
		"patch_group": v0["patch_group"],
	}
}

func TestPatchGroupStateUpgradeV0(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	expected := testResourcePatchGroupStateDataV1()
	actual, err := patchGroupStateUpgradeV0(ctx, testResourcePatchGroupStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
