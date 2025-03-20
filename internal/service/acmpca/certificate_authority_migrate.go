// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acmpca

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceCertificateAuthorityMigrateState(v int, is *terraform.InstanceState, meta any) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found ACM PCA Certificate Authority state v0; migrating to v1")
		return migrateCertificateAuthorityStateV0toV1(is)
	default:
		return is, fmt.Errorf("unexpected schema version: %d", v)
	}
}

func migrateCertificateAuthorityStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty ACM PCA Certificate Authority state; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// Add permanent_deletion_time_in_days virtual attribute with Default
	is.Attributes["permanent_deletion_time_in_days"] = "30"

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)

	return is, nil
}
