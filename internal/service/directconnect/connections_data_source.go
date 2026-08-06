// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @FrameworkDataSource("aws_dx_connections", name="Connections")
func newConnectionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &connectionsDataSource{}, nil
}

type connectionsDataSource struct {
	framework.DataSourceWithModel[connectionsDataSourceModel]
}

func (d *connectionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"connections": framework.DataSourceComputedListOfObjectAttribute[connectionsDataSourceItemModel](ctx),
		},
	}
}

func (d *connectionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().DirectConnectClient(ctx)

	var data connectionsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	// DescribeConnections has no server-side filter: the input accepts only a
	// single connection ID. Read everything and let users filter in their
	// configuration.
	output, err := findConnections(ctx, conn, &directconnect.DescribeConnectionsInput{}, tfslices.PredicateTrue[*awstypes.Connection]())

	// An account with no connections is not an error: `connections` must still
	// be an empty list so that it remains usable in a `for_each` expression.
	if err != nil && !retry.NotFound(err) {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	connections, diags := flattenConnections(ctx, output, d.Meta().Partition(ctx), d.Meta().IgnoreTagsConfig(ctx))
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Connections = connections

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

// flattenConnections takes the partition and ignore-tags configuration as plain
// parameters so that it can be exercised without a configured provider Meta().
//
// It wraps rather than replaces `fwflex.Flatten`: the two per-item fields below
// cannot come from AutoFlex. The ARN is absent from the API response and has to
// be synthesized, and the tags must flatten to an empty map rather than null.
// nosemgrep:ci.semgrep.framework.manual-flattener-functions
func flattenConnections(ctx context.Context, apiObjects []awstypes.Connection, partition string, ignoreTagsConfig *tftags.IgnoreConfig) (fwtypes.ListNestedObjectValueOf[connectionsDataSourceItemModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	// Non-nil so that zero connections flatten to an empty list, never null.
	items := make([]*connectionsDataSourceItemModel, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		var item connectionsDataSourceItemModel
		diags.Append(fwflex.Flatten(ctx, apiObject, &item)...)
		if diags.HasError() {
			return fwtypes.NewListNestedObjectValueOfUnknown[connectionsDataSourceItemModel](ctx), diags
		}

		// The API response carries no ARN, so build one from the connection's own
		// Region and owner account: for a hosted connection both differ from the
		// caller's.
		item.ARN = types.StringValue(arn.ARN{
			Partition: partition,
			Region:    aws.ToString(apiObject.Region),
			Service:   "directconnect",
			AccountID: aws.ToString(apiObject.OwnerAccount),
			Resource:  "dxcon/" + aws.ToString(apiObject.ConnectionId),
		}.String())

		// Transparent tagging doesn't work for data sources yet. Use the tags
		// already in the response rather than a DescribeTags call per connection.
		// The legacy flattener is required here: an untagged connection must
		// flatten to an empty map, not to null.
		item.Tags = tftags.NewMapFromMapValue(fwflex.FlattenFrameworkStringValueMapLegacy(ctx, keyValueTags(ctx, apiObject.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()))

		items = append(items, &item)
	}

	return fwtypes.NewListNestedObjectValueOfSlice(ctx, items, nil)
}

type connectionsDataSourceModel struct {
	framework.WithRegionModel
	Connections fwtypes.ListNestedObjectValueOf[connectionsDataSourceItemModel] `tfsdk:"connections"`
}

type connectionsDataSourceItemModel struct {
	ARN types.String `tfsdk:"arn" autoflex:"-"`
	// AwsDeviceV2 supersedes the deprecated AwsDevice field. The Go field keeps
	// the API name so that AutoFlex matches the right one.
	AwsDeviceV2     types.String `tfsdk:"aws_device"`
	Bandwidth       types.String `tfsdk:"bandwidth"`
	ConnectionID    types.String `tfsdk:"id"`
	ConnectionName  types.String `tfsdk:"name"`
	ConnectionState types.String `tfsdk:"state"`
	Location        types.String `tfsdk:"location"`
	OwnerAccount    types.String `tfsdk:"owner_account_id"`
	PartnerName     types.String `tfsdk:"partner_name"`
	ProviderName    types.String `tfsdk:"provider_name"`
	Tags            tftags.Map   `tfsdk:"tags"`
	Vlan            types.Int64  `tfsdk:"vlan_id"`
}
