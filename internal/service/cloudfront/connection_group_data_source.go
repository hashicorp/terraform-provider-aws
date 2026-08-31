// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudfront_connection_group", name="Connection Group")
// @Tags(identifierAttribute="arn")
func newConnectionGroupDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &connectionGroupDataSource{}
	return d, nil
}

const (
	DSNameConnectionGroup = "Connection Group Data Source"
)

type connectionGroupDataSource struct {
	framework.DataSourceWithModel[connectionGroupDataSourceModel]
}

func (d *connectionGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"anycast_ip_list_id": schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrEnabled: schema.BoolAttribute{
				Computed: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"ipv6_enabled": schema.BoolAttribute{
				Computed: true,
			},
			"is_default": schema.BoolAttribute{
				Computed: true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"routing_endpoint": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *connectionGroupDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data connectionGroupDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CloudFrontClient(ctx)

	var output any
	var err error

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		output, err = findConnectionGroupByID(ctx, conn, data.ID.ValueString())
	} else if !data.RoutingEndpoint.IsNull() && !data.RoutingEndpoint.IsUnknown() {
		output, err = findConnectionGroupByRoutingEndpoint(ctx, conn, data.RoutingEndpoint.ValueString())
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameConnectionGroup, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	var connectionGroup *awstypes.ConnectionGroup
	var etag *string

	switch v := output.(type) {
	case *cloudfront.GetConnectionGroupOutput:
		connectionGroup = v.ConnectionGroup
		etag = v.ETag
	case *cloudfront.GetConnectionGroupByRoutingEndpointOutput:
		connectionGroup = v.ConnectionGroup
		etag = v.ETag
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, connectionGroup, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = fwflex.StringToFramework(ctx, connectionGroup.Id)
	data.ETag = fwflex.StringToFramework(ctx, etag)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, connectionGroup.LastModifiedTime)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type connectionGroupDataSourceModel struct {
	AnycastIPListID  types.String      `tfsdk:"anycast_ip_list_id"`
	ARN              types.String      `tfsdk:"arn"`
	Enabled          types.Bool        `tfsdk:"enabled"`
	ETag             types.String      `tfsdk:"etag"`
	ID               types.String      `tfsdk:"id"`
	IPv6Enabled      types.Bool        `tfsdk:"ipv6_enabled"`
	IsDefault        types.Bool        `tfsdk:"is_default"`
	LastModifiedTime timetypes.RFC3339 `tfsdk:"last_modified_time"`
	Name             types.String      `tfsdk:"name"`
	RoutingEndpoint  types.String      `tfsdk:"routing_endpoint"`
	Status           types.String      `tfsdk:"status"`
	Tags             tftags.Map        `tfsdk:"tags"`
}

func (d *connectionGroupDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrID),
			path.MatchRoot("routing_endpoint"),
		),
	}
}
