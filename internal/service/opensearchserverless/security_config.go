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
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
func newSecurityConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &securityConfigResource{}, nil
}

const (
	ResNameSecurityConfig = "Security Config"
)

type securityConfigResource struct {
	framework.ResourceWithModel[securityConfigResourceModel]
}

func (r *securityConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"config_version": schema.StringAttribute{
				Description: "Version of the configuration.",
				Computed:    true,
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the security configuration.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
			names.AttrName: schema.StringAttribute{
				Description: "Name of the policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			names.AttrType: schema.StringAttribute{
				Description: "Type of configuration. Must be `saml`.",
				CustomType:  fwtypes.StringEnumType[awstypes.SecurityConfigType](),
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"saml_options": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[samlOptionsData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_attribute": schema.StringAttribute{
							Description: "Group attribute for this SAML integration.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
						"metadata": schema.StringAttribute{
							Description: "The XML IdP metadata file generated from your identity provider.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 20480),
							},
						},
						"session_timeout": schema.Int64Attribute{
							Description: "Session timeout, in minutes. Minimum is 5 minutes and maximum is 720 minutes (12 hours). Default is 60 minutes.",
							Optional:    true,
							Computed:    true,
							Validators: []validator.Int64{
								int64validator.Between(5, 1540),
							},
						},
						"user_attribute": schema.StringAttribute{
							Description: "User attribute for this SAML integration.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 2048),
							},
						},
					},
				},
			},
		},
	}
}

func (r *securityConfigResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := securityConfigSchemaV0()

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeSecurityConfigStateFromV0,
		},
	}
}

func (r *securityConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)
	var plan securityConfigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := opensearchserverless.CreateSecurityConfigInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(sdkid.UniqueId())

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

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out.SecurityConfigDetail, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state securityConfigResourceModel
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

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var plan, state securityConfigResourceModel
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
		resp.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.ClientToken = aws.String(sdkid.UniqueId())
		input.ConfigVersion = state.ConfigVersion.ValueStringPointer()

		out, err := conn.UpdateSecurityConfig(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating Security Policy (%s)", plan.Name.ValueString()), err.Error())
			return
		}

		resp.Diagnostics.Append(fwflex.Flatten(ctx, out.SecurityConfigDetail, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().OpenSearchServerlessClient(ctx)

	var state securityConfigResourceModel
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

func (r *securityConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, idSeparator)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		err := fmt.Errorf("unexpected format for ID (%[1]s), expected saml/account-id/name", req.ID)
		resp.Diagnostics.AddError(fmt.Sprintf("importing Security Policy (%s)", req.ID), err.Error())
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(names.AttrName), parts[2])...)
}

type securityConfigResourceModel struct {
	framework.WithRegionModel
	ID            types.String                                     `tfsdk:"id"`
	ConfigVersion types.String                                     `tfsdk:"config_version"`
	Description   types.String                                     `tfsdk:"description"`
	Name          types.String                                     `tfsdk:"name"`
	SamlOptions   fwtypes.ListNestedObjectValueOf[samlOptionsData] `tfsdk:"saml_options"`
	Type          fwtypes.StringEnum[awstypes.SecurityConfigType]  `tfsdk:"type"`
}

type samlOptionsData struct {
	GroupAttribute types.String `tfsdk:"group_attribute"`
	Metadata       types.String `tfsdk:"metadata"`
	SessionTimeout types.Int64  `tfsdk:"session_timeout"`
	UserAttribute  types.String `tfsdk:"user_attribute"`
}
