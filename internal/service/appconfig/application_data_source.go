// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_appconfig_application", name="Application")
func newDataSourceApplication(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplication{}, nil
}

const (
	DSNameApplication = "Application Data Source"
)

type dataSourceApplication struct {
	framework.DataSourceWithModel[dataSourceApplicationModel]
}

func (d *dataSourceApplication) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1024),
				},
			},
			names.AttrID: schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[0-9a-z]{4,7}$`),
						"value must contain 4-7 lowercase letters or numbers",
					),
				},
			},
			//names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrName: schema.StringAttribute{
				Required: true,
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
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationByName(ctx, conn, data.Name.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data, flex.WithFieldNamePrefix("Application")), smerr.ID, data.Name.String())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.Name.String())
}

type dataSourceApplicationModel struct {
	framework.WithRegionModel
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
}

func findApplicationByName(ctx context.Context, conn *appconfig.Client, name string) (*appconfig.GetApplicationOutput, error) {
	input := &appconfig.ListApplicationsInput{}

	pages := appconfig.NewListApplicationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, app := range page.Items {
			if aws.ToString(app.Name) == name {
				// AppConfig does not support duplicate names, so we can return the first match
				return findApplicationByID(ctx, conn, aws.ToString(app.Id))
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastError: fmt.Errorf("AppConfig Application (%s) not found", name),
	}
}
