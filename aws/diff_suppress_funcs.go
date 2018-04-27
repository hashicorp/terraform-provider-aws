package aws

import (
	"bytes"
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jen20/awspolicyequivalence"
)

func suppressEquivalentAwsPolicyDiffs(k, old, new string, d *schema.ResourceData) bool {
	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)
	if err != nil {
		return false
	}

	return equivalent
}

// Suppresses minor version changes to the db_instance engine_version attribute
func suppressAwsDbEngineVersionDiffs(k, old, new string, d *schema.ResourceData) bool {
	// First check if the old/new values are nil.
	// If both are nil, we have no state to compare the values with, so register a diff.
	// This populates the attribute field during a plan/apply with fresh state, allowing
	// the attribute to still be used in future resources.
	// See https://github.com/hashicorp/terraform/issues/11881
	if old == "" && new == "" {
		return false
	}

	if v, ok := d.GetOk("auto_minor_version_upgrade"); ok {
		if v.(bool) {
			// If we're set to auto upgrade minor versions
			// ignore a minor version diff between versions
			if strings.HasPrefix(old, new) {
				log.Printf("[DEBUG] Ignoring minor version diff")
				return true
			}
		}
	}

	// Throw a diff by default
	return false
}

func suppressEquivalentJsonDiffs(k, old, new string, d *schema.ResourceData) bool {
	ob := bytes.NewBufferString("")
	if err := json.Compact(ob, []byte(old)); err != nil {
		return false
	}

	nb := bytes.NewBufferString("")
	if err := json.Compact(nb, []byte(new)); err != nil {
		return false
	}

	return jsonBytesEqual(ob.Bytes(), nb.Bytes())
}

func suppressOpenIdURL(k, old, new string, d *schema.ResourceData) bool {
	oldUrl, err := url.Parse(old)
	if err != nil {
		return false
	}

	newUrl, err := url.Parse(new)
	if err != nil {
		return false
	}

	oldUrl.Scheme = "https"

	return oldUrl.String() == newUrl.String()
}

func suppressAutoscalingGroupAvailabilityZoneDiffs(k, old, new string, d *schema.ResourceData) bool {
	// If VPC zone identifiers are provided then there is no need to explicitly
	// specify availability zones.
	if _, ok := d.GetOk("vpc_zone_identifier"); ok {
		return true
	}

	return false
}

func suppressRoute53ZoneNameWithTrailingDot(k, old, new string, d *schema.ResourceData) bool {
	return strings.TrimSuffix(old, ".") == strings.TrimSuffix(new, ".")
}

func suppressEquivalentDynamodbTableTtlDiffs(k, old, new string, d *schema.ResourceData) bool {
	// if there are no changes then there's nothing to suppress
	if !d.HasChange("ttl") {
		return false
	}

	ttlOldInterface, ttlNewInterface := d.GetChange("ttl")
	ttlOldSet := ttlOldInterface.(*schema.Set)
	ttlNewSet := ttlNewInterface.(*schema.Set)

	// if the ttl block hasn't been added or removed there's nothing to suppress
	if ttlOldSet.Len() == ttlNewSet.Len() {
		return false
	}

	// if the ttl block has been added,
	// changes should be suppressed if the new block has `enabled => false`
	if ttlOldSet.Len() == 0 {
		enabled := ttlNewSet.List()[0].(map[string]interface{})["enabled"].(bool)
		if !enabled {
			return true
		}
	}

	// if the ttl block has been removed,
	// changes should be suppressed if the old block had `enabled => false`
	if ttlNewSet.Len() == 0 {
		enabled := ttlOldSet.List()[0].(map[string]interface{})["enabled"].(bool)
		if !enabled {
			return true
		}
	}
	return false
}
