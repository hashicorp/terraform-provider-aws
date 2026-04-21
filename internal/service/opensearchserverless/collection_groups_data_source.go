// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_opensearchserverless_collection_groups", name="Collection Groups")
func newCollectionGroupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &collectionGroupsDataSource{}, nil
}

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

func (d *collectionGroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data collectionGroupsDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	input := &opensearchserverless.ListCollectionGroupsInput{}
	var output []summariesData
	for item, err := range listCollectionGroups(ctx, conn, input) {
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err)
			return
		}

		s := summariesData{
			item,
			flex.Int64ToRFC3339StringValue(item.CreatedDate),
		}
		output = append(output, s)
	}

	da := struct {
		CollectionGroupSummaries []summariesData `json:"collection_group_summaries"`
	}{
		CollectionGroupSummaries: output,
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, da, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type collectionGroupsDataSourceModel struct {
	framework.WithRegionModel
	CollectionGroupSummaries fwtypes.ListNestedObjectValueOf[collectionGroupSummaryModel] `tfsdk:"collection_group_summaries"`
}

type collectionGroupSummaryModel struct {
	ARN                 types.String                                                   `tfsdk:"arn"`
	CapacityLimits      fwtypes.ListNestedObjectValueOf[capacityLimitsDataSourceModel] `tfsdk:"capacity_limits"`
	CreatedAt           timetypes.RFC3339                                              `tfsdk:"created_date"`
	ID                  types.String                                                   `tfsdk:"id"`
	Name                types.String                                                   `tfsdk:"name"`
	NumberOfCollections types.Int32                                                    `tfsdk:"number_of_collections"`
	StandbyReplicas     types.String                                                   `tfsdk:"standby_replicas"`
}

type summariesData struct {
	awstypes.CollectionGroupSummary
	CreatedAt string `tfsdk:"created_date"`
}
