// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
)

func testOntapStorageVirtualMachineStateDataV0() map[string]interface{} {
	return map[string]interface{}{
		"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name": "MeArrugoDerrito",
	}
}

func testOntapStorageVirtualMachineStateDataV1() map[string]interface{} {
	v0 := testOntapStorageVirtualMachineStateDataV0()
	return map[string]interface{}{
		"active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name": v0["active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name"],
	}
}

func TestOntapStorageVirtualMachineStateUpgradeV0(t *testing.T) {
	ctx := acctest.Context(t)
	t.Parallel()

	expected := testOntapStorageVirtualMachineStateDataV1()
	actual, err := tffsx.ResourceOntapStorageVirtualMachineStateUpgradeV0(ctx, testOntapStorageVirtualMachineStateDataV0(), nil)

	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
