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

// @FrameworkDataSource(name="Application Assignments")
func newDataSourceApplicationAssignments(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplicationAssignments{}, nil
}

const (
	DSNameApplicationAssignments = "Application Assignments Data Source"
)

type dataSourceApplicationAssignments struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceApplicationAssignments) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_ssoadmin_application_assignments"
}

func (d *dataSourceApplicationAssignments) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
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
func (d *dataSourceApplicationAssignments) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SSOAdminClient(ctx)

	var data dataSourceApplicationAssignmentsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paginator := ssoadmin.NewListApplicationAssignmentsPaginator(conn, &ssoadmin.ListApplicationAssignmentsInput{
		ApplicationArn: aws.String(data.ApplicationARN.ValueString()),
	})

	var out ssoadmin.ListApplicationAssignmentsOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionReading, DSNameApplicationAssignments, data.ApplicationARN.String(), err),
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

type dataSourceApplicationAssignmentsData struct {
	ApplicationARN         types.String                                               `tfsdk:"application_arn"`
	ApplicationAssignments fwtypes.ListNestedObjectValueOf[applicationAssignmentData] `tfsdk:"application_assignments"`
	ID                     types.String                                               `tfsdk:"id"`
}

type applicationAssignmentData struct {
	ApplicationARN types.String                               `tfsdk:"application_arn"`
	PrincipalID    types.String                               `tfsdk:"principal_id"`
	PrincipalType  fwtypes.StringEnum[awstypes.PrincipalType] `tfsdk:"principal_type"`
}
