// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
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
			"settings_all": schema.StringAttribute{
				CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
				Computed:   true,
			},
			"use_cognito_provided_values": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Bool{
					boolvalidator.ExactlyOneOf(
						path.MatchRoot("settings"),
						path.MatchRoot("use_cognito_provided_values"),
					),
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
			"asset": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[assetTypeModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeBetween(0, 40),
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
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"category": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AssetCategoryType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"color_mode": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ColorSchemeModeType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"extension": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.AssetExtensionType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrResourceID: schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
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

	// Additional fields.
	settings, diags := data.Settings.ToSmithyDocument(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	input.Settings = settings

	output, err := conn.CreateManagedLoginBranding(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Cognito Managed Login Branding (%s)", data.ClientID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	mlb := output.ManagedLoginBranding
	data.ManagedLoginBrandingID = fwflex.StringToFramework(ctx, mlb.ManagedLoginBrandingId)
	data.UseCognitoProvidedValues = fwflex.BoolValueToFramework(ctx, mlb.UseCognitoProvidedValues)

	userPoolID, managedLoginBrandingID := fwflex.StringValueFromFramework(ctx, data.UserPoolID), fwflex.StringValueFromFramework(ctx, data.ManagedLoginBrandingID)
	// Return all values.
	mlb, err = findManagedLoginBrandingByThreePartKey(ctx, conn, userPoolID, managedLoginBrandingID, true)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed Login Branding (%s)", managedLoginBrandingID), err.Error())

		return
	}

	settingsAll, diags := flattenManagedLoginBrandingSettings(ctx, mlb.Settings)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	data.SettingsAll = settingsAll

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *managedLoginBrandingResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data managedLoginBrandingResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	userPoolID, managedLoginBrandingID := fwflex.StringValueFromFramework(ctx, data.UserPoolID), fwflex.StringValueFromFramework(ctx, data.ManagedLoginBrandingID)
	// Return only customized values.
	mlb, err := findManagedLoginBrandingByThreePartKey(ctx, conn, userPoolID, managedLoginBrandingID, false)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed Login Branding (%s)", managedLoginBrandingID), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, mlb, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	settings, diags := flattenManagedLoginBrandingSettings(ctx, mlb.Settings)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Settings = settings

	// Return all values.
	mlb, err = findManagedLoginBrandingByThreePartKey(ctx, conn, userPoolID, managedLoginBrandingID, true)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed Login Branding (%s)", managedLoginBrandingID), err.Error())

		return
	}

	settingsAll, diags := flattenManagedLoginBrandingSettings(ctx, mlb.Settings)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	data.SettingsAll = settingsAll

	input := cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolID),
	}
	pages := cognitoidentityprovider.NewListUserPoolClientsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("listing Cognito User Pool (%s) Clients", userPoolID), err.Error())

			return
		}

		for _, v := range page.UserPoolClients {
			clientID := aws.ToString(v.ClientId)
			input := cognitoidentityprovider.DescribeManagedLoginBrandingByClientInput{
				ClientId:   aws.String(clientID),
				UserPoolId: aws.String(userPoolID),
			}
			mlb, err := findManagedLoginBrandingByClient(ctx, conn, &input)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed Login Branding by client (%s)", clientID), err.Error())

				return
			}

			if aws.ToString(mlb.ManagedLoginBrandingId) == managedLoginBrandingID {
				data.ClientID = fwflex.StringValueToFramework(ctx, clientID)
			}
		}
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

	userPoolID, managedLoginBrandingID := fwflex.StringValueFromFramework(ctx, new.UserPoolID), fwflex.StringValueFromFramework(ctx, new.ManagedLoginBrandingID)
	var input cognitoidentityprovider.UpdateManagedLoginBrandingInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	if oldSettings, newSettings := fwflex.StringValueFromFramework(ctx, old.Settings), fwflex.StringValueFromFramework(ctx, new.Settings); newSettings != oldSettings && newSettings != "" {
		var err error
		input.Settings, err = tfsmithy.DocumentFromJSONString(newSettings, document.NewLazyDocument)

		if err != nil {
			response.Diagnostics.AddError("creating Smithy document", err.Error())

			return
		}

		input.UseCognitoProvidedValues = false
	}

	_, err := conn.UpdateManagedLoginBranding(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Cognito Managed Login Branding (%s)", managedLoginBrandingID), err.Error())

		return
	}

	// Return all values.
	mlb, err := findManagedLoginBrandingByThreePartKey(ctx, conn, userPoolID, managedLoginBrandingID, true)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed Login Branding (%s)", managedLoginBrandingID), err.Error())

		return
	}

	settingsAll, diags := flattenManagedLoginBrandingSettings(ctx, mlb.Settings)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	new.SettingsAll = settingsAll
	new.UseCognitoProvidedValues = fwflex.BoolValueToFramework(ctx, mlb.UseCognitoProvidedValues)

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *managedLoginBrandingResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data managedLoginBrandingResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	userPoolID, managedLoginBrandingID := fwflex.StringValueFromFramework(ctx, data.UserPoolID), fwflex.StringValueFromFramework(ctx, data.ManagedLoginBrandingID)
	tflog.Debug(ctx, "deleting Cognito Managed Login Branding", map[string]any{
		"managed_login_branding_id": managedLoginBrandingID,
		names.AttrUserPoolID:        userPoolID,
	})
	input := cognitoidentityprovider.DeleteManagedLoginBrandingInput{
		ManagedLoginBrandingId: aws.String(managedLoginBrandingID),
		UserPoolId:             aws.String(userPoolID),
	}
	_, err := conn.DeleteManagedLoginBranding(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Cognito Managed Login Branding (%s)", managedLoginBrandingID), err.Error())

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

func findManagedLoginBrandingByThreePartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, managedLoginBrandingID string, returnMergedResources bool) (*awstypes.ManagedLoginBrandingType, error) {
	input := cognitoidentityprovider.DescribeManagedLoginBrandingInput{
		ManagedLoginBrandingId: aws.String(managedLoginBrandingID),
		ReturnMergedResources:  returnMergedResources,
		UserPoolId:             aws.String(userPoolID),
	}

	return findManagedLoginBranding(ctx, conn, &input)
}

func findManagedLoginBranding(ctx context.Context, conn *cognitoidentityprovider.Client, input *cognitoidentityprovider.DescribeManagedLoginBrandingInput) (*awstypes.ManagedLoginBrandingType, error) {
	output, err := conn.DescribeManagedLoginBranding(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ManagedLoginBranding == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ManagedLoginBranding, nil
}

func findManagedLoginBrandingByClient(ctx context.Context, conn *cognitoidentityprovider.Client, input *cognitoidentityprovider.DescribeManagedLoginBrandingByClientInput) (*awstypes.ManagedLoginBrandingType, error) {
	output, err := conn.DescribeManagedLoginBrandingByClient(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ManagedLoginBranding == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.ManagedLoginBranding, nil
}

type managedLoginBrandingResourceModel struct {
	framework.WithRegionModel
	Asset                    fwtypes.SetNestedObjectValueOf[assetTypeModel] `tfsdk:"asset"`
	ClientID                 types.String                                   `tfsdk:"client_id"`
	ManagedLoginBrandingID   types.String                                   `tfsdk:"managed_login_branding_id"`
	Settings                 fwtypes.SmithyJSON[document.Interface]         `tfsdk:"settings" autoflex:"-"`
	SettingsAll              fwtypes.SmithyJSON[document.Interface]         `tfsdk:"settings_all" autoflex:"-"`
	UseCognitoProvidedValues types.Bool                                     `tfsdk:"use_cognito_provided_values"`
	UserPoolID               types.String                                   `tfsdk:"user_pool_id"`
}

func flattenManagedLoginBrandingSettings(ctx context.Context, settings document.Interface) (fwtypes.SmithyJSON[document.Interface], diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	var diags diag.Diagnostics

	if settings == nil {
		return fwtypes.NewSmithyJSONNull[document.Interface](), diags
	}

	value, err := tfsmithy.DocumentToJSONString(settings)

	if err != nil {
		diags.AddError("reading Smithy document", err.Error())

		return fwtypes.NewSmithyJSONNull[document.Interface](), diags
	}

	settings, d := fwtypes.NewSmithyJSONValue(value, document.NewLazyDocument).ToSmithyDocument(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return fwtypes.NewSmithyJSONNull[document.Interface](), diags
	}

	value, err = tfsmithy.DocumentToJSONString(settings)

	if err != nil {
		diags.AddError("reading Smithy document", err.Error())

		return fwtypes.NewSmithyJSONNull[document.Interface](), diags
	}

	return fwtypes.NewSmithyJSONValue(value, document.NewLazyDocument), diags
}

type assetTypeModel struct {
	Bytes      types.String                                     `tfsdk:"bytes"`
	Category   fwtypes.StringEnum[awstypes.AssetCategoryType]   `tfsdk:"category"`
	ColorMode  fwtypes.StringEnum[awstypes.ColorSchemeModeType] `tfsdk:"color_mode"`
	Extension  fwtypes.StringEnum[awstypes.AssetExtensionType]  `tfsdk:"extension"`
	ResourceID types.String                                     `tfsdk:"resource_id"`
}

var (
	_ fwflex.Expander  = assetTypeModel{}
	_ fwflex.Flattener = &assetTypeModel{}
)

func (m assetTypeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	r := awstypes.AssetType{
		Category:   m.Category.ValueEnum(),
		ColorMode:  m.ColorMode.ValueEnum(),
		Extension:  m.Extension.ValueEnum(),
		ResourceId: fwflex.StringFromFramework(ctx, m.ResourceID),
	}

	if v, err := inttypes.Base64Decode(m.Bytes.ValueString()); err == nil {
		r.Bytes = v
	} else {
		diags.AddError(
			"decoding asset bytes",
			err.Error(),
		)

		return nil, diags
	}

	return &r, diags
}

func (m *assetTypeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := v.(type) {
	case awstypes.AssetType:
		m.Bytes = fwflex.StringValueToFramework(ctx, inttypes.Base64Encode(v.Bytes))
		m.Category = fwtypes.StringEnumValue(v.Category)
		m.ColorMode = fwtypes.StringEnumValue(v.ColorMode)
		m.Extension = fwtypes.StringEnumValue(v.Extension)
		m.ResourceID = fwflex.StringToFramework(ctx, v.ResourceId)
	default:
	}

	return diags
}
