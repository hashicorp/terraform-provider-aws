// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func keyPairMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Key Pair State v0; migrating to v1")
		return migrateKeyPairStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateKeyPairStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// replace public_key with a stripped version, removing `\n` from the end
	// see https://github.com/hashicorp/terraform/issues/3455
	is.Attributes[names.AttrPublicKey] = strings.TrimSpace(is.Attributes[names.AttrPublicKey])

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
