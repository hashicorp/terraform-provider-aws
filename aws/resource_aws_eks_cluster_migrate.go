package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func resourceAwsEksClusterMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found EKS Cluster state v0; migrating to v1")
		return migrateEksClusterStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateEksClusterStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty EKS Cluster state; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)
	for k, v := range is.Attributes {
		if k != "enabled_cluster_log_types.#" && strings.HasPrefix(k, "enabled_cluster_log_types.") {
			delete(is.Attributes, k)
			hash := schema.HashString(v)
			is.Attributes[fmt.Sprintf("enabled_cluster_log_types.%d", hash)] = v
		}
	}

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
