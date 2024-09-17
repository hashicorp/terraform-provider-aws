// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name=StandardsControlAssociations)
func newDataSourceStandardsControlAssociations(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceStandardsControlAssociations{}

	return d, nil
}

const (
	DSNameStandardsControlAssociations = "Standards Control Associations Data Source"
)

type dataSourceStandardsControlAssociations struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceStandardsControlAssociations) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_securityhub_standards_control_associations"
}

func (d *dataSourceStandardsControlAssociations) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"security_control_id": schema.StringAttribute{
				Required: true,
			},
			"standards_arns": schema.ListAttribute{
				ElementType: fwtypes.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceStandardsControlAssociations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceStandardsControlAssociationsData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SecurityHubClient(ctx)

	input := &securityhub.ListStandardsControlAssociationsInput{
		SecurityControlId: data.SecurityControlID.ValueStringPointer(),
	}

	standardsControlAssociations, err := findStandardsControlAssociations(ctx, conn, input)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionReading, DSNameStandardsControlAssociations, data.SecurityControlID.String(), err),
			err.Error(),
		)
		return
	}

	data.ID = fwtypes.StringValue(d.Meta().Region)
	data.StandardsARNs = flex.FlattenFrameworkStringValueList(ctx, tfslices.ApplyToAll(standardsControlAssociations, func(v types.StandardsControlAssociationSummary) string {
		return aws.ToString(v.StandardsArn)
	}))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceStandardsControlAssociationsData struct {
	ID                fwtypes.String `tfsdk:"id"`
	SecurityControlID fwtypes.String `tfsdk:"security_control_id"`
	StandardsARNs     fwtypes.List   `tfsdk:"standards_arns"`
}

func findStandardsControlAssociations(ctx context.Context, conn *securityhub.Client, input *securityhub.ListStandardsControlAssociationsInput) ([]types.StandardsControlAssociationSummary, error) {
	var output []types.StandardsControlAssociationSummary

	pages := securityhub.NewListStandardsControlAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.StandardsControlAssociationSummaries...)
	}

	return output, nil
}
