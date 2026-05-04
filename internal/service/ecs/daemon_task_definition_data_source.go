// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecs_daemon_task_definition", name="Daemon Task Definition")
func newDaemonTaskDefinitionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &daemonTaskDefinitionDataSource{}, nil
}

type daemonTaskDefinitionDataSource struct {
	framework.DataSourceWithModel[daemonTaskDefinitionDataSourceModel]
}

func (d *daemonTaskDefinitionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"container_definition": framework.DataSourceComputedListOfObjectAttribute[containerDefinitionModel](ctx),
			"cpu": schema.StringAttribute{
				Computed: true,
			},
			"daemon_task_definition": schema.StringAttribute{
				Required: true,
			},
			"delete_requested_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrExecutionRoleARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrFamily: schema.StringAttribute{
				Computed: true,
			},
			"memory": schema.StringAttribute{
				Computed: true,
			},
			"registered_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"registered_by": schema.StringAttribute{
				Computed: true,
			},
			"revision": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"task_role_arn": schema.StringAttribute{
				Computed: true,
			},
			"volume": framework.DataSourceComputedListOfObjectAttribute[volumeModel](ctx),
		},
	}
}

func (d *daemonTaskDefinitionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data daemonTaskDefinitionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECSClient(ctx)

	input := &ecs.DescribeDaemonTaskDefinitionInput{
		DaemonTaskDefinition: data.DaemonTaskDefinition.ValueStringPointer(),
	}

	output, err := conn.DescribeDaemonTaskDefinition(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ECS Daemon Task Definition (%s)", data.DaemonTaskDefinition.ValueString()), err.Error())
		return
	}

	dtd := output.DaemonTaskDefinition

	// AutoFlex handles DaemonTaskDefinitionArn, Family, Revision, Cpu, Memory,
	// ExecutionRoleArn, TaskRoleArn, Status, RegisteredBy
	response.Diagnostics.Append(fwflex.Flatten(ctx, dtd, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Manual: volumes (structural mismatch — Host.SourcePath → HostPath)
	var volumeDiags diag.Diagnostics
	data.Volumes, volumeDiags = fwtypes.NewListNestedObjectValueOfValueSlice(ctx, flattenDaemonVolumes(dtd.Volumes))
	response.Diagnostics.Append(volumeDiags...)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type daemonTaskDefinitionDataSourceModel struct {
	framework.WithRegionModel
	ContainerDefinitions    fwtypes.ListNestedObjectValueOf[containerDefinitionModel] `tfsdk:"container_definition"`
	Cpu                     types.String                                              `tfsdk:"cpu"`
	DaemonTaskDefinition    types.String                                              `tfsdk:"daemon_task_definition"`
	DaemonTaskDefinitionArn types.String                                              `tfsdk:"arn"`
	DeleteRequestedAt       timetypes.RFC3339                                         `tfsdk:"delete_requested_at"`
	ExecutionRoleArn        types.String                                              `tfsdk:"execution_role_arn"`
	Family                  types.String                                              `tfsdk:"family"`
	Memory                  types.String                                              `tfsdk:"memory"`
	RegisteredAt            timetypes.RFC3339                                         `tfsdk:"registered_at"`
	RegisteredBy            types.String                                              `tfsdk:"registered_by"`
	Revision                types.Int64                                               `tfsdk:"revision"`
	Status                  types.String                                              `tfsdk:"status"`
	TaskRoleArn             types.String                                              `tfsdk:"task_role_arn"`
	Volumes                 fwtypes.ListNestedObjectValueOf[volumeModel]              `tfsdk:"volume"`
}

type volumeModel struct {
	HostPath types.String `tfsdk:"host_path"`
	Name     types.String `tfsdk:"name"`
}
