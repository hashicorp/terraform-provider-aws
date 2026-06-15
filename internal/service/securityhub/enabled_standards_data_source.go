// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_securityhub_enabled_standards", name="Enabled Standards")
func newEnabledStandardsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &enabledStandardsDataSource{}

	return d, nil
}

type enabledStandardsDataSource struct {
	framework.DataSourceWithModel[enabledStandardsDataSourceModel]
}

func (d *enabledStandardsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"standards_subscriptions": framework.DataSourceComputedListOfObjectAttribute[standardsSubscriptionModel](ctx),
			"standards_subscription_arns": schema.ListAttribute{
				CustomType: fwtypes.ListOfARNType,
				Optional:   true,
			},
		},
	}
}

func (d *enabledStandardsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data enabledStandardsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SecurityHubClient(ctx)

	input := securityhub.GetEnabledStandardsInput{
		StandardsSubscriptionArns: fwflex.ExpandFrameworkStringValueList(ctx, data.StandardsSubscriptionARNs),
	}

	out, err := findStandardsSubscriptions(ctx, conn, &input)

	if err != nil {
		response.Diagnostics.AddError("reading Security Hub Enabled Standards", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data.StandardsSubscriptions)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type enabledStandardsDataSourceModel struct {
	framework.WithRegionModel
	StandardsSubscriptions    fwtypes.ListNestedObjectValueOf[standardsSubscriptionModel] `tfsdk:"standards_subscriptions"`
	StandardsSubscriptionARNs fwtypes.ListOfARN                                           `tfsdk:"standards_subscription_arns"`
}

type standardsSubscriptionModel struct {
	StandardsARN               fwtypes.ARN                                                 `tfsdk:"standards_arn"`
	StandardsControlsUpdatable fwtypes.StringEnum[awstypes.StandardsControlsUpdatable]     `tfsdk:"standards_controls_updatable"`
	StandardsInputs            fwtypes.MapOfString                                         `tfsdk:"standards_inputs"`
	StandardsStatus            fwtypes.StringEnum[awstypes.StandardsStatus]                `tfsdk:"standards_status"`
	StandardsStatusReason      fwtypes.ListNestedObjectValueOf[standardsStatusReasonModel] `tfsdk:"standards_status_reason"`
	StandardsSubscriptionARN   fwtypes.ARN                                                 `tfsdk:"standards_subscription_arn"`
}

type standardsStatusReasonModel struct {
	StatusReasonCode fwtypes.StringEnum[awstypes.StatusReasonCode] `tfsdk:"status_reason_code"`
}
