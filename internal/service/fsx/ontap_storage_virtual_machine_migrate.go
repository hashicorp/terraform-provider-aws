package fsx

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceOntapStorageVirtualMachineMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found FSx Ontap Storage Virtual Machine state v0; migrating to v1")
		return migrateOntapStorageVirtualMachineStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateOntapStorageVirtualMachineStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty FSx Ontap Storage Virtual Machine state; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)
	is.Attributes["active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name"] = is.Attributes["active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name"]
	delete(is.Attributes, "active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguidshed_name")

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
