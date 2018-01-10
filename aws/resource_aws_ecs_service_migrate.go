package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/terraform"
)

func resourceAwsEcsServiceMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found Aws Ecs Service State v0; migrating to v1")
		return migrateAwsEcsServiceStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateAwsEcsServiceStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	type placementStrategy struct {
		hash  string
		ttype string
		field string
	}
	appendPSWithoutDuplicate := func(slice []*placementStrategy, e *placementStrategy) []*placementStrategy {
		results := make([]*placementStrategy, 0, len(slice)+1)
		encountered := map[*placementStrategy]bool{}
		for i := 0; i < len(slice); i++ {
			if !encountered[slice[i]] {
				encountered[slice[i]] = true
				results = append(results, slice[i])
			}
		}
		if !encountered[e] {
			results = append(results, e)
		}
		return results
	}
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}
	newPlacementStrategies := make([]*placementStrategy, 0)
	numCountStr, ok := is.Attributes["placement_strategy.#"]
	if !ok || numCountStr == "0" {
		log.Println("[DEBUG] Empty placement_strategy in InstanceState; no need to migrate.")
		return is, nil
	}
	numCount, _ := strconv.Atoi(numCountStr)
	for k, v := range is.Attributes {
		if !strings.HasPrefix(k, "placement_strategy.") || strings.HasPrefix(k, "placement_strategy.#") {
			continue
		}
		path := strings.Split(k, ".")
		if len(path) != 3 {
			return is, fmt.Errorf("Found unexpected placement_strategy field: %#v", k)
		}
		hashcode, attr := path[1], path[2]
		ps := &placementStrategy{hash: hashcode}
		for _, vv := range newPlacementStrategies {
			if vv.hash == hashcode {
				ps = vv
			}
		}
		if attr == "field" {
			ps.field = v
		}
		if attr == "type" {
			ps.ttype = v
		}
		newPlacementStrategies = appendPSWithoutDuplicate(newPlacementStrategies, ps)
		delete(is.Attributes, k)
	}
	if len(newPlacementStrategies) != numCount {
		return is, fmt.Errorf("Num of placement_strategy slice should be %d, but %d", numCount, len(newPlacementStrategies))
	}
	for idx, v := range newPlacementStrategies {
		if v.field != "" {
			is.Attributes[fmt.Sprintf("placement_strategy.%d.field", idx)] = v.field
		}
		if v.ttype != "" {
			is.Attributes[fmt.Sprintf("placement_strategy.%d.type", idx)] = v.ttype
		}
	}
	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
