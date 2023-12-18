// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func ObfuscationSettingBlock(ctx context.Context) schema.ListNestedBlock {
	return schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[ObfuscationSettingData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"obfuscation_setting_type": schema.StringAttribute{
					CustomType: fwtypes.StringEnumType[awstypes.ObfuscationSettingType](),
					Required:   true,
				},
			},
		},
	}
}

type ObfuscationSettingData struct {
	ObfuscationSettingType fwtypes.StringEnum[awstypes.ObfuscationSettingType] `tfsdk:"obfuscation_setting_type"`
}
