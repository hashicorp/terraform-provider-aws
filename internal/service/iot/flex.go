// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandThingTypeProperties(config map[string]interface{}) *iot.ThingTypeProperties {
	properties := &iot.ThingTypeProperties{
		SearchableAttributes: flex.ExpandStringSet(config["searchable_attributes"].(*schema.Set)),
	}

	if v, ok := config[names.AttrDescription]; ok && v.(string) != "" {
		properties.ThingTypeDescription = aws.String(v.(string))
	}

	return properties
}

func flattenThingTypeProperties(s *iot.ThingTypeProperties) []map[string]interface{} {
	m := map[string]interface{}{
		names.AttrDescription:   "",
		"searchable_attributes": flex.FlattenStringSet(nil),
	}

	if s == nil {
		return []map[string]interface{}{m}
	}

	m[names.AttrDescription] = aws.StringValue(s.ThingTypeDescription)
	m["searchable_attributes"] = flex.FlattenStringSet(s.SearchableAttributes)

	return []map[string]interface{}{m}
}
