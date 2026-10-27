// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_continuum_environment", name="Environment")
func newEnvironmentDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &environmentDataSource{}, nil
}

type environmentDataSource struct {
	framework.DataSourceWithModel[environmentDataSourceModel]
}

func (d *environmentDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allowed_domains": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrARN:                    framework.ARNAttributeComputedOnly(),
			"collector_data_access_role_arn": schema.StringAttribute{Computed: true},
			names.AttrCreatedAt:              schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot(names.AttrID), path.MatchRoot(names.AttrName)),
				},
			},
			names.AttrName:          schema.StringAttribute{Optional: true, Computed: true},
			"patch_creation_policy": schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.PatchCreationPolicy](), Computed: true},
			"scan_frequency":        schema.StringAttribute{Computed: true},
			names.AttrStatus:        schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.EnvironmentStatus](), Computed: true},
			names.AttrStatusMessage: schema.StringAttribute{Computed: true},
			"updated_at":            schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true},
			"repository_source":     framework.DataSourceComputedListOfObjectAttribute[repositorySourceModel](ctx),
			"s3_source":             framework.DataSourceComputedListOfObjectAttribute[s3SourceModel](ctx),
		},
	}
}

func (d *environmentDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data environmentDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ContinuumClient(ctx)

	id := data.ID.ValueString()
	if id == "" {
		var err error
		id, err = findEnvironmentIDByName(ctx, conn, data.Name.ValueString())
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
			return
		}
	}

	output, err := findEnvironmentByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	// The resource model has the same shape, so reuse its flatten.
	var model environmentResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, model.flatten(ctx, output))
	if response.Diagnostics.HasError() {
		return
	}

	data.AllowedDomains = model.AllowedDomains
	data.ARN = model.ARN
	data.CollectorDataAccessRoleARN = model.CollectorDataAccessRoleARN
	data.CreatedAt = model.CreatedAt
	data.ID = model.ID
	data.Name = model.Name
	data.PatchCreationPolicy = model.PatchCreationPolicy

	data.RepositorySources = model.RepositorySources
	data.S3Sources = model.S3Sources
	data.ScanFrequency = model.ScanFrequency
	data.Status = model.Status
	data.StatusMessage = model.StatusMessage
	data.UpdatedAt = model.UpdatedAt

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func findEnvironmentIDByName(ctx context.Context, conn *continuum.Client, name string) (string, error) {
	pages := continuum.NewListEnvironmentsPaginator(conn, &continuum.ListEnvironmentsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return "", err
		}
		for _, v := range page.Items {
			if aws.ToString(v.Name) == name {
				return aws.ToString(v.EnvironmentId), nil
			}
		}
	}

	return "", &retry.NotFoundError{Message: "no Continuum environment named " + name}
}

type environmentDataSourceModel struct {
	framework.WithRegionModel
	AllowedDomains             fwtypes.SetOfString                                    `tfsdk:"allowed_domains"`
	ARN                        types.String                                           `tfsdk:"arn"`
	CollectorDataAccessRoleARN types.String                                           `tfsdk:"collector_data_access_role_arn"`
	CreatedAt                  timetypes.RFC3339                                      `tfsdk:"created_at"`
	ID                         types.String                                           `tfsdk:"id"`
	Name                       types.String                                           `tfsdk:"name"`
	PatchCreationPolicy        fwtypes.StringEnum[awstypes.PatchCreationPolicy]       `tfsdk:"patch_creation_policy"`
	RepositorySources          fwtypes.ListNestedObjectValueOf[repositorySourceModel] `tfsdk:"repository_source"`
	S3Sources                  fwtypes.ListNestedObjectValueOf[s3SourceModel]         `tfsdk:"s3_source"`
	ScanFrequency              types.String                                           `tfsdk:"scan_frequency"`
	Status                     fwtypes.StringEnum[awstypes.EnvironmentStatus]         `tfsdk:"status"`
	StatusMessage              types.String                                           `tfsdk:"status_message"`
	UpdatedAt                  timetypes.RFC3339                                      `tfsdk:"updated_at"`
}
