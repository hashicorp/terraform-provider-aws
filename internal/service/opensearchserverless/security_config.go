// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearchserverless_security_config", name="Security Config")
func newResourceSecurityConfig(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceSecurityConfig{}, nil
}

const (
	ResNameSecurityConfig = "Security Config"
)

type resourceSecurityConfig struct {
	framework.ResourceWithConfigure
}

func (r *resourceSecurityConfig) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"config_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SecurityConfigType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"saml_options": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				Attributes: map[string]schema.Attribute{
					"group_attribute": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 2048),
						},
					},
					"metadata": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 20480),
						},
					},
					"session_timeout": schema.Int64Attribute{
						Optional: true,
						Computed: true,
						Validators: []validator.Int64{
							int64validator.Between(5, 1540),
						},
					},
					"user_attribute": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 2048),
						},
					},
				},
			},
		},
	}
}

func (r *resourceSecurityConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)
	var plan resourceSecurityConfigData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := opensearchserverless.CreateSecurityConfigInput{}
	ignoreFieldOption := fwflex.WithIgnoredFieldNamesAppend("SamlOptions")
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, ignoreFieldOption)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(sdkid.UniqueId())
	input.SamlOptions = expandSAMLOptions(ctx, plan.SamlOptions, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateSecurityConfig(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameSecurityConfig, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}

	if out == nil || out.SecurityConfigDetail == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionCreating, ResNameSecurityConfig, plan.Name.String(), nil),
			"Empty response.",
		)
		return
	}

	state := plan
	resp.Diagnostics.Append(fwflex.Flatten(ctx, out.SecurityConfigDetail, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SamlOptions = flattenSAMLOptions(ctx, out.SecurityConfigDetail.SamlOptions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSecurityConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceSecurityConfigData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSecurityConfigByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, ResNameSecurityConfig, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state, fwflex.WithIgnoredFieldNamesAppend("SamlOptions"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.SamlOptions = flattenSAMLOptions(ctx, out.SamlOptions)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSecurityConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan, state resourceSecurityConfigData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, diags := fwflex.Diff(ctx, plan, state)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	if diff.HasChanges() {
		input := opensearchserverless.UpdateSecurityConfigInput{}
		flexIgnoreOption := fwflex.WithIgnoredFieldNamesAppend("SamlOptions")
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, flexIgnoreOption)...)
		if resp.Diagnostics.HasError() {
			return
		}

		input.ClientToken = aws.String(sdkid.UniqueId())
		input.ConfigVersion = state.ConfigVersion.ValueStringPointer()
		input.SamlOptions = expandSAMLOptions(ctx, plan.SamlOptions, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateSecurityConfig(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating Security Policy (%s)", plan.Name.ValueString()), err.Error())
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, out.SecurityConfigDetail, &plan, flexIgnoreOption)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSecurityConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state resourceSecurityConfigData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteSecurityConfig(ctx, &opensearchserverless.DeleteSecurityConfigInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Id:          state.ID.ValueStringPointer(),
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionDeleting, ResNameSecurityConfig, state.Name.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceSecurityConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, idSeparator)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		err := fmt.Errorf("unexpected format for ID (%[1]s), expected saml/account-id/name", req.ID)
		resp.Diagnostics.AddError(fmt.Sprintf("importing Security Policy (%s)", req.ID), err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrID), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrName), parts[2])...)
}

type resourceSecurityConfigData struct {
	ID            types.String                                    `tfsdk:"id"`
	ConfigVersion types.String                                    `tfsdk:"config_version"`
	Description   types.String                                    `tfsdk:"description"`
	Name          types.String                                    `tfsdk:"name"`
	SamlOptions   types.Object                                    `tfsdk:"saml_options"`
	Type          fwtypes.StringEnum[awstypes.SecurityConfigType] `tfsdk:"type"`
}

type samlOptions struct {
	GroupAttribute types.String `tfsdk:"group_attribute"`
	Metadata       types.String `tfsdk:"metadata"`
	SessionTimeout types.Int64  `tfsdk:"session_timeout"`
	UserAttribute  types.String `tfsdk:"user_attribute"`
}

func (so *samlOptions) expand(_ context.Context) *awstypes.SamlConfigOptions {
	if so == nil {
		return nil
	}

	result := &awstypes.SamlConfigOptions{
		Metadata:       so.Metadata.ValueStringPointer(),
		GroupAttribute: so.GroupAttribute.ValueStringPointer(),
		UserAttribute:  so.UserAttribute.ValueStringPointer(),
	}

	if so.SessionTimeout.ValueInt64() != 0 {
		result.SessionTimeout = aws.Int32(int32(so.SessionTimeout.ValueInt64()))
	}

	return result
}

func expandSAMLOptions(ctx context.Context, object types.Object, diags *diag.Diagnostics) *awstypes.SamlConfigOptions {
	var options samlOptions
	diags.Append(object.As(ctx, &options, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return options.expand(ctx)
}

func flattenSAMLOptions(ctx context.Context, so *awstypes.SamlConfigOptions) types.Object {
	if so == nil {
		return fwtypes.NewObjectValueOfNull[samlOptions](ctx).ObjectValue
	}

	attributeTypes := fwtypes.AttributeTypesMust[samlOptions](ctx)
	attrs := map[string]attr.Value{}
	attrs["group_attribute"] = fwflex.StringToFramework(ctx, so.GroupAttribute)
	attrs["metadata"] = fwflex.StringToFramework(ctx, so.Metadata)
	timeout := int64(*so.SessionTimeout)
	attrs["session_timeout"] = fwflex.Int64ToFramework(ctx, &timeout)
	attrs["user_attribute"] = fwflex.StringToFramework(ctx, so.UserAttribute)

	return types.ObjectValueMust(attributeTypes, attrs)
}
