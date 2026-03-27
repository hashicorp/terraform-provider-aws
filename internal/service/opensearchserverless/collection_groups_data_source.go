// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_opensearchserverless_collection_groups", name="Collection Groups")
func newCollectionGroupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &collectionGroupsDataSource{}, nil
}

const (
	DSNameCollectionGroups = "Collection Groups Data Source"
)

type collectionGroupsDataSource struct {
	framework.DataSourceWithModel[collectionGroupsDataSourceModel]
}

func (d *collectionGroupsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"collection_group_summaries": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[collectionGroupSummaryModel](ctx),
				Description: "List of collection group summaries.",
				Computed:    true,
				ElementType: fwtypes.NewObjectTypeOf[collectionGroupSummaryModel](ctx),
			},
		},
	}
}

func (d *collectionGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data collectionGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &opensearchserverless.ListCollectionGroupsInput{}

	output, err := conn.ListCollectionGroups(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameCollectionGroups, "", err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type collectionGroupsDataSourceModel struct {
	framework.WithRegionModel
	CollectionGroupSummaries fwtypes.ListNestedObjectValueOf[collectionGroupSummaryModel] `tfsdk:"collection_group_summaries"`
}

type collectionGroupSummaryModel struct {
	ARN                  types.String                                  `tfsdk:"arn"`
	CapacityLimits       fwtypes.ObjectValueOf[capacityLimitsModel]    `tfsdk:"capacity_limits"`
	CreatedDate          types.Int64                                   `tfsdk:"created_date"`
	ID                   types.String                                  `tfsdk:"id"`
	Name                 types.String                                  `tfsdk:"name"`
	NumberOfCollections  types.Int64                                   `tfsdk:"number_of_collections"`
	StandbyReplicas      types.String                                  `tfsdk:"standby_replicas"`
}
