// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceUserPoolClient(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceUserPoolClient{}
	r.SetMigratedFromPluginSDK(true)

	return r, nil
}

type resourceUserPoolClient struct {
	framework.ResourceWithConfigure
}

func (r *resourceUserPoolClient) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cognito_user_pool_client"
}

// Schema returns the schema for this resource.
func (r *resourceUserPoolClient) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_token_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"allowed_oauth_flows": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(3),
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(cognitoidentityprovider.OAuthFlowType_Values()...),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"allowed_oauth_flows_user_pool_client": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"allowed_oauth_scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(50),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"auth_session_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(3, 15),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"callback_urls": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(100),
					setvalidator.ValueStringsAre(
						userPoolClientURLValidator...,
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_redirect_uri": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: userPoolClientURLValidator,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_propagate_additional_user_context_data": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_token_revocation": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"explicit_auth_flows": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(cognitoidentityprovider.ExplicitAuthFlowsType_Values()...),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"generate_secret": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"id_token_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"logout_urls": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(100),
					setvalidator.ValueStringsAre(
						userPoolClientURLValidator...,
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:   true,
				Validators: userPoolClientNameValidator,
			},
			"prevent_user_existence_errors": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf(cognitoidentityprovider.PreventUserExistenceErrorTypes_Values()...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"read_attributes": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"refresh_token_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"supported_identity_providers": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						userPoolClientIdentityProviderValidator...,
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"user_pool_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"write_attributes": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"analytics_configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"application_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Validators: []validator.String{
								stringvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("application_arn"),
									path.MatchRelative().AtParent().AtName("application_id"),
								),
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName("external_id"),
									path.MatchRelative().AtParent().AtName("role_arn"),
								),
							},
						},
						"application_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("external_id"),
									path.MatchRelative().AtParent().AtName("role_arn"),
								),
							},
						},
						"external_id": schema.StringAttribute{
							Optional: true,
						},
						"role_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							Computed:   true,
						},
						"user_data_shared": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"token_validity_units": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"access_token": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(cognitoidentityprovider.TimeUnitsTypeHours),
							Validators: []validator.String{
								stringvalidator.OneOf(cognitoidentityprovider.TimeUnitsType_Values()...),
							},
						},
						"id_token": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(cognitoidentityprovider.TimeUnitsTypeHours),
							Validators: []validator.String{
								stringvalidator.OneOf(cognitoidentityprovider.TimeUnitsType_Values()...),
							},
						},
						"refresh_token": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  stringdefault.StaticString(cognitoidentityprovider.TimeUnitsTypeDays),
							Validators: []validator.String{
								stringvalidator.OneOf(cognitoidentityprovider.TimeUnitsType_Values()...),
							},
						},
					},
				},
			},
		},
	}

	response.Schema = s
}

func (r *resourceUserPoolClient) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPConn(ctx)

	var config resourceUserPoolClientData
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceUserPoolClientData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := plan.createInput(ctx, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := conn.CreateUserPoolClientWithContext(ctx, params)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("creating Cognito User Pool Client (%s)", plan.Name.ValueString()),
			err.Error(),
		)
		return
	}

	poolClient := resp.UserPoolClient

	config.AccessTokenValidity = flex.Int64ToFrameworkLegacy(ctx, poolClient.AccessTokenValidity)
	config.AllowedOauthFlows = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.AllowedOAuthFlows)
	config.AllowedOauthFlowsUserPoolClient = flex.BoolToFramework(ctx, poolClient.AllowedOAuthFlowsUserPoolClient)
	config.AllowedOauthScopes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.AllowedOAuthScopes)
	config.AnalyticsConfiguration = flattenAnaylticsConfiguration(ctx, poolClient.AnalyticsConfiguration)
	config.AuthSessionValidity = flex.Int64ToFramework(ctx, poolClient.AuthSessionValidity)
	config.CallbackUrls = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.CallbackURLs)
	config.ClientSecret = flex.StringToFrameworkLegacy(ctx, poolClient.ClientSecret)
	config.DefaultRedirectUri = flex.StringToFrameworkLegacy(ctx, poolClient.DefaultRedirectURI)
	config.EnablePropagateAdditionalUserContextData = flex.BoolToFramework(ctx, poolClient.EnablePropagateAdditionalUserContextData)
	config.EnableTokenRevocation = flex.BoolToFramework(ctx, poolClient.EnableTokenRevocation)
	config.ExplicitAuthFlows = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.ExplicitAuthFlows)
	config.ID = flex.StringToFramework(ctx, poolClient.ClientId)
	config.IdTokenValidity = flex.Int64ToFrameworkLegacy(ctx, poolClient.IdTokenValidity)
	config.LogoutUrls = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.LogoutURLs)
	config.Name = flex.StringToFramework(ctx, poolClient.ClientName)
	config.PreventUserExistenceErrors = flex.StringToFrameworkLegacy(ctx, poolClient.PreventUserExistenceErrors)
	config.ReadAttributes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.ReadAttributes)
	config.RefreshTokenValidity = flex.Int64ToFramework(ctx, poolClient.RefreshTokenValidity)
	config.SupportedIdentityProviders = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.SupportedIdentityProviders)
	config.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits)
	config.UserPoolID = flex.StringToFramework(ctx, poolClient.UserPoolId)
	config.WriteAttributes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.WriteAttributes)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &config)...)
}

func (r *resourceUserPoolClient) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resourceUserPoolClientData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPConn(ctx)

	poolClient, err := FindCognitoUserPoolClientByID(ctx, conn, state.UserPoolID.ValueString(), state.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolClient, state.ID.ValueString())
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.Append(create.DiagErrorFramework(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolClient, state.ID.ValueString(), err))
		return
	}

	state.AccessTokenValidity = flex.Int64ToFrameworkLegacy(ctx, poolClient.AccessTokenValidity)
	state.AllowedOauthFlows = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.AllowedOAuthFlows)
	state.AllowedOauthFlowsUserPoolClient = flex.BoolToFramework(ctx, poolClient.AllowedOAuthFlowsUserPoolClient)
	state.AllowedOauthScopes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.AllowedOAuthScopes)
	state.AnalyticsConfiguration = flattenAnaylticsConfiguration(ctx, poolClient.AnalyticsConfiguration)
	state.AuthSessionValidity = flex.Int64ToFramework(ctx, poolClient.AuthSessionValidity)
	state.CallbackUrls = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.CallbackURLs)
	state.ClientSecret = flex.StringToFrameworkLegacy(ctx, poolClient.ClientSecret)
	state.DefaultRedirectUri = flex.StringToFrameworkLegacy(ctx, poolClient.DefaultRedirectURI)
	state.EnablePropagateAdditionalUserContextData = flex.BoolToFramework(ctx, poolClient.EnablePropagateAdditionalUserContextData)
	state.EnableTokenRevocation = flex.BoolToFramework(ctx, poolClient.EnableTokenRevocation)
	state.ExplicitAuthFlows = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.ExplicitAuthFlows)
	state.ID = flex.StringToFramework(ctx, poolClient.ClientId)
	state.IdTokenValidity = flex.Int64ToFrameworkLegacy(ctx, poolClient.IdTokenValidity)
	state.LogoutUrls = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.LogoutURLs)
	state.Name = flex.StringToFramework(ctx, poolClient.ClientName)
	state.PreventUserExistenceErrors = flex.StringToFrameworkLegacy(ctx, poolClient.PreventUserExistenceErrors)
	state.ReadAttributes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.ReadAttributes)
	state.RefreshTokenValidity = flex.Int64ToFramework(ctx, poolClient.RefreshTokenValidity)
	state.SupportedIdentityProviders = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.SupportedIdentityProviders)
	if state.TokenValidityUnits.IsNull() && isDefaultTokenValidityUnits(poolClient.TokenValidityUnits) {
		elemType := fwtypes.NewObjectTypeOf[tokenValidityUnits](ctx).ObjectType
		state.TokenValidityUnits = types.ListNull(elemType)
	} else {
		state.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits)
	}
	state.UserPoolID = flex.StringToFramework(ctx, poolClient.UserPoolId)
	state.WriteAttributes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.WriteAttributes)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceUserPoolClient) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var config resourceUserPoolClientData
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceUserPoolClientData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state resourceUserPoolClientData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPConn(ctx)

	params := plan.updateInput(ctx, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	// If removing `token_validity_units`, reset to defaults
	if !state.TokenValidityUnits.IsNull() && plan.TokenValidityUnits.IsNull() {
		params.TokenValidityUnits.AccessToken = aws.String(cognitoidentityprovider.TimeUnitsTypeHours)
		params.TokenValidityUnits.IdToken = aws.String(cognitoidentityprovider.TimeUnitsTypeHours)
		params.TokenValidityUnits.RefreshToken = aws.String(cognitoidentityprovider.TimeUnitsTypeDays)
	}

	output, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.UpdateUserPoolClientWithContext(ctx, params)
	}, cognitoidentityprovider.ErrCodeConcurrentModificationException)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("updating Cognito User Pool Client (%s)", plan.ID.ValueString()),
			err.Error(),
		)
		return
	}

	poolClient := output.(*cognitoidentityprovider.UpdateUserPoolClientOutput).UserPoolClient

	config.AccessTokenValidity = flex.Int64ToFrameworkLegacy(ctx, poolClient.AccessTokenValidity)
	config.AllowedOauthFlows = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.AllowedOAuthFlows)
	config.AllowedOauthFlowsUserPoolClient = flex.BoolToFramework(ctx, poolClient.AllowedOAuthFlowsUserPoolClient)
	config.AllowedOauthScopes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.AllowedOAuthScopes)
	config.AnalyticsConfiguration = flattenAnaylticsConfiguration(ctx, poolClient.AnalyticsConfiguration)
	config.AuthSessionValidity = flex.Int64ToFramework(ctx, poolClient.AuthSessionValidity)
	config.CallbackUrls = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.CallbackURLs)
	config.ClientSecret = flex.StringToFrameworkLegacy(ctx, poolClient.ClientSecret)
	config.DefaultRedirectUri = flex.StringToFrameworkLegacy(ctx, poolClient.DefaultRedirectURI)
	config.EnablePropagateAdditionalUserContextData = flex.BoolToFramework(ctx, poolClient.EnablePropagateAdditionalUserContextData)
	config.EnableTokenRevocation = flex.BoolToFramework(ctx, poolClient.EnableTokenRevocation)
	config.ExplicitAuthFlows = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.ExplicitAuthFlows)
	config.ID = flex.StringToFramework(ctx, poolClient.ClientId)
	config.IdTokenValidity = flex.Int64ToFrameworkLegacy(ctx, poolClient.IdTokenValidity)
	config.LogoutUrls = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.LogoutURLs)
	config.Name = flex.StringToFramework(ctx, poolClient.ClientName)
	config.PreventUserExistenceErrors = flex.StringToFrameworkLegacy(ctx, poolClient.PreventUserExistenceErrors)
	config.ReadAttributes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.ReadAttributes)
	config.RefreshTokenValidity = flex.Int64ToFramework(ctx, poolClient.RefreshTokenValidity)
	config.SupportedIdentityProviders = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.SupportedIdentityProviders)
	if !state.TokenValidityUnits.IsNull() && plan.TokenValidityUnits.IsNull() && isDefaultTokenValidityUnits(poolClient.TokenValidityUnits) {
		elemType := fwtypes.NewObjectTypeOf[tokenValidityUnits](ctx).ObjectType
		config.TokenValidityUnits = types.ListNull(elemType)
	} else {
		config.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits)
	}
	config.UserPoolID = flex.StringToFramework(ctx, poolClient.UserPoolId)
	config.WriteAttributes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.WriteAttributes)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &config)...)
}

func (r *resourceUserPoolClient) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resourceUserPoolClientData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := state.deleteInput(ctx)

	tflog.Debug(ctx, "deleting Cognito User Pool Client", map[string]interface{}{
		"id":           state.ID.ValueString(),
		"user_pool_id": state.UserPoolID.ValueString(),
	})

	conn := r.Meta().CognitoIDPConn(ctx)

	_, err := conn.DeleteUserPoolClientWithContext(ctx, params)
	if tfawserr.ErrCodeEquals(err, cognitoidentityprovider.ErrCodeResourceNotFoundException) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("deleting Cognito User Pool Client (%s)", state.ID.ValueString()),
			err.Error(),
		)
		return
	}
}

func (r *resourceUserPoolClient) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, "/")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf("wrong format of import ID (%s), use: 'user-pool-id/client-id'", request.ID))
		return
	}
	userPoolId := parts[0]
	clientId := parts[1]
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), clientId)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_pool_id"), userPoolId)...)
}

func (r *resourceUserPoolClient) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourceUserPoolClientAccessTokenValidityValidator{
			resourceUserPoolClientValidityValidator{
				attr:        "access_token_validity",
				min:         5 * time.Minute,
				max:         24 * time.Hour,
				defaultUnit: time.Hour,
			},
		},
		resourceUserPoolClientIDTokenValidityValidator{
			resourceUserPoolClientValidityValidator{
				attr:        "id_token_validity",
				min:         5 * time.Minute,
				max:         24 * time.Hour,
				defaultUnit: time.Hour,
			},
		},
		resourceUserPoolClientRefreshTokenValidityValidator{
			resourceUserPoolClientValidityValidator{
				attr:        "refresh_token_validity",
				min:         60 * time.Minute,
				max:         315360000 * time.Second,
				defaultUnit: 24 * time.Hour,
			},
		},
	}
}

type resourceUserPoolClientData struct {
	AccessTokenValidity                      types.Int64  `tfsdk:"access_token_validity"`
	AllowedOauthFlows                        types.Set    `tfsdk:"allowed_oauth_flows"`
	AllowedOauthFlowsUserPoolClient          types.Bool   `tfsdk:"allowed_oauth_flows_user_pool_client"`
	AllowedOauthScopes                       types.Set    `tfsdk:"allowed_oauth_scopes"`
	AnalyticsConfiguration                   types.List   `tfsdk:"analytics_configuration"`
	AuthSessionValidity                      types.Int64  `tfsdk:"auth_session_validity"`
	CallbackUrls                             types.Set    `tfsdk:"callback_urls"`
	ClientSecret                             types.String `tfsdk:"client_secret"`
	DefaultRedirectUri                       types.String `tfsdk:"default_redirect_uri"`
	EnablePropagateAdditionalUserContextData types.Bool   `tfsdk:"enable_propagate_additional_user_context_data"`
	EnableTokenRevocation                    types.Bool   `tfsdk:"enable_token_revocation"`
	ExplicitAuthFlows                        types.Set    `tfsdk:"explicit_auth_flows"`
	GenerateSecret                           types.Bool   `tfsdk:"generate_secret"`
	ID                                       types.String `tfsdk:"id"`
	IdTokenValidity                          types.Int64  `tfsdk:"id_token_validity"`
	LogoutUrls                               types.Set    `tfsdk:"logout_urls"`
	Name                                     types.String `tfsdk:"name"`
	PreventUserExistenceErrors               types.String `tfsdk:"prevent_user_existence_errors"`
	ReadAttributes                           types.Set    `tfsdk:"read_attributes"`
	RefreshTokenValidity                     types.Int64  `tfsdk:"refresh_token_validity"`
	SupportedIdentityProviders               types.Set    `tfsdk:"supported_identity_providers"`
	TokenValidityUnits                       types.List   `tfsdk:"token_validity_units"`
	UserPoolID                               types.String `tfsdk:"user_pool_id"`
	WriteAttributes                          types.Set    `tfsdk:"write_attributes"`
}

func (data resourceUserPoolClientData) createInput(ctx context.Context, diags *diag.Diagnostics) *cognitoidentityprovider.CreateUserPoolClientInput {
	return &cognitoidentityprovider.CreateUserPoolClientInput{
		AccessTokenValidity:                      flex.Int64FromFrameworkLegacy(ctx, data.AccessTokenValidity),
		AllowedOAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthFlows),
		AllowedOAuthFlowsUserPoolClient:          flex.BoolFromFramework(ctx, data.AllowedOauthFlowsUserPoolClient),
		AllowedOAuthScopes:                       flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthScopes),
		AnalyticsConfiguration:                   expandAnaylticsConfiguration(ctx, data.AnalyticsConfiguration, diags),
		AuthSessionValidity:                      flex.Int64FromFramework(ctx, data.AuthSessionValidity),
		CallbackURLs:                             flex.ExpandFrameworkStringSet(ctx, data.CallbackUrls),
		ClientName:                               flex.StringFromFramework(ctx, data.Name),
		DefaultRedirectURI:                       flex.StringFromFrameworkLegacy(ctx, data.DefaultRedirectUri),
		EnablePropagateAdditionalUserContextData: flex.BoolFromFramework(ctx, data.EnablePropagateAdditionalUserContextData),
		EnableTokenRevocation:                    flex.BoolFromFramework(ctx, data.EnableTokenRevocation),
		ExplicitAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.ExplicitAuthFlows),
		GenerateSecret:                           flex.BoolFromFramework(ctx, data.GenerateSecret),
		IdTokenValidity:                          flex.Int64FromFrameworkLegacy(ctx, data.IdTokenValidity),
		LogoutURLs:                               flex.ExpandFrameworkStringSet(ctx, data.LogoutUrls),
		PreventUserExistenceErrors:               flex.StringFromFrameworkLegacy(ctx, data.PreventUserExistenceErrors),
		ReadAttributes:                           flex.ExpandFrameworkStringSet(ctx, data.ReadAttributes),
		RefreshTokenValidity:                     flex.Int64FromFramework(ctx, data.RefreshTokenValidity),
		SupportedIdentityProviders:               flex.ExpandFrameworkStringSet(ctx, data.SupportedIdentityProviders),
		TokenValidityUnits:                       expandTokenValidityUnits(ctx, data.TokenValidityUnits, diags),
		UserPoolId:                               flex.StringFromFramework(ctx, data.UserPoolID),
		WriteAttributes:                          flex.ExpandFrameworkStringSet(ctx, data.WriteAttributes),
	}
}

func (data resourceUserPoolClientData) updateInput(ctx context.Context, diags *diag.Diagnostics) *cognitoidentityprovider.UpdateUserPoolClientInput {
	return &cognitoidentityprovider.UpdateUserPoolClientInput{
		AccessTokenValidity:                      flex.Int64FromFrameworkLegacy(ctx, data.AccessTokenValidity),
		AllowedOAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthFlows),
		AllowedOAuthFlowsUserPoolClient:          flex.BoolFromFramework(ctx, data.AllowedOauthFlowsUserPoolClient),
		AllowedOAuthScopes:                       flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthScopes),
		AnalyticsConfiguration:                   expandAnaylticsConfiguration(ctx, data.AnalyticsConfiguration, diags),
		AuthSessionValidity:                      flex.Int64FromFramework(ctx, data.AuthSessionValidity),
		CallbackURLs:                             flex.ExpandFrameworkStringSet(ctx, data.CallbackUrls),
		ClientId:                                 flex.StringFromFramework(ctx, data.ID),
		ClientName:                               flex.StringFromFramework(ctx, data.Name),
		DefaultRedirectURI:                       flex.StringFromFrameworkLegacy(ctx, data.DefaultRedirectUri),
		EnablePropagateAdditionalUserContextData: flex.BoolFromFramework(ctx, data.EnablePropagateAdditionalUserContextData),
		EnableTokenRevocation:                    flex.BoolFromFramework(ctx, data.EnableTokenRevocation),
		ExplicitAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.ExplicitAuthFlows),
		IdTokenValidity:                          flex.Int64FromFrameworkLegacy(ctx, data.IdTokenValidity),
		LogoutURLs:                               flex.ExpandFrameworkStringSet(ctx, data.LogoutUrls),
		PreventUserExistenceErrors:               flex.StringFromFrameworkLegacy(ctx, data.PreventUserExistenceErrors),
		ReadAttributes:                           flex.ExpandFrameworkStringSet(ctx, data.ReadAttributes),
		RefreshTokenValidity:                     flex.Int64FromFramework(ctx, data.RefreshTokenValidity),
		SupportedIdentityProviders:               flex.ExpandFrameworkStringSet(ctx, data.SupportedIdentityProviders),
		TokenValidityUnits:                       expandTokenValidityUnits(ctx, data.TokenValidityUnits, diags),
		UserPoolId:                               flex.StringFromFramework(ctx, data.UserPoolID),
		WriteAttributes:                          flex.ExpandFrameworkStringSet(ctx, data.WriteAttributes),
	}
}

func (data resourceUserPoolClientData) deleteInput(ctx context.Context) *cognitoidentityprovider.DeleteUserPoolClientInput {
	return &cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   flex.StringFromFramework(ctx, data.ID),
		UserPoolId: flex.StringFromFramework(ctx, data.UserPoolID),
	}
}

type analyticsConfiguration struct {
	ApplicationARN fwtypes.ARN  `tfsdk:"application_arn"`
	ApplicationID  types.String `tfsdk:"application_id"`
	ExternalID     types.String `tfsdk:"external_id"`
	RoleARN        fwtypes.ARN  `tfsdk:"role_arn"`
	UserDataShared types.Bool   `tfsdk:"user_data_shared"`
}

func (ac *analyticsConfiguration) expand(ctx context.Context) *cognitoidentityprovider.AnalyticsConfigurationType {
	if ac == nil {
		return nil
	}
	result := &cognitoidentityprovider.AnalyticsConfigurationType{
		ApplicationArn: flex.StringFromFramework(ctx, ac.ApplicationARN),
		ApplicationId:  flex.StringFromFramework(ctx, ac.ApplicationID),
		ExternalId:     flex.StringFromFramework(ctx, ac.ExternalID),
		RoleArn:        flex.StringFromFramework(ctx, ac.RoleARN),
		UserDataShared: flex.BoolFromFramework(ctx, ac.UserDataShared),
	}

	return result
}

func expandAnaylticsConfiguration(ctx context.Context, list types.List, diags *diag.Diagnostics) *cognitoidentityprovider.AnalyticsConfigurationType {
	var analytics []analyticsConfiguration
	diags.Append(list.ElementsAs(ctx, &analytics, false)...)
	if diags.HasError() {
		return nil
	}

	if len(analytics) == 1 {
		return analytics[0].expand(ctx)
	}
	return nil
}

func flattenAnaylticsConfiguration(ctx context.Context, ac *cognitoidentityprovider.AnalyticsConfigurationType) types.List {
	attributeTypes := fwtypes.AttributeTypesMust[analyticsConfiguration](ctx)
	elemType := types.ObjectType{AttrTypes: attributeTypes}

	if ac == nil {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{}
	attrs["application_arn"] = flex.StringToFrameworkARN(ctx, ac.ApplicationArn)
	attrs["application_id"] = flex.StringToFramework(ctx, ac.ApplicationId)
	attrs["external_id"] = flex.StringToFramework(ctx, ac.ExternalId)
	attrs["role_arn"] = flex.StringToFrameworkARN(ctx, ac.RoleArn)
	attrs["user_data_shared"] = flex.BoolToFramework(ctx, ac.UserDataShared)

	val := types.ObjectValueMust(attributeTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{val})
}

type tokenValidityUnits struct {
	AccessToken  types.String `tfsdk:"access_token"`
	IdToken      types.String `tfsdk:"id_token"`
	RefreshToken types.String `tfsdk:"refresh_token"`
}

func isDefaultTokenValidityUnits(tvu *cognitoidentityprovider.TokenValidityUnitsType) bool {
	if tvu == nil {
		return false
	}
	return aws.StringValue(tvu.AccessToken) == cognitoidentityprovider.TimeUnitsTypeHours &&
		aws.StringValue(tvu.IdToken) == cognitoidentityprovider.TimeUnitsTypeHours &&
		aws.StringValue(tvu.RefreshToken) == cognitoidentityprovider.TimeUnitsTypeDays
}

func (tvu *tokenValidityUnits) expand(ctx context.Context) *cognitoidentityprovider.TokenValidityUnitsType {
	if tvu == nil {
		return nil
	}
	return &cognitoidentityprovider.TokenValidityUnitsType{
		AccessToken:  flex.StringFromFramework(ctx, tvu.AccessToken),
		IdToken:      flex.StringFromFramework(ctx, tvu.IdToken),
		RefreshToken: flex.StringFromFramework(ctx, tvu.RefreshToken),
	}
}

func resolveTokenValidityUnits(ctx context.Context, list types.List, diags *diag.Diagnostics) *tokenValidityUnits {
	var units []tokenValidityUnits
	diags.Append(list.ElementsAs(ctx, &units, false)...)
	if diags.HasError() {
		return nil
	}

	if len(units) == 1 {
		return &units[0]
	}
	return nil
}

func expandTokenValidityUnits(ctx context.Context, list types.List, diags *diag.Diagnostics) *cognitoidentityprovider.TokenValidityUnitsType {
	if tvu := resolveTokenValidityUnits(ctx, list, diags); tvu != nil {
		return tvu.expand(ctx)
	}
	return &cognitoidentityprovider.TokenValidityUnitsType{}
}

func flattenTokenValidityUnits(ctx context.Context, tvu *cognitoidentityprovider.TokenValidityUnitsType) types.List {
	attributeTypes := fwtypes.AttributeTypesMust[tokenValidityUnits](ctx)
	elemType := types.ObjectType{AttrTypes: attributeTypes}

	if tvu == nil || (tvu.AccessToken == nil && tvu.IdToken == nil && tvu.RefreshToken == nil) {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{}
	attrs["access_token"] = flex.StringToFramework(ctx, tvu.AccessToken)
	attrs["id_token"] = flex.StringToFramework(ctx, tvu.IdToken)
	attrs["refresh_token"] = flex.StringToFramework(ctx, tvu.RefreshToken)

	val := types.ObjectValueMust(attributeTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{val})
}

var _ resource.ConfigValidator = &resourceUserPoolClientAccessTokenValidityValidator{}

type resourceUserPoolClientAccessTokenValidityValidator struct {
	resourceUserPoolClientValidityValidator
}

func (v resourceUserPoolClientAccessTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceUserPoolClientData) types.Int64 {
			return rupcd.AccessTokenValidity
		},
		func(tvu *tokenValidityUnits) types.String {
			return tvu.AccessToken
		},
	)
}

var _ resource.ConfigValidator = &resourceUserPoolClientIDTokenValidityValidator{}

type resourceUserPoolClientIDTokenValidityValidator struct {
	resourceUserPoolClientValidityValidator
}

func (v resourceUserPoolClientIDTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceUserPoolClientData) types.Int64 {
			return rupcd.IdTokenValidity
		},
		func(tvu *tokenValidityUnits) types.String {
			return tvu.IdToken
		},
	)
}

var _ resource.ConfigValidator = &resourceUserPoolClientRefreshTokenValidityValidator{}

type resourceUserPoolClientRefreshTokenValidityValidator struct {
	resourceUserPoolClientValidityValidator
}

func (v resourceUserPoolClientRefreshTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceUserPoolClientData) types.Int64 {
			return rupcd.RefreshTokenValidity
		},
		func(tvu *tokenValidityUnits) types.String {
			return tvu.RefreshToken
		},
	)
}

type resourceUserPoolClientValidityValidator struct {
	min         time.Duration
	max         time.Duration
	attr        string
	defaultUnit time.Duration
}

func (v resourceUserPoolClientValidityValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v resourceUserPoolClientValidityValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("must have a duration between %s and %s", v.min, v.max)
}

func (v resourceUserPoolClientValidityValidator) validate(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, valF func(resourceUserPoolClientData) types.Int64, unitF func(*tokenValidityUnits) types.String) {
	var config resourceUserPoolClientData
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	x := valF(config)

	if x.IsUnknown() || x.IsNull() {
		return
	}

	val := aws.Int64Value(flex.Int64FromFramework(ctx, x))

	var duration time.Duration

	units := resolveTokenValidityUnits(ctx, config.TokenValidityUnits, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if units == nil {
		duration = time.Duration(val * int64(v.defaultUnit))
	} else {
		switch aws.StringValue(flex.StringFromFramework(ctx, unitF(units))) {
		case cognitoidentityprovider.TimeUnitsTypeSeconds:
			duration = time.Duration(val * int64(time.Second))
		case cognitoidentityprovider.TimeUnitsTypeMinutes:
			duration = time.Duration(val * int64(time.Minute))
		case cognitoidentityprovider.TimeUnitsTypeHours:
			duration = time.Duration(val * int64(time.Hour))
		case cognitoidentityprovider.TimeUnitsTypeDays:
			duration = time.Duration(val * 24 * int64(time.Hour))
		}
	}

	if duration < v.min || duration > v.max {
		resp.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			path.Root(v.attr),
			v.Description(ctx),
			duration.String(),
		))
	}
}
