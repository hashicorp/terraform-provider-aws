// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	dataschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Terraform Plugin Framework variants of tags schemas.

func TagsAttribute() schema.Attribute {
	return schema.MapAttribute{
		CustomType:  MapType,
		ElementType: types.StringType,
		Optional:    true,
	}
}

func TagsAttributeComputedOnly() dataschema.MapAttribute {
	return dataschema.MapAttribute{
		CustomType:  MapType,
		ElementType: types.StringType,
		Computed:    true,
	}
}

func TagsAttributeForceNew() schema.Attribute {
	return schema.MapAttribute{
		CustomType:  MapType,
		ElementType: types.StringType,
		Optional:    true,
		PlanModifiers: []planmodifier.Map{
			mapplanmodifier.RequiresReplace(),
		},
	}
}

func TagsAttributeRequired() schema.Attribute {
	return schema.MapAttribute{
		CustomType:  MapType,
		ElementType: types.StringType,
		Required:    true,
	}
}

var (
	Unknown = types.MapUnknown(types.StringType)
)

var (
	Null = NewMapValueNull()
)
