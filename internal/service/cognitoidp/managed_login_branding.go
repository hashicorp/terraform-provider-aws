// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cognito_managed_login_branding", name="Managed Login Branding")
func newResourceManagedLoginBranding(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceManagedLoginBranding{}

	return r, nil
}

const (
	ResNameManagedLoginBranding = "Managed Login Branding"
)

type resourceManagedLoginBranding struct {
	framework.ResourceWithConfigure
}

func (r *resourceManagedLoginBranding) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrClientID: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\w+]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"creation_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"managed_login_branding_id": schema.StringAttribute{
				Computed: true,
			},
			"settings": schema.StringAttribute{
				CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("use_cognito_provided_values"),
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"use_cognito_provided_values": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"user_pool_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 55),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\w-]+_[0-9a-zA-Z]+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"asset": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[assetModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeAtMost(40),
				},
				PlanModifiers: []planmodifier.Set{
					// The update API allows updating an asset.
					// However, if the (`category`, `color`) pair differs from existing ones,
					// the API treats the asset as new and adds it accordingly.
					// This can result in a mismatch between the Terraform plan and the actual state
					// (e.g., the plan contains one asset, but the state contains two or more).
					// To preserve declarative behavior, the resource is replaced whenever the `asset` is modified.
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bytes": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1000000),
							},
						},
						"category": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.AssetCategoryType](),
							},
						},
						"color_mode": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.ColorSchemeModeType](),
							},
						},
						"extension": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								enum.FrameworkValidate[awstypes.AssetExtensionType](),
							},
						},
						"resource_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 40),
								stringvalidator.RegexMatches(regexache.MustCompile(`^[\w\- ]+$`), ""),
							},
						},
					},
				},
			},
		},
	}
}

func managedLoginBrandingCalcID(ctx context.Context, managedLoginBrandingId, userPoolId, clientId types.String) types.String {
	return flex.StringToFramework(ctx, aws.String(managedLoginBrandingId.ValueString()+"|"+userPoolId.ValueString()+"|"+clientId.ValueString()))
}

func flattenManagedLoginBrandingSettings(settings document.Interface) (fwtypes.SmithyJSON[document.Interface], diag.Diagnostics) {
	var diags diag.Diagnostics

	if settings == nil {
		return fwtypes.SmithyJSON[document.Interface]{}, diags
	}

	// This code serializes the settings value (a Go value) obtained from the API response into a Smithy document JSON,
	// deserializes it back into a Go value, and then serializes it again.
	// This process ensures that the data is normalized in the same way as JSON values defined in configuration files.
	bytes, err := document.Interface.MarshalSmithyDocument(settings)
	if err != nil {
		diags.AddError(
			"Failed to marshal settings",
			err.Error(),
		)
		return fwtypes.SmithyJSON[document.Interface]{}, diags
	}
	r, diags := fwtypes.SmithyJSONValue[document.Interface](string(bytes), document.NewLazyDocument).ValueInterface()
	if diags.HasError() {
		diags.AddError(
			"Failed to get value interface from SmithyJSON",
			"",
		)
		return fwtypes.SmithyJSON[document.Interface]{}, diags
	}
	bytes2, err := document.Interface.MarshalSmithyDocument(r)
	if err != nil {
		diags.AddError(
			"Failed to marshal value interface",
			err.Error(),
		)
		return fwtypes.SmithyJSON[document.Interface]{}, diags
	}
	return fwtypes.SmithyJSONValue[document.Interface](string(bytes2), document.NewLazyDocument), diags
}

func (r *resourceManagedLoginBranding) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan resourceManagedLoginBrandingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input cognitoidentityprovider.CreateManagedLoginBrandingInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if v, err := plan.Settings.ValueInterface(); err == nil {
		input.Settings = v
	} else {
		resp.Diagnostics.AddError(
			"Failed to expand settings",
			"",
		)
		return
	}
	out, err := conn.CreateManagedLoginBranding(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionCreating, ResNameManagedLoginBranding, plan.UserPoolId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ManagedLoginBranding == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionCreating, ResNameManagedLoginBranding, plan.UserPoolId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out.ManagedLoginBranding, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = managedLoginBrandingCalcID(ctx, plan.ManagedLoginBrandingId, plan.UserPoolId, plan.ClientId)
	flattenedSettings, diags := flattenManagedLoginBrandingSettings(out.ManagedLoginBranding.Settings)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	plan.Settings = flattenedSettings
	if !plan.Settings.IsNull() {
		plan.UseCognitoProvidedValues = types.BoolPointerValue(nil)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)

}

func (r *resourceManagedLoginBranding) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state resourceManagedLoginBrandingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findManagedLoginBrandingByID(ctx, conn, state.ManagedLoginBrandingId.ValueString(), state.UserPoolId.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionReading, ResNameManagedLoginBranding, state.ManagedLoginBrandingId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ID = managedLoginBrandingCalcID(ctx, state.ManagedLoginBrandingId, state.UserPoolId, state.ClientId)
	flattenedSettings, diags := flattenManagedLoginBrandingSettings(out.Settings)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	state.Settings = flattenedSettings

	if !state.Settings.IsNull() {
		state.UseCognitoProvidedValues = types.BoolPointerValue(nil)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceManagedLoginBranding) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan, state resourceManagedLoginBrandingModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := cognitoidentityprovider.UpdateManagedLoginBrandingInput{
			ManagedLoginBrandingId: state.ManagedLoginBrandingId.ValueStringPointer(),
		}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if v, err := plan.Settings.ValueInterface(); err == nil {
			input.Settings = v
		} else {
			resp.Diagnostics.AddError(
				"Failed to expand settings",
				"",
			)
			return
		}

		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateManagedLoginBranding(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionUpdating, ResNameManagedLoginBranding, plan.ManagedLoginBrandingId.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.ManagedLoginBranding == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionUpdating, ResNameManagedLoginBranding, plan.ManagedLoginBrandingId.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out.ManagedLoginBranding, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

		flattenedSettings, diags := flattenManagedLoginBrandingSettings(out.ManagedLoginBranding.Settings)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		state.Settings = flattenedSettings

	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceManagedLoginBranding) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state resourceManagedLoginBrandingModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := cognitoidentityprovider.DeleteManagedLoginBrandingInput{
		ManagedLoginBrandingId: state.ManagedLoginBrandingId.ValueStringPointer(),
		UserPoolId:             state.UserPoolId.ValueStringPointer(),
	}

	_, err := conn.DeleteManagedLoginBranding(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionDeleting, ResNameManagedLoginBranding, state.ManagedLoginBrandingId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceManagedLoginBranding) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "|")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: attr_one,attr_two. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("managed_login_branding_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_pool_id"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("client_id"), idParts[2])...)
}

func findManagedLoginBrandingByID(ctx context.Context, conn *cognitoidentityprovider.Client, id string, userPoolId string) (*awstypes.ManagedLoginBrandingType, error) {
	input := cognitoidentityprovider.DescribeManagedLoginBrandingInput{
		ManagedLoginBrandingId: aws.String(id),
		UserPoolId:             aws.String(userPoolId),
	}

	out, err := conn.DescribeManagedLoginBranding(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.ManagedLoginBranding == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.ManagedLoginBranding, nil
}

type resourceManagedLoginBrandingModel struct {
	Assets                   fwtypes.SetNestedObjectValueOf[assetModel] `tfsdk:"asset"`
	ClientId                 types.String                               `tfsdk:"client_id"`
	CreationDate             timetypes.RFC3339                          `tfsdk:"creation_date"`
	ID                       types.String                               `tfsdk:"id"`
	ManagedLoginBrandingId   types.String                               `tfsdk:"managed_login_branding_id"`
	LastModifiedDate         timetypes.RFC3339                          `tfsdk:"last_modified_date"`
	Settings                 fwtypes.SmithyJSON[document.Interface]     `tfsdk:"settings" autoflex:"-"`
	UseCognitoProvidedValues types.Bool                                 `tfsdk:"use_cognito_provided_values"`
	UserPoolId               types.String                               `tfsdk:"user_pool_id"`
}

type assetModel struct {
	Bytes      types.String                                     `tfsdk:"bytes"`
	Category   fwtypes.StringEnum[awstypes.AssetCategoryType]   `tfsdk:"category"`
	ColorMode  fwtypes.StringEnum[awstypes.ColorSchemeModeType] `tfsdk:"color_mode"`
	Extension  fwtypes.StringEnum[awstypes.AssetExtensionType]  `tfsdk:"extension"`
	ResourceID types.String                                     `tfsdk:"resource_id"`
}

var (
	_ flex.Expander  = assetModel{}
	_ flex.Flattener = &assetModel{}
)

func (m assetModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := awstypes.AssetType{}

	if v, err := itypes.Base64Decode(m.Bytes.ValueString()); err == nil {
		r.Bytes = v
	} else {
		diags.AddError(
			"Failed to decode asset bytes",
			err.Error(),
		)
		return nil, diags
	}
	r.Category = awstypes.AssetCategoryType(m.Category.ValueString())
	r.ColorMode = awstypes.ColorSchemeModeType(m.ColorMode.ValueString())
	r.Extension = awstypes.AssetExtensionType(m.Extension.ValueString())
	r.ResourceId = m.ResourceID.ValueStringPointer()
	return &r, diags
}

func (m *assetModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch t := v.(type) {
	case awstypes.AssetType:
		m.Bytes = flex.StringToFramework(ctx, aws.String(itypes.Base64Encode(t.Bytes)))
		m.Category = fwtypes.StringEnumValue[awstypes.AssetCategoryType](t.Category)
		m.ColorMode = fwtypes.StringEnumValue[awstypes.ColorSchemeModeType](t.ColorMode)
		m.Extension = fwtypes.StringEnumValue[awstypes.AssetExtensionType](t.Extension)
		m.ResourceID = flex.StringToFramework(ctx, t.ResourceId)
	default:
		diags.AddError(
			"Failed to flatten asset",
			fmt.Sprintf("Expected assetType, got %T", v),
		)
	}
	return diags
}
