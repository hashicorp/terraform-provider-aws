// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	awstypes "github.com/aws/aws-sdk-go-v2/service/accountaccess/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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

func (d *applicationDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrARN),
			path.MatchRoot("identity_center_instance_arn"),
		),
	}
}

func (d *applicationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"identity_center_application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"identity_center_instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
			},
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
	var data applicationDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AccountAccessClient(ctx)

	lookupID := data.ARN.ValueString()
	var (
		arn string
		app *accountaccess.GetApplicationOutput
		err error
	)

	switch {
	case !data.ARN.IsNull():
		arn = lookupID
		app, err = findApplicationByARN(ctx, conn, arn)
	case !data.IdentityCenterInstanceARN.IsNull():
		lookupID = data.IdentityCenterInstanceARN.ValueString()
		arn, err = findApplicationARNByIdentityCenterInstance(ctx, conn, lookupID)
		if err == nil && arn != "" {
			app, err = findApplicationByARN(ctx, conn, arn)
		}
	}

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, lookupID)
		return
	}
	if app == nil || arn == "" {
		response.Diagnostics.AddError(
			"reading Account Access Application: not found",
			"no Application matched the supplied lookup attributes",
		)
		return
	}

	data.ARN = fwtypes.ARNValue(arn)
	data.Status = fwtypes.StringEnumValue(app.Status)
	data.TenantID = fwflex.StringToFramework(ctx, app.TenantId)
	data.CreatedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(app.CreatedAt))
	data.UpdatedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(app.UpdatedAt))

	if details, ok := app.IdentitySource.(*awstypes.IdentitySourceDetailsMemberIdentityCenter); ok {
		data.IdentityCenterInstanceARN = fwtypes.ARNValue(aws.ToString(details.Value.InstanceArn))
		data.IdentityCenterApplicationARN = fwtypes.ARNValue(aws.ToString(details.Value.ApplicationArn))
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

// findApplicationARNByIdentityCenterInstance lists applications and matches
// each candidate's Identity Center instance against the requested ARN.
func findApplicationARNByIdentityCenterInstance(ctx context.Context, conn *accountaccess.Client, instanceARN string) (string, error) {
	var input accountaccess.ListApplicationsInput
	for summary, err := range listApplications(ctx, conn, &input) {
		if err != nil {
			return "", err
		}

		applicationARN := aws.ToString(summary.ApplicationArn)
		app, err := findApplicationByARN(ctx, conn, applicationARN)
		if err != nil {
			return "", err
		}

		details, ok := app.IdentitySource.(*awstypes.IdentitySourceDetailsMemberIdentityCenter)
		if ok && aws.ToString(details.Value.InstanceArn) == instanceARN {
			return applicationARN, nil
		}
	}

	return "", nil
}

type applicationDataSourceModel struct {
	framework.WithRegionModel
	ARN                          fwtypes.ARN                         `tfsdk:"arn"`
	CreatedAt                    timetypes.RFC3339                   `tfsdk:"created_at"`
	IdentityCenterApplicationARN fwtypes.ARN                         `tfsdk:"identity_center_application_arn"`
	IdentityCenterInstanceARN    fwtypes.ARN                         `tfsdk:"identity_center_instance_arn"`
	Status                       fwtypes.StringEnum[awstypes.Status] `tfsdk:"status"`
	Tags                         tftags.Map                          `tfsdk:"tags"`
	TenantID                     types.String                        `tfsdk:"tenant_id"`
	UpdatedAt                    timetypes.RFC3339                   `tfsdk:"updated_at"`
}
