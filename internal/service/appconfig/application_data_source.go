// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"iter"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_appconfig_application", name="Application")
func newDataSourceApplication(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplication{}, nil
}

type dataSourceApplication struct {
	framework.DataSourceWithModel[dataSourceApplicationModel]
}

func (d *dataSourceApplication) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[0-9a-z]{4,7}$`),
						"value must contain 4-7 lowercase letters or numbers",
					),
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
		},
	}
}

func (d *dataSourceApplication) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AppConfigClient(ctx)

	var data dataSourceApplicationModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var out *awstypes.Application
	var err error
	var input appconfig.ListApplicationsInput
	if !data.ID.IsNull() {
		out, err = findApplicationWithFilter(ctx, conn, &input, func(v *awstypes.Application) bool {
			return aws.ToString(v.Id) == data.ID.ValueString()
		}, tfslices.WithReturnFirstMatch)
	}

	if !data.Name.IsNull() {
		out, err = findApplicationWithFilter(ctx, conn, &input, func(v *awstypes.Application) bool {
			return aws.ToString(v.Name) == data.Name.ValueString()
		}, tfslices.WithReturnFirstMatch)
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data), smerr.ID, data.Name.String())
	if resp.Diagnostics.HasError() {
		return
	}

	data.ARN = flex.StringValueToFramework(ctx, d.Meta().RegionalARN(ctx, "appconfig", "application/"+data.ID.ValueString()))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.Name.String())
}

func (d *dataSourceApplication) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrID),
			path.MatchRoot(names.AttrName),
		),
	}
}

type dataSourceApplicationModel struct {
	framework.WithRegionModel
	ARN         types.String `tfsdk:"arn"`
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
}

func findApplicationWithFilter(ctx context.Context, conn *appconfig.Client, input *appconfig.ListApplicationsInput, filter tfslices.Predicate[*awstypes.Application], optFns ...tfslices.FinderOptionsFunc) (*awstypes.Application, error) {
	opts := tfslices.NewFinderOptions(optFns)
	var output []awstypes.Application
	for value, err := range listApplications(ctx, conn, input, filter) {
		if err != nil {
			return nil, err
		}

		output = append(output, value)
		if opts.ReturnFirstMatch() {
			break
		}
	}

	return tfresource.AssertSingleValueResult(output)
}

func listApplications(ctx context.Context, conn *appconfig.Client, input *appconfig.ListApplicationsInput, filter tfslices.Predicate[*awstypes.Application]) iter.Seq2[awstypes.Application, error] {
	return func(yield func(awstypes.Application, error) bool) {
		pages := appconfig.NewListApplicationsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Application{}, fmt.Errorf("listing AppConfig Applications: %w", err))
				return
			}

			for _, v := range page.Items {
				if filter(&v) {
					if !yield(v, nil) {
						return
					}
				}
			}
		}
	}
}
