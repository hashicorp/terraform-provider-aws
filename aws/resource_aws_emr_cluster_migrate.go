package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/terraform"
)

func resourceAwsEMRClusterMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS EMR Cluster State v0; migrating to v1")
		return migrateEMRClusterStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateEMRClusterStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty EMR Cluster State; nothing to migrate.")
		return is, nil
	}

	ebsConfigCount := 0
	newEBSConfigKeys := make(map[string]string)
	ebsConfigMapHashToInt := make(map[int]int)
	keys := make([]string, len(is.Attributes))
	for k := range is.Attributes {
		keys = append(keys, k)
	}
	for _, k := range keys {
		if !strings.HasPrefix(k, "instance_group.") {
			continue
		}

		if !strings.Contains(k, "ebs_config") {
			continue
		}

		if strings.HasSuffix(k, "ebs_config.#") {
			continue
		}

		kParts := strings.Split(k, ".")
		if len(kParts) != 5 {
			return nil, fmt.Errorf("migration error: found `instance_group` key in unexpected format: %s", k)
		}
		instanceGroupHash, err := strconv.Atoi(kParts[1])
		if err != nil {
			return nil, fmt.Errorf("migration error: found `instance_group` hash in unexpected format: %s", kParts[1])
		}
		ebsConfigHash, err := strconv.Atoi(kParts[3])
		if err != nil {
			return nil, fmt.Errorf("migration error: found `ebs_config` hash in unexpected format: %s", kParts[3])
		}
		if _, ok := ebsConfigMapHashToInt[ebsConfigHash]; !ok {
			ebsConfigMapHashToInt[ebsConfigHash] = ebsConfigCount
			ebsConfigCount++
		}
		newKey := fmt.Sprintf("instance_group.%d.ebs_config.%d.%s", instanceGroupHash, ebsConfigMapHashToInt[ebsConfigHash], kParts[4])
		newEBSConfigKeys[newKey] = is.Attributes[k]
		delete(is.Attributes, k)
	}

	for k, v := range newEBSConfigKeys {
		is.Attributes[k] = v

	}

	log.Printf("[DEBUG] EMR Cluster Attributes after migration: %#v", is.Attributes)

	return is, nil
}

func updateV0ToV1EMRClusterEBSConfig(is *terraform.InstanceState) error {

	return nil
}
