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

// @FrameworkDataSource("aws_opensearchserverless_collection", name="Collection")
// @Tags(identifierAttribute="arn")
func newCollectionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &collectionDataSource{}, nil
}

const (
	DSNameCollection = "Collection Data Source"
)

type collectionDataSource struct {
	framework.DataSourceWithModel[collectionDataSourceModel]
}

func (d *collectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"collection_endpoint": schema.StringAttribute{
				Description: "Collection-specific endpoint used to submit index, search, and data upload requests to an OpenSearch Serverless collection.",
				Computed:    true,
			},
			names.AttrCreatedDate: schema.StringAttribute{
				Description: "Date the Collection was created.",
				Computed:    true,
			},
			"dashboard_endpoint": schema.StringAttribute{
				Description: "Collection-specific endpoint used to access OpenSearch Dashboards.",
				Computed:    true,
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the collection.",
				Computed:    true,
			},
			"failure_message": schema.StringAttribute{
				Description: "A failure reason associated with the collection.",
				Computed:    true,
			},
			"failure_code": schema.StringAttribute{
				Description: "A failure code associated with the collection.",
				Computed:    true,
			},
			names.AttrID: schema.StringAttribute{
				Description: "ID of the collection.",
				Optional:    true,
				Computed:    true,
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
				Description: "The ARN of the Amazon Web Services KMS key used to encrypt the collection.",
				Computed:    true,
			},
			"last_modified_date": schema.StringAttribute{
				Description: "Date the Collection was last modified.",
				Computed:    true,
			},
			names.AttrName: schema.StringAttribute{
				Description: "Name of the collection.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(
						path.MatchRelative().AtParent().AtName(names.AttrID),
					),
				},
			},
			"standby_replicas": schema.StringAttribute{
				Description: "Indicates whether standby replicas should be used for a collection.",
				Computed:    true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				Description: "Type of collection.",
				Computed:    true,
			},
		},
	}
}
func (d *collectionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data collectionDataSourceModel
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

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data, flex.WithIgnoredFieldNames([]string{"CreatedDate", "LastModifiedDate"}))...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Special handling for Unix time conversion
	data.CreatedDate = flex.StringValueToFramework(ctx, time.UnixMilli(aws.ToInt64(out.CreatedDate)).Format(time.RFC3339))
	data.LastModifiedDate = flex.StringValueToFramework(ctx, time.UnixMilli(aws.ToInt64(out.LastModifiedDate)).Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type collectionDataSourceModel struct {
	framework.WithRegionModel
	ARN                types.String `tfsdk:"arn"`
	CollectionEndpoint types.String `tfsdk:"collection_endpoint"`
	CreatedDate        types.String `tfsdk:"created_date"`
	FailureMessage     types.String `tfsdk:"failure_message"`
	FailureCode        types.String `tfsdk:"failure_code"`
	DashboardEndpoint  types.String `tfsdk:"dashboard_endpoint"`
	Description        types.String `tfsdk:"description"`
	ID                 types.String `tfsdk:"id"`
	KmsKeyARN          types.String `tfsdk:"kms_key_arn"`
	LastModifiedDate   types.String `tfsdk:"last_modified_date"`
	Name               types.String `tfsdk:"name"`
	StandbyReplicas    types.String `tfsdk:"standby_replicas"`
	Tags               tftags.Map   `tfsdk:"tags"`
	Type               types.String `tfsdk:"type"`
}
