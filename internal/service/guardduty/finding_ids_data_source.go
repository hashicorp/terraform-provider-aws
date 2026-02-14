// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package guardduty

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_guardduty_finding_ids", name="Finding Ids")
func newFindingIDsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &findingIDsDataSource{}, nil
}

const (
	DSNameFindingIds = "Finding Ids Data Source"
)

type findingIDsDataSource struct {
	framework.DataSourceWithModel[findingIDsDataSourceModel]
}

func (d *findingIDsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"detector_id": schema.StringAttribute{
				Required: true,
			},
			"has_findings": schema.BoolAttribute{
				Computed: true,
			},
			"finding_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (d *findingIDsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().GuardDutyClient(ctx)

	var data findingIDsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFindingIds(ctx, conn, data.DetectorID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.GuardDuty, create.ErrActionReading, DSNameFindingIds, data.DetectorID.String(), err),
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(data.DetectorID.ValueString())
	data.FindingIDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, out)
	data.HasFindings = types.BoolValue((len(out) > 0))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findFindingIds(ctx context.Context, conn *guardduty.Client, id string) ([]string, error) {
	in := &guardduty.ListFindingsInput{
		DetectorId: aws.String(id),
	}

	var findingIds []string

	pages := guardduty.NewListFindingsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The request is rejected because the input detectorId is not owned by the current account.") {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		findingIds = append(findingIds, page.FindingIds...)
	}

	return findingIds, nil
}

type findingIDsDataSourceModel struct {
	framework.WithRegionModel
	DetectorID  types.String         `tfsdk:"detector_id"`
	HasFindings types.Bool           `tfsdk:"has_findings"`
	FindingIDs  fwtypes.ListOfString `tfsdk:"finding_ids"`
	ID          types.String         `tfsdk:"id"`
}
