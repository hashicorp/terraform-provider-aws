// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cloudfront_key_group", name="Key Group")
func newKeyGroupDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &keyGroupDataSource{}

	return d, nil
}

const (
	DSNameKeyGroup = "Key Group Data Source"
)

type keyGroupDataSource struct {
	framework.DataSourceWithModel[keyGroupDataSourceModel]
}

func (d *keyGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrComment: schema.StringAttribute{
				Computed: true,
			},
			attrETag: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			attrItems: schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (d *keyGroupDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrID),
			path.MatchRoot(names.AttrName),
		),
	}
}

func (d *keyGroupDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data keyGroupDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CloudFrontClient(ctx)

	id := data.ID.ValueString()
	if data.ID.IsNull() {
		name := data.Name.ValueString()

		var err error
		id, err = findKeyGroupIDByName(ctx, conn, name)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, name, err),
				err.Error(),
			)
			return
		}
	}

	output, err := findKeyGroupByID(ctx, conn, id)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameKeyGroup, id, err),
			err.Error(),
		)
		return
	}

	keyGroup := output.KeyGroup
	response.Diagnostics.Append(fwflex.Flatten(ctx, keyGroup.KeyGroupConfig, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ETag = fwflex.StringToFramework(ctx, output.ETag)
	data.ID = fwflex.StringToFramework(ctx, keyGroup.Id)
	data.LastModifiedTime = fwflex.TimeToFramework(ctx, keyGroup.LastModifiedTime)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findKeyGroupIDByName(ctx context.Context, conn *cloudfront.Client, name string) (string, error) {
	var id string
	input := cloudfront.ListKeyGroupsInput{}

	err := listKeyGroupsPages(ctx, conn, &input, func(page *cloudfront.ListKeyGroupsOutput, lastPage bool) bool {
		if page == nil || page.KeyGroupList == nil {
			return !lastPage
		}

		for _, v := range page.KeyGroupList.Items {
			if keyGroup := v.KeyGroup; keyGroup != nil && keyGroup.KeyGroupConfig != nil && aws.ToString(keyGroup.KeyGroupConfig.Name) == name {
				id = aws.ToString(keyGroup.Id)

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return "", err
	}

	if id == "" {
		return "", &retry.NotFoundError{}
	}

	return id, nil
}

type keyGroupDataSourceModel struct {
	Comment          types.String        `tfsdk:"comment"`
	ETag             types.String        `tfsdk:"etag"`
	ID               types.String        `tfsdk:"id"`
	Items            fwtypes.SetOfString `tfsdk:"items"`
	LastModifiedTime timetypes.RFC3339   `tfsdk:"last_modified_time"`
	Name             types.String        `tfsdk:"name"`
}
