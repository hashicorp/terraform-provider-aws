// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Retention Configuration")
func newRetentionConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &retentionConfigurationResource{}, nil
}

type retentionConfigurationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *retentionConfigurationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_config_retention_configuration"
}

func (r *retentionConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"retention_period_in_days": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(30, 2557),
				},
			},
		},
	}
}

func (r *retentionConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data retentionConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConfigServiceClient(ctx)

	input := &configservice.PutRetentionConfigurationInput{
		RetentionPeriodInDays: fwflex.Int32FromFramework(ctx, data.RetentionPeriodInDays),
	}

	output, err := conn.PutRetentionConfiguration(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating ConfigService Retention Configuration", err.Error())

		return
	}

	// Set values for unknowns.
	data.Name = fwflex.StringToFramework(ctx, output.RetentionConfiguration.Name)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *retentionConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data retentionConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().ConfigServiceClient(ctx)

	name := data.ID.ValueString()
	retentionConfiguration, err := findRetentionConfigurationByName(ctx, conn, name)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ConfigService Retention Configuration (%s)", name), err.Error())

		return
	}

	data.RetentionPeriodInDays = fwflex.Int32ToFramework(ctx, retentionConfiguration.RetentionPeriodInDays)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *retentionConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new retentionConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConfigServiceClient(ctx)

	input := &configservice.PutRetentionConfigurationInput{
		RetentionPeriodInDays: fwflex.Int32FromFramework(ctx, new.RetentionPeriodInDays),
	}

	_, err := conn.PutRetentionConfiguration(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating ConfigService Retention Configuration (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *retentionConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data retentionConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ConfigServiceClient(ctx)

	name := data.ID.ValueString()
	_, err := conn.DeleteRetentionConfiguration(ctx, &configservice.DeleteRetentionConfigurationInput{
		RetentionConfigurationName: aws.String(name),
	})

	if errs.IsA[*awstypes.NoSuchRetentionConfigurationException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting ConfigService Retention Configuration (%s)", name), err.Error())

		return
	}
}

func findRetentionConfigurationByName(ctx context.Context, conn *configservice.Client, name string) (*awstypes.RetentionConfiguration, error) {
	input := &configservice.DescribeRetentionConfigurationsInput{
		RetentionConfigurationNames: []string{name},
	}

	return findRetentionConfiguration(ctx, conn, input)
}

func findRetentionConfiguration(ctx context.Context, conn *configservice.Client, input *configservice.DescribeRetentionConfigurationsInput) (*awstypes.RetentionConfiguration, error) {
	output, err := findRetentionConfigurations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findRetentionConfigurations(ctx context.Context, conn *configservice.Client, input *configservice.DescribeRetentionConfigurationsInput) ([]awstypes.RetentionConfiguration, error) {
	var output []awstypes.RetentionConfiguration

	pages := configservice.NewDescribeRetentionConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchRetentionConfigurationException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.RetentionConfigurations...)
	}

	return output, nil
}

type retentionConfigurationResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	RetentionPeriodInDays types.Int64  `tfsdk:"retention_period_in_days"`
}

func (model *retentionConfigurationResourceModel) InitFromID() error {
	model.Name = model.ID

	return nil
}

func (model *retentionConfigurationResourceModel) setID() {
	model.ID = model.Name
}
