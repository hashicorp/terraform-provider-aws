// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Finding Ids")
func newDataSourceFindingIds(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceFindingIds{}, nil
}

const (
	DSNameFindingIds = "Finding Ids Data Source"
)

type dataSourceFindingIds struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceFindingIds) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_guardduty_finding_ids"
}

func (d *dataSourceFindingIds) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"detector_id": schema.StringAttribute{
				Required: true,
			},
			"has_findings": schema.BoolAttribute{
				Computed: true,
			},
			"finding_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (d *dataSourceFindingIds) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().GuardDutyConn(ctx)

	var data dataSourceFindingIdsData
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
	data.FindingIDs = flex.FlattenFrameworkStringList(ctx, out)
	data.HasFindings = types.BoolValue((len(out) > 0))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findFindingIds(ctx context.Context, conn *guardduty.GuardDuty, id string) ([]*string, error) {
	in := &guardduty.ListFindingsInput{
		DetectorId: aws.String(id),
	}

	var findingIds []*string
	err := conn.ListFindingsPagesWithContext(ctx, in, func(page *guardduty.ListFindingsOutput, lastPage bool) bool {
		findingIds = append(findingIds, page.FindingIds...)
		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	return findingIds, nil
}

type dataSourceFindingIdsData struct {
	DetectorID  types.String `tfsdk:"detector_id"`
	HasFindings types.Bool   `tfsdk:"has_findings"`
	FindingIDs  types.List   `tfsdk:"finding_ids"`
	ID          types.String `tfsdk:"id"`
}
