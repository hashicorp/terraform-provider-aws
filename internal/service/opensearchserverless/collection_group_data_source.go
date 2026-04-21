// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_opensearchserverless_collection_group", name="Collection Group")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newCollectionGroupDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &collectionGroupDataSource{}, nil
}

type collectionGroupDataSource struct {
	framework.DataSourceWithModel[collectionGroupDataSourceModel]
}

func (d *collectionGroupDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			"capacity_limits": framework.DataSourceComputedListOfObjectAttribute[capacityLimitsDataSourceModel](ctx),
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Description: "Date the collection group was created.",
				Computed:    true,
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the collection group.",
				Computed:    true,
			},
			names.AttrID: schema.StringAttribute{
				Description: "ID of the collection group.",
				Optional:    true,
				Computed:    true,
			},
			names.AttrName: schema.StringAttribute{
				Description: "Name of the collection group.",
				Optional:    true,
				Computed:    true,
			},
			"standby_replicas": schema.StringAttribute{
				Description: "Indicates whether standby replicas should be used for collections in this group.",
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *collectionGroupDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data collectionGroupDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	var input opensearchserverless.BatchGetCollectionGroupInput
	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		input.Ids = []string{data.ID.ValueString()}
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		input.Names = []string{data.Name.ValueString()}
	}

	output, err := findCollectionGroup(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.String())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data, fwflex.WithIgnoredFieldNamesAppend("CreatedDate")))
	if response.Diagnostics.HasError() {
		return
	}

	data.CreatedDate = timetypes.NewRFC3339ValueMust(flex.Int64ToRFC3339StringValue(output.CreatedDate))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *collectionGroupDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot(names.AttrID),
			path.MatchRoot(names.AttrName),
		),
	}
}

type collectionGroupDataSourceModel struct {
	framework.WithRegionModel
	ARN             types.String                                                   `tfsdk:"arn"`
	CapacityLimits  fwtypes.ListNestedObjectValueOf[capacityLimitsDataSourceModel] `tfsdk:"capacity_limits"`
	CreatedDate     timetypes.RFC3339                                              `tfsdk:"created_date"`
	Description     types.String                                                   `tfsdk:"description"`
	ID              types.String                                                   `tfsdk:"id"`
	Name            types.String                                                   `tfsdk:"name"`
	StandbyReplicas types.String                                                   `tfsdk:"standby_replicas"`
	Tags            tftags.Map                                                     `tfsdk:"tags"`
}

type capacityLimitsDataSourceModel struct {
	MinIndexingCapacityInOCU types.Float32 `tfsdk:"min_indexing_capacity_in_ocu"`
	MaxIndexingCapacityInOCU types.Float32 `tfsdk:"max_indexing_capacity_in_ocu"`
	MinSearchCapacityInOCU   types.Float32 `tfsdk:"min_search_capacity_in_ocu"`
	MaxSearchCapacityInOCU   types.Float32 `tfsdk:"max_search_capacity_in_ocu"`
}
