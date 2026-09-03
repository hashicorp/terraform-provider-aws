// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_accountaccess_application", name="Application")
// @Tags(identifierAttribute="arn")
// @Testing(generator=false)
// @Testing(preCheck="testAccPreCheck", serialize=true)
func newApplicationDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &applicationDataSource{}, nil
}

type applicationDataSource struct {
	framework.DataSourceWithModel[applicationDataSourceModel]
}

func (d *applicationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot(names.AttrARN),
						path.MatchRoot("identity_center_instance_arn"),
					),
				},
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"identity_center_instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"identity_source": framework.DataSourceComputedListOfObjectAttribute[identitySourceModel](ctx),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Status](),
				Computed:   true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"tenant_id": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *applicationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config applicationDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &config))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AccountAccessClient(ctx)

	var (
		arn string
		app *accountaccess.GetApplicationOutput
		err error
	)

	switch {
	case !config.ApplicationARN.IsNull():
		arn = fwflex.StringValueFromFramework(ctx, config.ApplicationARN)
		app, err = findApplicationByARN(ctx, conn, arn)
	case !config.IdentityCenterInstanceARN.IsNull():
		arn = fwflex.StringValueFromFramework(ctx, config.IdentityCenterInstanceARN)
		app, err = findApplicationByIdentityCenterInstanceARN(ctx, conn, arn)
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, app, &config))
	if response.Diagnostics.HasError() {
		return
	}

	// The tags returned from GetApplication are not to be trusted.
	// setTagsOut(ctx, app.Tags)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &config))
}

func findApplicationByIdentityCenterInstanceARN(ctx context.Context, conn *accountaccess.Client, identityCenterInstanceARN string) (*accountaccess.GetApplicationOutput, error) {
	var input accountaccess.ListApplicationsInput
	for item, err := range listApplications(ctx, conn, &input) {
		if err != nil {
			return nil, err
		}

		applicationARN := aws.ToString(item.ApplicationArn)
		app, err := findApplicationByARN(ctx, conn, applicationARN)
		if err != nil {
			return nil, err
		}

		if v, ok := app.IdentitySource.(*awstypes.IdentitySourceDetailsMemberIdentityCenter); ok && aws.ToString(v.Value.InstanceArn) == identityCenterInstanceARN {
			return app, nil
		}
	}

	return nil, tfresource.NewEmptyResultError()
}

type applicationDataSourceModel struct {
	framework.WithRegionModel
	ApplicationARN            fwtypes.ARN                                          `tfsdk:"arn"`
	CreatedAt                 timetypes.RFC3339                                    `tfsdk:"created_at"`
	IdentityCenterInstanceARN fwtypes.ARN                                          `tfsdk:"identity_center_instance_arn"`
	IdentitySource            fwtypes.ListNestedObjectValueOf[identitySourceModel] `tfsdk:"identity_source"`
	Status                    fwtypes.StringEnum[awstypes.Status]                  `tfsdk:"status"`
	Tags                      tftags.Map                                           `tfsdk:"tags"`
	TenantID                  types.String                                         `tfsdk:"tenant_id"`
	UpdatedAt                 timetypes.RFC3339                                    `tfsdk:"updated_at"`
}
