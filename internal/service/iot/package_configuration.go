// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iot/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_iot_package_configuration", name="Package Configuration")
func newResourcePackageConfiguration(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourcePackageConfiguration{}, nil
}

const (
	ResNamePackageConfiguration = "Package Configuration"
)

type resourcePackageConfiguration struct {
	framework.ResourceWithConfigure
}

func (r *resourcePackageConfiguration) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_iot_package_configuration"
}

func (r *resourcePackageConfiguration) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Blocks: map[string]schema.Block{
			"version_update_by_jobs": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[versionUpdateByJobsConfig](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrEnabled: schema.BoolAttribute{
							Required: true,
						},
						names.AttrRoleARN: schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourcePackageConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourcePackageConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IoTClient(ctx)

	input := &iot.UpdatePackageConfigurationInput{
		VersionUpdateByJobsConfig: &awstypes.VersionUpdateByJobsConfig{},
	}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdatePackageConfiguration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionCreating, ResNamePackageConfiguration, "", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionCreating, ResNamePackageConfiguration, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePackageConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourcePackageConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IoTClient(ctx)

	input := &iot.UpdatePackageConfigurationInput{
		VersionUpdateByJobsConfig: &awstypes.VersionUpdateByJobsConfig{},
	}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdatePackageConfiguration(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionCreating, ResNamePackageConfiguration, "", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionCreating, ResNamePackageConfiguration, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePackageConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePackageConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().IoTClient(ctx)

	out, err := conn.GetPackageConfiguration(ctx, &iot.GetPackageConfigurationInput{})
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionSetting, ResNamePackageConfiguration, "", err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePackageConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().IoTClient(ctx)

	input := &iot.UpdatePackageConfigurationInput{
		VersionUpdateByJobsConfig: &awstypes.VersionUpdateByJobsConfig{
			Enabled: aws.Bool(false),
			RoleArn: nil,
		},
	}

	_, err := conn.UpdatePackageConfiguration(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IoT, create.ErrActionDeleting, ResNamePackageConfiguration, "", err),
			err.Error(),
		)
		return
	}
}

type resourcePackageConfigurationModel struct {
	VersionUpdateByJobsConfig fwtypes.ListNestedObjectValueOf[versionUpdateByJobsConfig] `tfsdk:"version_update_by_jobs"`
}

type versionUpdateByJobsConfig struct {
	Enabled types.Bool   `tfsdk:"enabled"`
	RoleARN types.String `tfsdk:"role_arn"`
}
