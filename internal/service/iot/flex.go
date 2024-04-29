// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandThingTypeProperties(config map[string]interface{}) *awstypes.ThingTypeProperties {
	properties := &awstypes.ThingTypeProperties{
		SearchableAttributes: flex.ExpandStringValueSet(config["searchable_attributes"].(*schema.Set)),
	}

	if v, ok := config["description"]; ok && v.(string) != "" {
		properties.ThingTypeDescription = aws.String(v.(string))
	}

	return properties
}

func flattenThingTypeProperties(s *awstypes.ThingTypeProperties) []map[string]interface{} {
	m := map[string]interface{}{
		"description":           "",
		"searchable_attributes": flex.FlattenStringSet(nil),
	}

	if s == nil {
		return []map[string]interface{}{m}
	}

	m["description"] = aws.ToString(s.ThingTypeDescription)
	m["searchable_attributes"] = flex.FlattenStringValueSet(s.SearchableAttributes)

	return []map[string]interface{}{m}
}
