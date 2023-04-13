package ssm_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfssm "github.com/hashicorp/terraform-provider-aws/internal/service/ssm"
)

func testResourcePatchGroupStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"id":          "testgroup",
		"baseline_id": "pb-0c4e592064EXAMPLE",
		"patch_group": "testgroup",
	}
}

func testResourcePatchGroupStateDataV1() map[string]interface{} {
	v0 := testResourcePatchGroupStateDataV0()
	return map[string]interface{}{
		"id":          fmt.Sprintf("%s,%s", v0["patch_group"], v0["baseline_id"]),
		"baseline_id": v0["baseline_id"],
		"patch_group": v0["patch_group"],
	}
}

func TestPatchGroupStateUpgradeV0(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	expected := testResourcePatchGroupStateDataV1()
	actual, err := tfssm.PatchGroupStateUpgradeV0(ctx, testResourcePatchGroupStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
