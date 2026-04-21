// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ecs_daemon", name="Daemon")
// @Tags
func newDaemonDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &daemonDataSource{}, nil
}

type daemonDataSource struct {
	framework.DataSourceWithModel[daemonDataSourceModel]
}

func (d *daemonDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Required: true,
			},
			"capacity_provider_arns": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			"cluster": schema.StringAttribute{
				Computed: true,
			},
			"daemon_task_definition": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *daemonDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data daemonDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ECSClient(ctx)

	arn := data.DaemonArn.ValueString()
	daemon, err := findDaemonByARN(ctx, conn, arn)

	if err != nil {
		response.Diagnostics.AddError("reading ECS Daemon ("+arn+")", err.Error())
		return
	}

	// AutoFlex handles DaemonArn, ClusterArn, Status
	response.Diagnostics.Append(fwflex.Flatten(ctx, daemon, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Manual: extract daemon name from ARN
	if daemon.DaemonArn != nil {
		arnParts := strings.Split(aws.ToString(daemon.DaemonArn), "/")
		if len(arnParts) >= 3 {
			data.DaemonName = types.StringValue(arnParts[len(arnParts)-1])
		}
	}

	// Manual: get task definition and capacity providers from current revisions
	if len(daemon.CurrentRevisions) > 0 {
		currentRevision := daemon.CurrentRevisions[0]

		if currentRevision.Arn != nil {
			revision, err := findDaemonRevisionByARN(ctx, conn, aws.ToString(currentRevision.Arn))
			if err != nil {
				response.Diagnostics.AddError("reading ECS Daemon Revision ("+aws.ToString(currentRevision.Arn)+")", err.Error())
				return
			}
			if revision.DaemonTaskDefinitionArn != nil {
				data.DaemonTaskDefinitionArn = types.StringPointerValue(revision.DaemonTaskDefinitionArn)
			}
		}

		cpArns := make([]string, 0)
		for _, cp := range currentRevision.CapacityProviders {
			if cp.Arn != nil {
				cpArns = append(cpArns, aws.ToString(cp.Arn))
			}
		}
		data.CapacityProviderArns = fwflex.FlattenFrameworkStringValueListOfString(ctx, cpArns)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type daemonDataSourceModel struct {
	DaemonArn               types.String         `tfsdk:"arn"`
	CapacityProviderArns    fwtypes.ListOfString `tfsdk:"capacity_provider_arns"`
	ClusterArn              types.String         `tfsdk:"cluster"`
	DaemonTaskDefinitionArn types.String         `tfsdk:"daemon_task_definition"`
	DaemonName              types.String         `tfsdk:"name"`
	Region                  types.String         `tfsdk:"region"`
	Status                  types.String         `tfsdk:"status"`
	Tags                    tftags.Map           `tfsdk:"tags"`
}
