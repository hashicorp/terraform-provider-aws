// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func recordMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Route53 Record State v0; migrating to v1 then v2")
		v1InstanceState := migrateRecordStateV0toV1(is)
		return migrateRecordStateV1toV2(v1InstanceState)
	case 1:
		log.Println("[INFO] Found AWS Route53 Record State v1; migrating to v2")
		return migrateRecordStateV1toV2(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateRecordStateV0toV1(is *terraform.InstanceState) *terraform.InstanceState {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)
	newName := strings.TrimSuffix(is.Attributes[names.AttrName], ".")
	is.Attributes[names.AttrName] = newName
	log.Printf("[DEBUG] Attributes after migration: %#v, new name: %s", is.Attributes, newName)
	return is
}

func migrateRecordStateV1toV2(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}
	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)
	if is.Attributes[names.AttrWeight] != "" && is.Attributes[names.AttrWeight] != "-1" {
		is.Attributes["weighted_routing_policy.#"] = "1"
		key := "weighted_routing_policy.0.weight"
		is.Attributes[key] = is.Attributes[names.AttrWeight]
	}
	if is.Attributes["failover"] != "" {
		is.Attributes["failover_routing_policy.#"] = "1"
		key := "failover_routing_policy.0.type"
		is.Attributes[key] = is.Attributes["failover"]
	}
	delete(is.Attributes, names.AttrWeight)
	delete(is.Attributes, "failover")
	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
