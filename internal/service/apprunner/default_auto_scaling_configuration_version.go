// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Default AutoScaling Configuration Version")
func newResourceIndex(context.Context) (resource.ResourceWithConfigure, error) {
	r := &defaultAutoScalingConfigurationVersionResource{}

	return r, nil
}

type defaultAutoScalingConfigurationVersionResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *defaultAutoScalingConfigurationVersionResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_apprunner_default_auto_scaling_configuration_version"
}

func (r *defaultAutoScalingConfigurationVersionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auto_scaling_configuration_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *defaultAutoScalingConfigurationVersionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data defaultAutoScalingConfigurationVersionResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppRunnerClient(ctx)

	if err := putDefaultAutoScalingConfiguration(ctx, conn, data.AutoScalingConfigurationARN.ValueString()); err != nil {
		response.Diagnostics.AddError("creating App Runner Default AutoScaling Configuration Version", err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *defaultAutoScalingConfigurationVersionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data defaultAutoScalingConfigurationVersionResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppRunnerClient(ctx)

	output, err := findDefaultAutoScalingConfigurationSummary(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading App Runner Default AutoScaling Configuration Version (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.AutoScalingConfigurationARN = fwtypes.ARNValue(aws.ToString(output.AutoScalingConfigurationArn))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *defaultAutoScalingConfigurationVersionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new defaultAutoScalingConfigurationVersionResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AppRunnerClient(ctx)

	if err := putDefaultAutoScalingConfiguration(ctx, conn, new.AutoScalingConfigurationARN.ValueString()); err != nil {
		response.Diagnostics.AddError("updating App Runner Default AutoScaling Configuration Version", err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (*defaultAutoScalingConfigurationVersionResource) Delete(context.Context, resource.DeleteRequest, *resource.DeleteResponse) {
	// NoOp.
}

func findDefaultAutoScalingConfigurationSummary(ctx context.Context, conn *apprunner.Client) (*awstypes.AutoScalingConfigurationSummary, error) {
	input := &apprunner.ListAutoScalingConfigurationsInput{}

	return findAutoScalingConfigurationSummary(ctx, conn, input, func(v *awstypes.AutoScalingConfigurationSummary) bool {
		return string(v.Status) == autoScalingConfigurationStatusActive && aws.ToBool(v.IsDefault)
	})
}

func putDefaultAutoScalingConfiguration(ctx context.Context, conn *apprunner.Client, arn string) error {
	input := &apprunner.UpdateDefaultAutoScalingConfigurationInput{
		AutoScalingConfigurationArn: aws.String(arn),
	}

	_, err := conn.UpdateDefaultAutoScalingConfiguration(ctx, input)

	if err != nil {
		return fmt.Errorf("putting App Runner AutoScaling Configuration Version (%s) as the default: %w", arn, err)
	}

	return nil
}

type defaultAutoScalingConfigurationVersionResourceModel struct {
	AutoScalingConfigurationARN fwtypes.ARN  `tfsdk:"auto_scaling_configuration_arn"`
	ID                          types.String `tfsdk:"id"`
}
