// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package paymentcryptography

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/paymentcryptography/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func keySchemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"deletion_window_in_days": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(defaultDeletionWindowInDays),
			},
			names.AttrEnabled: schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"exportable": schema.BoolAttribute{
				Required: true,
			},
			"key_check_value": schema.StringAttribute{
				Computed: true,
			},
			"key_check_value_algorithm": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.KeyCheckValueAlgorithm](),
				Optional:   true,
				Computed:   true,
			},
			"key_origin": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.KeyOrigin](),
			},
			"key_state": schema.StringAttribute{
				Computed:   true,
				CustomType: fwtypes.StringEnumType[awstypes.KeyState](),
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"key_attributes": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				CustomType: fwtypes.NewObjectTypeOf[keyAttributesModelV0](ctx),
				Attributes: map[string]schema.Attribute{
					"key_algorithm": schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.StringEnumType[awstypes.KeyAlgorithm](),
					},
					"key_class": schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.StringEnumType[awstypes.KeyClass](),
					},
					"key_usage": schema.StringAttribute{
						Required:   true,
						CustomType: fwtypes.StringEnumType[awstypes.KeyUsage](),
					},
				},
				Blocks: map[string]schema.Block{
					"key_modes_of_use": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType: fwtypes.NewObjectTypeOf[keyModesOfUseModel](ctx),
						Attributes: map[string]schema.Attribute{
							"decrypt": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"derive_key": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"encrypt": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"generate": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"no_restrictions": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"sign": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"unwrap": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"verify": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
							"wrap": schema.BoolAttribute{
								Optional: true,
								Computed: true,
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

type resourceKeyModelV0 struct {
	KeyArn                 types.String                                        `tfsdk:"arn"`
	DeletionWindowInDays   types.Int64                                         `tfsdk:"deletion_window_in_days"`
	Enabled                types.Bool                                          `tfsdk:"enabled"`
	Exportable             types.Bool                                          `tfsdk:"exportable"`
	ID                     types.String                                        `tfsdk:"id"`
	KeyAttributes          fwtypes.ObjectValueOf[keyAttributesModelV0]         `tfsdk:"key_attributes"`
	KeyCheckValue          types.String                                        `tfsdk:"key_check_value"`
	KeyCheckValueAlgorithm fwtypes.StringEnum[awstypes.KeyCheckValueAlgorithm] `tfsdk:"key_check_value_algorithm"`
	KeyOrigin              fwtypes.StringEnum[awstypes.KeyOrigin]              `tfsdk:"key_origin"`
	KeyState               fwtypes.StringEnum[awstypes.KeyState]               `tfsdk:"key_state"`
	Tags                   tftags.Map                                          `tfsdk:"tags"`
	TagsAll                tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                      `tfsdk:"timeouts"`
}

type keyAttributesModelV0 struct {
	KeyAlgorithm  fwtypes.StringEnum[awstypes.KeyAlgorithm] `tfsdk:"key_algorithm"`
	KeyClass      fwtypes.StringEnum[awstypes.KeyClass]     `tfsdk:"key_class"`
	KeyModesOfUse fwtypes.ObjectValueOf[keyModesOfUseModel] `tfsdk:"key_modes_of_use"`
	KeyUsage      fwtypes.StringEnum[awstypes.KeyUsage]     `tfsdk:"key_usage"`
}

func upgradeKeyStateV0toV1(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var keyDataV0 resourceKeyModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &keyDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	keyDataV1 := keyResourceModel{
		KeyARN:                 keyDataV0.KeyArn,
		DeletionWindowInDays:   keyDataV0.DeletionWindowInDays,
		Enabled:                keyDataV0.Enabled,
		Exportable:             keyDataV0.Exportable,
		ID:                     keyDataV0.ID,
		KeyAttributes:          upgradeKeyAttributesStateFromV0(ctx, keyDataV0.KeyAttributes, &response.Diagnostics),
		KeyCheckValue:          keyDataV0.KeyCheckValue,
		KeyCheckValueAlgorithm: keyDataV0.KeyCheckValueAlgorithm,
		KeyOrigin:              keyDataV0.KeyOrigin,
		KeyState:               keyDataV0.KeyState,
		Tags:                   keyDataV0.Tags,
		TagsAll:                keyDataV0.TagsAll,
		Timeouts:               keyDataV0.Timeouts,
	}

	response.Diagnostics.Append(response.State.Set(ctx, keyDataV1)...)
}

func upgradeKeyAttributesStateFromV0(ctx context.Context, old fwtypes.ObjectValueOf[keyAttributesModelV0], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[keyAttributesModel] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[keyAttributesModel](ctx)
	}

	var oldObj keyAttributesModelV0
	diags.Append(old.As(ctx, &oldObj, basetypes.ObjectAsOptions{})...)

	newObj := []keyAttributesModel{
		{
			KeyAlgorithm:  oldObj.KeyAlgorithm,
			KeyClass:      oldObj.KeyClass,
			KeyModesOfUse: upgradeKeyModesOfUseModelStateFromV0(ctx, oldObj.KeyModesOfUse, diags),
			KeyUsage:      oldObj.KeyUsage,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newObj)
	diags.Append(d...)

	return result
}

func upgradeKeyModesOfUseModelStateFromV0(ctx context.Context, old fwtypes.ObjectValueOf[keyModesOfUseModel], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[keyModesOfUseModel] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[keyModesOfUseModel](ctx)
	}

	var oldObj keyModesOfUseModel
	diags.Append(old.As(ctx, &oldObj, basetypes.ObjectAsOptions{})...)

	newObj := []keyModesOfUseModel{
		{
			Decrypt:        oldObj.Decrypt,
			DeriveKey:      oldObj.DeriveKey,
			Encrypt:        oldObj.Encrypt,
			Generate:       oldObj.Generate,
			NoRestrictions: oldObj.NoRestrictions,
			Sign:           oldObj.Sign,
			Unwrap:         oldObj.Unwrap,
			Verify:         oldObj.Verify,
			Wrap:           oldObj.Wrap,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newObj)
	diags.Append(d...)

	return result
}
