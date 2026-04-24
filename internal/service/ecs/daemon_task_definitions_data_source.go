// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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

	var input ecs.ListDaemonTaskDefinitionsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	results, err := findDaemonTaskDefinitions(ctx, conn, &input)
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
			s.RegisteredAt = timetypes.NewRFC3339TimePointerValue(summary.RegisteredAt)
		}

		if summary.RegisteredBy != nil {
			s.RegisteredBy = types.StringPointerValue(summary.RegisteredBy)
		}

		if summary.DeleteRequestedAt != nil {
			s.DeleteRequestedAt = timetypes.NewRFC3339TimePointerValue(summary.DeleteRequestedAt)
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
	RegisteredAt      timetypes.RFC3339 `tfsdk:"registered_at"`
	RegisteredBy      types.String `tfsdk:"registered_by"`
	DeleteRequestedAt timetypes.RFC3339 `tfsdk:"delete_requested_at"`
	Status            types.String `tfsdk:"status"`
}
