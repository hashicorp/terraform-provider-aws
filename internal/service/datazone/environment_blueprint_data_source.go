// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Environment Blueprint")
func newDataSourceEnvironmentBlueprint(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceEnvironmentBlueprint{}, nil
}

const (
	DSNameEnvironmentBlueprint = "Environment Blueprint Data Source"
)

type dataSourceEnvironmentBlueprint struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceEnvironmentBlueprint) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_datazone_environment_blueprint"
}

func (d *dataSourceEnvironmentBlueprint) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"blueprint_provider": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"domain_id": schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
			"managed": schema.BoolAttribute{
				Required: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceEnvironmentBlueprint) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().DataZoneClient(ctx)

	var data environmentBlueprintDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEnvironmentBlueprintByName(ctx, conn, data.DomainId.ValueString(), data.Name.ValueString(), data.Managed.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionReading, DSNameEnvironmentBlueprint, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	data.BlueprintProvider = flex.StringToFramework(ctx, out.Provider)
	data.Description = flex.StringToFramework(ctx, out.Description)
	data.ID = flex.StringToFramework(ctx, out.Id)
	data.Name = flex.StringToFramework(ctx, out.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findEnvironmentBlueprintByName(ctx context.Context, conn *datazone.Client, domainId, name string, managed bool) (*awstypes.EnvironmentBlueprintSummary, error) {
	return _findEnvironmentBlueprintByName(ctx, conn, domainId, name, managed, nil)
}

func _findEnvironmentBlueprintByName(ctx context.Context, conn *datazone.Client, domainId, name string, managed bool, nextToken *string) (*awstypes.EnvironmentBlueprintSummary, error) {
	in := &datazone.ListEnvironmentBlueprintsInput{
		DomainIdentifier: aws.String(domainId),
		Managed:          aws.Bool(managed),
	}

	if nextToken != nil {
		in.NextToken = aws.String(*nextToken)
	}

	out, err := conn.ListEnvironmentBlueprints(ctx, in)
	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	for i := range out.Items {
		blueprint := out.Items[i]
		if name == aws.ToString(blueprint.Name) {
			return &blueprint, nil
		}
	}

	if out.NextToken == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return _findEnvironmentBlueprintByName(ctx, conn, domainId, name, managed, out.NextToken)
}

type environmentBlueprintDataSourceModel struct {
	BlueprintProvider types.String `tfsdk:"blueprint_provider"`
	Description       types.String `tfsdk:"description"`
	DomainId          types.String `tfsdk:"domain_id"`
	ID                types.String `tfsdk:"id"`
	Managed           types.Bool   `tfsdk:"managed"`
	Name              types.String `tfsdk:"name"`
}
