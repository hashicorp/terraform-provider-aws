// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func schemaV0(ctx context.Context) *schema.Schema {
	return &schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrRegion: schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"log_configurations": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"event_source": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.EventSourceName](),
						},
						"log_level": schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.StringEnumType[awstypes.LogLevel](),
						},
					},
					Blocks: map[string]schema.Block{
						"cloud_watch_logs_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchLogsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"log_group_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"firehose_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[firehoseConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrStreamARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
						"s3_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func upgradeLogDeliveryConfigurationSchemaV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	type resourceLogDeliveryConfigurationModelV0 struct {
		framework.WithRegionModel
		UserPoolID        types.String                                           `tfsdk:"user_pool_id"`
		LogConfigurations fwtypes.ListNestedObjectValueOf[logConfigurationModel] `tfsdk:"log_configurations"`
	}
	var logDeliveryConfigurationDataV0 resourceLogDeliveryConfigurationModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &logDeliveryConfigurationDataV0)...)
	if resp.Diagnostics.HasError() {
		return
	}
	logConfigurationSlice, diag := logDeliveryConfigurationDataV0.LogConfigurations.ToSlice(ctx)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}
	logDeliveryConfigurationDataV1 := resourceLogDeliveryConfigurationModel{
		UserPoolID:        logDeliveryConfigurationDataV0.UserPoolID,
		LogConfigurations: fwtypes.NewSetNestedObjectValueOfSliceMust[logConfigurationModel](ctx, logConfigurationSlice),
	}
	logDeliveryConfigurationDataV1.Region = logDeliveryConfigurationDataV0.Region
	resp.Diagnostics.Append(resp.State.Set(ctx, logDeliveryConfigurationDataV1)...)
}
