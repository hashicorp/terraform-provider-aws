package aws

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func testresourceAwsSsmPatchGroupStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"id":          "testgroup",
		"baseline_id": "pb-0c4e592064EXAMPLE",
		"patch_group": "testgroup",
	}
}

func testresourceAwsSsmPatchGroupStateDataV1() map[string]interface{} {
	v0 := testresourceAwsSsmPatchGroupStateDataV0()
	return map[string]interface{}{
		"id":          fmt.Sprintf("%s,%s", v0["patch_group"], v0["baseline_id"]),
		"baseline_id": v0["baseline_id"],
		"patch_group": v0["patch_group"],
	}
}

func TestResourceAWSSSMPatchGroupStateUpgradeV0(t *testing.T) {
	expected := testresourceAwsSsmPatchGroupStateDataV1()
	actual, err := resourceAwsSsmPatchGroupStateUpgradeV0(context.Background(), testresourceAwsSsmPatchGroupStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
