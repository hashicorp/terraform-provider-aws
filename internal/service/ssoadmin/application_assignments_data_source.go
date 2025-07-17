// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

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

// @FrameworkDataSource("aws_ssoadmin_application_assignments", name="Application Assignments")
func newApplicationAssignmentsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &applicationAssignmentsDataSource{}, nil
}

const (
	DSNameApplicationAssignments = "Application Assignments Data Source"
)

type applicationAssignmentsDataSource struct {
	framework.DataSourceWithModel[applicationAssignmentsDataSourceModel]
}

func (d *applicationAssignmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				Required: true,
			},
			"application_assignments": framework.DataSourceComputedListOfObjectAttribute[applicationAssignmentModel](ctx),
			names.AttrID:              framework.IDAttribute(),
		},
	}
}
func (d *applicationAssignmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SSOAdminClient(ctx)

	var data applicationAssignmentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paginator := ssoadmin.NewListApplicationAssignmentsPaginator(conn, &ssoadmin.ListApplicationAssignmentsInput{
		ApplicationArn: data.ApplicationARN.ValueStringPointer(),
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

type applicationAssignmentsDataSourceModel struct {
	framework.WithRegionModel
	ApplicationARN         types.String                                                `tfsdk:"application_arn"`
	ApplicationAssignments fwtypes.ListNestedObjectValueOf[applicationAssignmentModel] `tfsdk:"application_assignments"`
	ID                     types.String                                                `tfsdk:"id"`
}

type applicationAssignmentModel struct {
	ApplicationARN types.String                               `tfsdk:"application_arn"`
	PrincipalID    types.String                               `tfsdk:"principal_id"`
	PrincipalType  fwtypes.StringEnum[awstypes.PrincipalType] `tfsdk:"principal_type"`
}
