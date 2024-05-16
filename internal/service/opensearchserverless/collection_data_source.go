// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Collection")
func newDataSourceCollection(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCollection{}, nil
}

const (
	DSNameCollection = "Collection Data Source"
)

type dataSourceCollection struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceCollection) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_opensearchserverless_collection"
}

func (d *dataSourceCollection) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"collection_endpoint": schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreatedDate: schema.StringAttribute{
				Computed: true,
			},
			"dashboard_endpoint": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName(names.AttrName),
					),
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName(names.AttrName),
					),
				},
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				Computed: true,
			},
			"last_modified_date": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName(names.AttrID),
					),
				},
			},
			"standby_replicas": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}
func (d *dataSourceCollection) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data dataSourceCollectionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var out *awstypes.CollectionDetail

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		output, err := findCollectionByID(ctx, conn, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameCollection, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		out = output
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		output, err := findCollectionByName(ctx, conn, data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameCollection, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		out = output
	}

	createdDate := time.UnixMilli(aws.ToInt64(out.CreatedDate))
	data.CreatedDate = flex.StringValueToFramework(ctx, createdDate.Format(time.RFC3339))

	lastModifiedDate := time.UnixMilli(aws.ToInt64(out.LastModifiedDate))
	data.LastModifiedDate = flex.StringValueToFramework(ctx, lastModifiedDate.Format(time.RFC3339))

	ignoreTagsConfig := d.Meta().IgnoreTagsConfig
	tags, err := listTags(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameCollection, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	tags = tags.IgnoreConfig(ignoreTagsConfig)
	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceCollectionData struct {
	ARN                types.String `tfsdk:"arn"`
	CollectionEndpoint types.String `tfsdk:"collection_endpoint"`
	CreatedDate        types.String `tfsdk:"created_date"`
	DashboardEndpoint  types.String `tfsdk:"dashboard_endpoint"`
	Description        types.String `tfsdk:"description"`
	ID                 types.String `tfsdk:"id"`
	KmsKeyARN          types.String `tfsdk:"kms_key_arn"`
	LastModifiedDate   types.String `tfsdk:"last_modified_date"`
	Name               types.String `tfsdk:"name"`
	StandbyReplicas    types.String `tfsdk:"standby_replicas"`
	Tags               types.Map    `tfsdk:"tags"`
	Type               types.String `tfsdk:"type"`
}
