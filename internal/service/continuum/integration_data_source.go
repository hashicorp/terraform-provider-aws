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

// @FrameworkDataSource("aws_continuum_integration", name="Integration")
func newIntegrationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &integrationDataSource{}, nil
}

type integrationDataSource struct {
	framework.DataSourceWithModel[integrationDataSourceModel]
}

func (d *integrationDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN:       framework.ARNAttributeComputedOnly(),
			names.AttrBucket:    schema.StringAttribute{Computed: true},
			names.AttrCreatedAt: schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot(names.AttrID), path.MatchRoot(names.AttrName)),
				},
			},
			"instance_url":      schema.StringAttribute{Computed: true},
			"integration_type":  schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.IntegrationType](), Computed: true},
			names.AttrName:      schema.StringAttribute{Optional: true, Computed: true},
			"organization_name": schema.StringAttribute{Computed: true},
			names.AttrRoleARN:   schema.StringAttribute{Computed: true},
			"target_path":       schema.StringAttribute{Computed: true},
			"target_type":       schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.GitLabTargetType](), Computed: true},
			"updated_at":        schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true},
		},
	}
}

func (d *integrationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data integrationDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ContinuumClient(ctx)

	id := data.ID.ValueString()
	if id == "" {
		var err error
		id, err = findIntegrationIDByName(ctx, conn, data.Name.ValueString())
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.Name.String())
			return
		}
	}

	output, err := findIntegrationByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	data.ARN = types.StringPointerValue(output.IntegrationArn)
	data.CreatedAt = timetypes.NewRFC3339TimePointerValue(output.CreatedAt)
	data.ID = types.StringPointerValue(output.IntegrationId)
	data.IntegrationType = fwtypes.StringEnumValue(output.IntegrationType)
	data.Name = types.StringPointerValue(output.Name)
	data.UpdatedAt = timetypes.NewRFC3339TimePointerValue(output.UpdatedAt)

	switch v := output.ProviderDetails.(type) {
	case *awstypes.ProviderDetailsMemberS3:
		data.Bucket = types.StringPointerValue(v.Value.Bucket)
		data.RoleARN = types.StringPointerValue(v.Value.RoleArn)
	case *awstypes.ProviderDetailsMemberGitlab:
		data.InstanceURL = types.StringPointerValue(v.Value.InstanceUrl)
		data.TargetPath = types.StringPointerValue(v.Value.TargetPath)
		data.TargetType = fwtypes.StringEnumValue(v.Value.TargetType)
	case *awstypes.ProviderDetailsMemberGithub:
		data.OrganizationName = types.StringPointerValue(v.Value.OrganizationName)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func findIntegrationIDByName(ctx context.Context, conn *continuum.Client, name string) (string, error) {
	pages := continuum.NewListIntegrationsPaginator(conn, &continuum.ListIntegrationsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return "", err
		}
		for _, v := range page.Items {
			if aws.ToString(v.Name) == name {
				return aws.ToString(v.IntegrationId), nil
			}
		}
	}

	return "", &retry.NotFoundError{Message: "no Continuum integration named " + name}
}

type integrationDataSourceModel struct {
	framework.WithRegionModel
	ARN              types.String                                  `tfsdk:"arn"`
	Bucket           types.String                                  `tfsdk:"bucket"`
	CreatedAt        timetypes.RFC3339                             `tfsdk:"created_at"`
	ID               types.String                                  `tfsdk:"id"`
	InstanceURL      types.String                                  `tfsdk:"instance_url"`
	IntegrationType  fwtypes.StringEnum[awstypes.IntegrationType]  `tfsdk:"integration_type"`
	Name             types.String                                  `tfsdk:"name"`
	OrganizationName types.String                                  `tfsdk:"organization_name"`
	RoleARN          types.String                                  `tfsdk:"role_arn"`
	TargetPath       types.String                                  `tfsdk:"target_path"`
	TargetType       fwtypes.StringEnum[awstypes.GitLabTargetType] `tfsdk:"target_type"`
	UpdatedAt        timetypes.RFC3339                             `tfsdk:"updated_at"`
}
