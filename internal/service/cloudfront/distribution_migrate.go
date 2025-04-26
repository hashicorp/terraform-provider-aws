// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceDistributionMigrateState(v int, is *terraform.InstanceState, meta any) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found CloudFront Distribution state v0; migrating to v1")
		return migrateDistributionStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateDistributionStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty CloudFront Distribution state; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// Add wait_for_deployment virtual attribute with Default
	is.Attributes["wait_for_deployment"] = "true"

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)

	return is, nil
}
