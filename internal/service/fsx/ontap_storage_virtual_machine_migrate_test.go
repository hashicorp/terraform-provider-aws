// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
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
	ctx := context.Background() // Don't use acctest.Context as it leads to an import cycle.
	t.Parallel()

	want := testOntapStorageVirtualMachineStateDataV1()
	got, err := resourceONTAPStorageVirtualMachineStateUpgradeV0(ctx, testOntapStorageVirtualMachineStateDataV0(), nil)

	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}
