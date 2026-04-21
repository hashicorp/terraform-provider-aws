// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_ecs_daemon_task_definitions", name="Daemon Task Definitions")
func newDaemonTaskDefinitionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &daemonTaskDefinitionsDataSource{}, nil
}

type daemonTaskDefinitionsDataSource struct {
	framework.DataSourceWithModel[daemonTaskDefinitionsDataSourceModel]
}

func (d *daemonTaskDefinitionsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"family": schema.StringAttribute{
				Optional: true,
			},
			"family_prefix": schema.StringAttribute{
				Optional: true,
			},
			"revision": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.DaemonTaskDefinitionRevisionFilter](),
			},
			"sort": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.SortOrder](),
			},
			"status": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.DaemonTaskDefinitionStatusFilter](),
			},
			"daemon_task_definitions": framework.DataSourceComputedListOfObjectAttribute[daemonTaskDefinitionSummaryModel](ctx),
		},
	}
}

func (d *daemonTaskDefinitionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data daemonTaskDefinitionsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECSClient(ctx)

	input := &ecs.ListDaemonTaskDefinitionsInput{}

	if !data.Family.IsNull() {
		input.Family = aws.String(data.Family.ValueString())
	}

	if !data.FamilyPrefix.IsNull() {
		input.FamilyPrefix = aws.String(data.FamilyPrefix.ValueString())
	}

	if !data.Revision.IsNull() {
		input.Revision = data.Revision.ValueEnum()
	}

	if !data.Sort.IsNull() {
		input.Sort = data.Sort.ValueEnum()
	}

	if !data.Status.IsNull() {
		input.Status = data.Status.ValueEnum()
	}

	results, err := findDaemonTaskDefinitions(ctx, conn, input)
	if err != nil {
		response.Diagnostics.AddError("listing ECS Daemon Task Definitions", err.Error())
		return
	}

	var summaries []daemonTaskDefinitionSummaryModel
	for _, summary := range results {
		s := daemonTaskDefinitionSummaryModel{
			Arn:    types.StringValue(aws.ToString(summary.Arn)),
			Status: types.StringValue(string(summary.Status)),
		}

		if summary.RegisteredAt != nil {
			s.RegisteredAt = types.StringValue(aws.ToTime(summary.RegisteredAt).Format(time.RFC3339))
		}

		if summary.RegisteredBy != nil {
			s.RegisteredBy = types.StringPointerValue(summary.RegisteredBy)
		}

		if summary.DeleteRequestedAt != nil {
			s.DeleteRequestedAt = types.StringValue(aws.ToTime(summary.DeleteRequestedAt).Format(time.RFC3339))
		}

		summaries = append(summaries, s)
	}

	data.DaemonTaskDefinitions = fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, summaries)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type daemonTaskDefinitionsDataSourceModel struct {
	Family                types.String                                                      `tfsdk:"family"`
	FamilyPrefix          types.String                                                      `tfsdk:"family_prefix"`
	Region                types.String                                                      `tfsdk:"region"`
	Revision              fwtypes.StringEnum[awstypes.DaemonTaskDefinitionRevisionFilter]   `tfsdk:"revision"`
	Sort                  fwtypes.StringEnum[awstypes.SortOrder]                            `tfsdk:"sort"`
	Status                fwtypes.StringEnum[awstypes.DaemonTaskDefinitionStatusFilter]     `tfsdk:"status"`
	DaemonTaskDefinitions fwtypes.ListNestedObjectValueOf[daemonTaskDefinitionSummaryModel] `tfsdk:"daemon_task_definitions"`
}

type daemonTaskDefinitionSummaryModel struct {
	Arn               types.String `tfsdk:"arn"`
	RegisteredAt      types.String `tfsdk:"registered_at"`
	RegisteredBy      types.String `tfsdk:"registered_by"`
	DeleteRequestedAt types.String `tfsdk:"delete_requested_at"`
	Status            types.String `tfsdk:"status"`
}
