// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Principal Application Assignments")
func newDataSourcePrincipalApplicationAssignments(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourcePrincipalApplicationAssignments{}, nil
}

const (
	DSNamePrincipalApplicationAssignments = "Principal Application Assignments Data Source"
)

type dataSourcePrincipalApplicationAssignments struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourcePrincipalApplicationAssignments) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_ssoadmin_principal_application_assignments"
}

func (d *dataSourcePrincipalApplicationAssignments) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"principal_id": schema.StringAttribute{
				Required: true,
			},
			"principal_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PrincipalType](),
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"application_assignments": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[applicationAssignmentData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"application_arn": schema.StringAttribute{
							Computed: true,
						},
						"principal_id": schema.StringAttribute{
							Computed: true,
						},
						"principal_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.PrincipalType](),
							Computed:   true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourcePrincipalApplicationAssignments) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SSOAdminClient(ctx)

	var data dataSourcePrincipalApplicationAssignmentsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paginator := ssoadmin.NewListApplicationAssignmentsForPrincipalPaginator(conn, &ssoadmin.ListApplicationAssignmentsForPrincipalInput{
		InstanceArn:   aws.String(data.InstanceARN.ValueString()),
		PrincipalId:   aws.String(data.PrincipalID.ValueString()),
		PrincipalType: awstypes.PrincipalType(data.PrincipalType.ValueString()),
	})

	var out ssoadmin.ListApplicationAssignmentsForPrincipalOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionReading, DSNamePrincipalApplicationAssignments, data.PrincipalID.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.ApplicationAssignments) > 0 {
			out.ApplicationAssignments = append(out.ApplicationAssignments, page.ApplicationAssignments...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourcePrincipalApplicationAssignmentsData struct {
	InstanceARN            fwtypes.ARN                                                `tfsdk:"instance_arn"`
	ApplicationAssignments fwtypes.ListNestedObjectValueOf[applicationAssignmentData] `tfsdk:"application_assignments"`
	ID                     types.String                                               `tfsdk:"id"`
	PrincipalID            types.String                                               `tfsdk:"principal_id"`
	PrincipalType          fwtypes.StringEnum[awstypes.PrincipalType]                 `tfsdk:"principal_type"`
}
