// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type SlotResolutionSettingData struct {
	SlotResolutionStrategy fwtypes.StringEnum[awstypes.SlotResolutionStrategy] `tfsdk:"slot_resolution_strategy"`
}

func SlotResolutionSettingBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[SlotResolutionSettingData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"slot_resolution_strategy": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.SlotResolutionStrategy](),
					Required:   true,
				},
			},
		},
	}
}
