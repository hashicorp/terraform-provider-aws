// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cognito_managed_login_branding", name="Managed Login Branding")
func newManagedLoginBrandingResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &managedLoginBrandingResource{}

	return r, nil
}

type managedLoginBrandingResource struct {
	framework.ResourceWithModel[managedLoginBrandingResourceModel]
}

func (r *managedLoginBrandingResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrClientID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"managed_login_branding_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"settings": schema.StringAttribute{
				CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Optional:   true,
			},
			"use_cognito_provided_values": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Bool{
					boolvalidator.ExactlyOneOf(
						path.MatchRoot("settings"),
					),
					boolvalidator.ConflictsWith(
						path.MatchRoot("asset"),
						path.MatchRoot("settings"),
					),
				},
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"asset": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[assetTypeModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 40),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bytes": schema.StringAttribute{
							Optional: true,
						},
						"category": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AssetCategoryType](),
							Required:   true,
						},
						"color_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ColorSchemeModeType](),
							Required:   true,
						},
						"extension": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AssetExtensionType](),
							Required:   true,
						},
						"resource_id": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *managedLoginBrandingResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data managedLoginBrandingResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	var input cognitoidentityprovider.CreateManagedLoginBrandingInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateManagedLoginBranding(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Cognito Managed Login Branding (%s)", data.ClientID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	mlb := output.ManagedLoginBranding
	data.ManagedLoginBrandingID = fwflex.StringToFramework(ctx, mlb.ManagedLoginBrandingId)
	data.UseCognitoProvidedValues = fwflex.BoolValueToFramework(ctx, mlb.UseCognitoProvidedValues)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *managedLoginBrandingResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data managedLoginBrandingResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	mlb, err := findManagedLoginBrandingByTwoPartKey(ctx, conn, data.UserPoolID.ValueString(), data.ManagedLoginBrandingID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed Login Branding (%s)", data.ManagedLoginBrandingID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, mlb, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *managedLoginBrandingResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new managedLoginBrandingResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	var input cognitoidentityprovider.UpdateManagedLoginBrandingInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateManagedLoginBranding(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Cognito Managed Login Branding (%s)", new.ManagedLoginBrandingID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *managedLoginBrandingResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data managedLoginBrandingResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	tflog.Debug(ctx, "deleting Cognito Managed Login Branding", map[string]any{
		"managed_login_branding_id": data.ManagedLoginBrandingID.ValueString(),
		names.AttrUserPoolID:        data.UserPoolID.ValueString(),
	})
	input := cognitoidentityprovider.DeleteManagedLoginBrandingInput{
		ManagedLoginBrandingId: fwflex.StringFromFramework(ctx, data.ManagedLoginBrandingID),
		UserPoolId:             fwflex.StringFromFramework(ctx, data.UserPoolID),
	}
	_, err := conn.DeleteManagedLoginBranding(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Cognito Managed Login Branding (%s)", data.ManagedLoginBrandingID.ValueString()), err.Error())

		return
	}
}

func (r *managedLoginBrandingResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		managedLoginBrandingIDParts = 2
	)
	parts, err := intflex.ExpandResourceId(request.ID, managedLoginBrandingIDParts, true)

	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrUserPoolID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("managed_login_branding_id"), parts[1])...)
}

func findManagedLoginBrandingByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, managedLoginBrandingID string) (*awstypes.ManagedLoginBrandingType, error) {
	input := cognitoidentityprovider.DescribeManagedLoginBrandingInput{
		ManagedLoginBrandingId: aws.String(managedLoginBrandingID),
		ReturnMergedResources:  false, // Return only customized values.
		UserPoolId:             aws.String(userPoolID),
	}

	output, err := conn.DescribeManagedLoginBranding(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ManagedLoginBranding == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ManagedLoginBranding, nil
}

type managedLoginBrandingResourceModel struct {
	framework.WithRegionModel
	Asset                    fwtypes.ListNestedObjectValueOf[assetTypeModel] `tfsdk:"asset"`
	ClientID                 types.String                                    `tfsdk:"client_id"`
	ManagedLoginBrandingID   types.String                                    `tfsdk:"managed_login_branding_id"`
	Settings                 fwtypes.SmithyJSON[document.Interface]          `tfsdk:"settings"`
	UseCognitoProvidedValues types.Bool                                      `tfsdk:"use_cognito_provided_values"`
	UserPoolID               types.String                                    `tfsdk:"user_pool_id"`
}

type assetTypeModel struct {
	Bytes      types.String                                     `tfsdk:"bytes"`
	Category   fwtypes.StringEnum[awstypes.AssetCategoryType]   `tfsdk:"category"`
	ColorMode  fwtypes.StringEnum[awstypes.ColorSchemeModeType] `tfsdk:"color_mode"`
	Extension  fwtypes.StringEnum[awstypes.AssetExtensionType]  `tfsdk:"extension"`
	ResourceID types.String                                     `tfsdk:"resource_id"`
}
