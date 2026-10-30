// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_networkmanager_core_networks", name="Core Networks")
func newCoreNetworksDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &coreNetworksDataSource{}, nil
}

type coreNetworksDataSource struct {
	framework.DataSourceWithModel[coreNetworksDataSourceModel]
}

func (d *coreNetworksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"core_networks": framework.DataSourceComputedListOfObjectAttribute[coreNetworkSummaryModel](ctx),
			names.AttrIDs: schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttribute(),
		},
	}
}

func (d *coreNetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().NetworkManagerClient(ctx)

	var data coreNetworksDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	filter := tfslices.PredicateTrue[*awstypes.CoreNetworkSummary]()
	if tagsToMatch := tftags.New(ctx, data.Tags).IgnoreAWS().IgnoreConfig(d.Meta().IgnoreTagsConfig(ctx)); len(tagsToMatch) > 0 {
		filter = func(v *awstypes.CoreNetworkSummary) bool {
			return keyValueTags(ctx, v.Tags).ContainsAll(tagsToMatch)
		}
	}

	out, err := findCoreNetworks(ctx, conn, &networkmanager.ListCoreNetworksInput{}, filter)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data.CoreNetworks, fwflex.WithFieldNamePrefix("CoreNetwork")))
	if resp.Diagnostics.HasError() {
		return
	}

	data.IDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, tfslices.ApplyToAll(out, func(v awstypes.CoreNetworkSummary) string {
		return aws.ToString(v.CoreNetworkId)
	}))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findCoreNetworks(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListCoreNetworksInput, filter tfslices.Predicate[*awstypes.CoreNetworkSummary]) ([]awstypes.CoreNetworkSummary, error) {
	var output []awstypes.CoreNetworkSummary

	pages := networkmanager.NewListCoreNetworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.CoreNetworks {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

type coreNetworksDataSourceModel struct {
	CoreNetworks fwtypes.ListNestedObjectValueOf[coreNetworkSummaryModel] `tfsdk:"core_networks"`
	IDs          fwtypes.ListOfString                                     `tfsdk:"ids"`
	Tags         tftags.Map                                               `tfsdk:"tags"`
}

type coreNetworkSummaryModel struct {
	ARN             types.String `tfsdk:"arn"`
	Description     types.String `tfsdk:"description"`
	GlobalNetworkID types.String `tfsdk:"global_network_id"`
	ID              types.String `tfsdk:"core_network_id"`
	OwnerAccountID  types.String `tfsdk:"owner_account_id"`
	State           types.String `tfsdk:"state"`
}
