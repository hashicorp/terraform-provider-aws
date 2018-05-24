package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func resourceAwsCodebuildMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Codebuild State v0; migrating to v1")
		return migrateCodebuildStateV0toV1(is)
	case 1:
		log.Println("[INFO] Found AWS Codebuild State v1; migrating to v2")
		return migrateCodebuildStateV1toV2(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateCodebuildStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	if is.Attributes["timeout"] != "" {
		is.Attributes["build_timeout"] = strings.TrimSpace(is.Attributes["timeout"])
	}

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}

func migrateCodebuildStateV1toV2(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration v1: %#v", is.Attributes)

	prefix := "source"
	entity := resourceAwsCodeBuildProject()

	// Read old keys
	reader := &schema.MapFieldReader{
		Schema: entity.Schema,
		Map:    schema.BasicMapReader(is.Attributes),
	}
	result, err := reader.ReadField([]string{prefix})
	if err != nil {
		return nil, err
	}

	oldKeys, ok := result.Value.(*schema.Set)
	if !ok {
		return nil, fmt.Errorf("Got unexpected value from state: %#v", result.Value)
	}

	// Delete old keys
	for k := range is.Attributes {
		if strings.HasPrefix(k, fmt.Sprintf("%s.", prefix)) {
			delete(is.Attributes, k)
		}
	}

	// We just need the defaults for the new keys, no custom values needed

	// Write new keys
	writer := schema.MapFieldWriter{
		Schema: entity.Schema,
	}
	if err := writer.WriteField([]string{prefix}, oldKeys); err != nil {
		return is, err
	}
	for k, v := range writer.Map() {
		is.Attributes[k] = v
	}

	log.Printf("[DEBUG] Attributes after migration v2: %#v", is.Attributes)
	return is, nil
}
