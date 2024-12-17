// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Key Group")
func newDataSourceKeyGroup(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceKeyGroup{}, nil
}

const (
	DSNameKeyGroup = "Key Group Data Source"
)

type dataSourceKeyGroup struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceKeyGroup) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_cloudfront_key_group"
}

func (d *dataSourceKeyGroup) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName(names.AttrID),
						path.MatchRelative().AtParent().AtName(names.AttrName),
					),
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"items": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrComment: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *dataSourceKeyGroup) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().CloudFrontClient(ctx)

	var data dataSourceKeyGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var keyGroupID string

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		keyGroupID = data.ID.ValueString()
	} else {
		name := data.Name.ValueString()
		input := &cloudfront.ListKeyGroupsInput{}

		err := listKeyGroupsPages(ctx, conn, input, func(page *cloudfront.ListKeyGroupsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, policySummary := range page.KeyGroupList.Items {
				if keyGroup := policySummary.KeyGroup; aws.ToString(keyGroup.KeyGroupConfig.Name) == name {
					keyGroupID = aws.ToString(keyGroup.Id)

					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, data.Name.String(), err),
				err.Error(),
			)
			return
		}

		if keyGroupID == "" {
			err := fmt.Errorf("no matching CloudFront Key Group (%s)", name)
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, data.Name.String(), err),
				err.Error(),
			)
			return
		}
	}

	out, err := findKeyGroupByID(ctx, conn, keyGroupID)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, keyGroupID, err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.KeyGroup.KeyGroupConfig, &data, flex.WithFieldNamePrefix("KeyGroup"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = flex.StringToFramework(ctx, out.KeyGroup.Id)
	data.Name = flex.StringToFramework(ctx, out.KeyGroup.KeyGroupConfig.Name)
	data.Items = flex.FlattenFrameworkStringValueList(ctx, out.KeyGroup.KeyGroupConfig.Items)

	var comment *string
	if out.KeyGroup.KeyGroupConfig.Comment != nil {
		comment = out.KeyGroup.KeyGroupConfig.Comment
	} else {
		emptyString := ""
		comment = &emptyString
	}
	data.Comment = flex.StringToFramework(ctx, comment)

	data.Etag = flex.StringToFramework(ctx, out.ETag)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceKeyGroupModel struct {
	Etag    types.String `tfsdk:"etag"`
	Items   types.List   `tfsdk:"items"`
	Comment types.String `tfsdk:"comment"`
	ID      types.String `tfsdk:"id"`
	Name    types.String `tfsdk:"name"`
}
