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
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceManagedUserPoolClient(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceManagedUserPoolClient{}, nil
}

type resourceManagedUserPoolClient struct {
	framework.ResourceWithConfigure
}

func (r *resourceManagedUserPoolClient) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cognito_managed_user_pool_client"
}

// Schema returns the schema for this resource.
func (r *resourceManagedUserPoolClient) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name_pattern": schema.StringAttribute{
				CustomType: fwtypes.RegexpType,
				Optional:   true,
				Validators: append(
					userPoolClientNameValidator,
					stringvalidator.ExactlyOneOf(
						path.MatchRelative().AtParent().AtName("name_prefix"),
						path.MatchRelative().AtParent().AtName("name_pattern"),
					),
				),
			},
			"name_prefix": schema.StringAttribute{
				Optional:   true,
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

func (r *resourceManagedUserPoolClient) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPConn(ctx)

	var config resourceManagedUserPoolClientData
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceManagedUserPoolClientData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	userPoolId := plan.UserPoolID.ValueString()

	var nameMatcher cognitoUserPoolClientDescriptionNameFilter
	if namePattern := plan.NamePattern; !namePattern.IsUnknown() && !namePattern.IsNull() {
		nameMatcher = func(name string) (bool, error) {
			return namePattern.ValueRegexp().MatchString(name), nil
		}
	}
	if namePrefix := plan.NamePrefix; !namePrefix.IsUnknown() && !namePrefix.IsNull() {
		nameMatcher = func(name string) (bool, error) {
			return strings.HasPrefix(name, namePrefix.ValueString()), nil
		}
	}

	poolClient, err := FindCognitoUserPoolClientByName(ctx, conn, userPoolId, nameMatcher)
	if err != nil {
		response.Diagnostics.AddError(
			"acquiring Cognito User Pool Client",
			err.Error(),
		)
		return
	}

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

	needsUpdate := false

	if !plan.AccessTokenValidity.IsUnknown() && !plan.AccessTokenValidity.Equal(config.AccessTokenValidity) {
		needsUpdate = true
		config.AccessTokenValidity = plan.AccessTokenValidity
	}
	if !plan.AllowedOauthFlows.IsUnknown() && !plan.AllowedOauthFlows.Equal(config.AllowedOauthFlows) {
		needsUpdate = true
		config.AllowedOauthFlows = plan.AllowedOauthFlows
	}
	if !plan.AllowedOauthFlowsUserPoolClient.IsUnknown() && !plan.AllowedOauthFlowsUserPoolClient.Equal(config.AllowedOauthFlowsUserPoolClient) {
		needsUpdate = true
		config.AllowedOauthFlowsUserPoolClient = plan.AllowedOauthFlowsUserPoolClient
	}
	if !plan.AllowedOauthScopes.IsUnknown() && !plan.AllowedOauthScopes.Equal(config.AllowedOauthScopes) {
		needsUpdate = true
		config.AllowedOauthScopes = plan.AllowedOauthScopes
	}
	if !plan.AnalyticsConfiguration.IsUnknown() && !plan.AnalyticsConfiguration.Equal(config.AnalyticsConfiguration) {
		needsUpdate = true
		config.AnalyticsConfiguration = plan.AnalyticsConfiguration
	}
	if !plan.AuthSessionValidity.IsUnknown() && !plan.AuthSessionValidity.Equal(config.AuthSessionValidity) {
		needsUpdate = true
		config.AuthSessionValidity = plan.AuthSessionValidity
	}
	if !plan.CallbackUrls.IsUnknown() && !plan.CallbackUrls.Equal(config.CallbackUrls) {
		needsUpdate = true
		config.CallbackUrls = plan.CallbackUrls
	}
	if !plan.DefaultRedirectUri.IsUnknown() && !plan.DefaultRedirectUri.Equal(config.DefaultRedirectUri) {
		needsUpdate = true
		config.DefaultRedirectUri = plan.DefaultRedirectUri
	}
	if !plan.EnablePropagateAdditionalUserContextData.IsUnknown() && !plan.EnablePropagateAdditionalUserContextData.Equal(config.EnablePropagateAdditionalUserContextData) {
		needsUpdate = true
		config.EnablePropagateAdditionalUserContextData = plan.EnablePropagateAdditionalUserContextData
	}
	if !plan.EnableTokenRevocation.IsUnknown() && !plan.EnableTokenRevocation.Equal(config.EnableTokenRevocation) {
		needsUpdate = true
		config.EnableTokenRevocation = plan.EnableTokenRevocation
	}
	if !plan.ExplicitAuthFlows.IsUnknown() && !plan.ExplicitAuthFlows.Equal(config.ExplicitAuthFlows) {
		needsUpdate = true
		config.ExplicitAuthFlows = plan.ExplicitAuthFlows
	}
	if !plan.IdTokenValidity.IsUnknown() && !plan.IdTokenValidity.Equal(config.IdTokenValidity) {
		needsUpdate = true
		config.IdTokenValidity = plan.IdTokenValidity
	}
	if !plan.LogoutUrls.IsUnknown() && !plan.LogoutUrls.Equal(config.LogoutUrls) {
		needsUpdate = true
		config.LogoutUrls = plan.LogoutUrls
	}
	if !plan.PreventUserExistenceErrors.IsUnknown() && !plan.PreventUserExistenceErrors.Equal(config.PreventUserExistenceErrors) {
		needsUpdate = true
		config.PreventUserExistenceErrors = plan.PreventUserExistenceErrors
	}
	if !plan.ReadAttributes.IsUnknown() && !plan.ReadAttributes.Equal(config.ReadAttributes) {
		needsUpdate = true
		config.ReadAttributes = plan.ReadAttributes
	}
	if !plan.RefreshTokenValidity.IsUnknown() && !plan.RefreshTokenValidity.Equal(config.RefreshTokenValidity) {
		needsUpdate = true
		config.RefreshTokenValidity = plan.RefreshTokenValidity
	}
	if !plan.SupportedIdentityProviders.IsUnknown() && !plan.SupportedIdentityProviders.Equal(config.SupportedIdentityProviders) {
		needsUpdate = true
		config.SupportedIdentityProviders = plan.SupportedIdentityProviders
	}
	if !plan.TokenValidityUnits.IsUnknown() && !plan.TokenValidityUnits.Equal(config.TokenValidityUnits) {
		needsUpdate = true
		config.TokenValidityUnits = plan.TokenValidityUnits
	}
	if !plan.WriteAttributes.IsUnknown() && !plan.WriteAttributes.Equal(config.WriteAttributes) {
		needsUpdate = true
		config.WriteAttributes = plan.WriteAttributes
	}

	if needsUpdate {
		params := config.updateInput(ctx, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
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
		config.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits)
		config.UserPoolID = flex.StringToFramework(ctx, poolClient.UserPoolId)
		config.WriteAttributes = flex.FlattenFrameworkStringSetLegacy(ctx, poolClient.WriteAttributes)

		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &config)...)
}

func (r *resourceManagedUserPoolClient) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resourceManagedUserPoolClientData
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

func (r *resourceManagedUserPoolClient) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var config resourceManagedUserPoolClientData
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceManagedUserPoolClientData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state resourceManagedUserPoolClientData
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

func (r *resourceManagedUserPoolClient) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resourceManagedUserPoolClientData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.AddWarning(
		"Cognito User Pool Client (%s) not deleted",
		"User Pool Client is managed by another service and will be deleted when that resource is deleted. Removed from Terraform state.",
	)
}

func (r *resourceManagedUserPoolClient) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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

func (r *resourceManagedUserPoolClient) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourceManagedUserPoolClientAccessTokenValidityValidator{
			resourceManagedUserPoolClientValidityValidator{
				attr:        "access_token_validity",
				min:         5 * time.Minute,
				max:         24 * time.Hour,
				defaultUnit: time.Hour,
			},
		},
		resourceManagedUserPoolClientIDTokenValidityValidator{
			resourceManagedUserPoolClientValidityValidator{
				attr:        "id_token_validity",
				min:         5 * time.Minute,
				max:         24 * time.Hour,
				defaultUnit: time.Hour,
			},
		},
		resourceManagedUserPoolClientRefreshTokenValidityValidator{
			resourceManagedUserPoolClientValidityValidator{
				attr:        "refresh_token_validity",
				min:         60 * time.Minute,
				max:         315360000 * time.Second,
				defaultUnit: 24 * time.Hour,
			},
		},
	}
}

type resourceManagedUserPoolClientData struct {
	AccessTokenValidity                      types.Int64    `tfsdk:"access_token_validity"`
	AllowedOauthFlows                        types.Set      `tfsdk:"allowed_oauth_flows"`
	AllowedOauthFlowsUserPoolClient          types.Bool     `tfsdk:"allowed_oauth_flows_user_pool_client"`
	AllowedOauthScopes                       types.Set      `tfsdk:"allowed_oauth_scopes"`
	AnalyticsConfiguration                   types.List     `tfsdk:"analytics_configuration"`
	AuthSessionValidity                      types.Int64    `tfsdk:"auth_session_validity"`
	CallbackUrls                             types.Set      `tfsdk:"callback_urls"`
	ClientSecret                             types.String   `tfsdk:"client_secret"`
	DefaultRedirectUri                       types.String   `tfsdk:"default_redirect_uri"`
	EnablePropagateAdditionalUserContextData types.Bool     `tfsdk:"enable_propagate_additional_user_context_data"`
	EnableTokenRevocation                    types.Bool     `tfsdk:"enable_token_revocation"`
	ExplicitAuthFlows                        types.Set      `tfsdk:"explicit_auth_flows"`
	ID                                       types.String   `tfsdk:"id"`
	IdTokenValidity                          types.Int64    `tfsdk:"id_token_validity"`
	LogoutUrls                               types.Set      `tfsdk:"logout_urls"`
	Name                                     types.String   `tfsdk:"name"`
	NamePattern                              fwtypes.Regexp `tfsdk:"name_pattern"`
	NamePrefix                               types.String   `tfsdk:"name_prefix"`
	PreventUserExistenceErrors               types.String   `tfsdk:"prevent_user_existence_errors"`
	ReadAttributes                           types.Set      `tfsdk:"read_attributes"`
	RefreshTokenValidity                     types.Int64    `tfsdk:"refresh_token_validity"`
	SupportedIdentityProviders               types.Set      `tfsdk:"supported_identity_providers"`
	TokenValidityUnits                       types.List     `tfsdk:"token_validity_units"`
	UserPoolID                               types.String   `tfsdk:"user_pool_id"`
	WriteAttributes                          types.Set      `tfsdk:"write_attributes"`
}

func (data resourceManagedUserPoolClientData) updateInput(ctx context.Context, diags *diag.Diagnostics) *cognitoidentityprovider.UpdateUserPoolClientInput {
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

var _ resource.ConfigValidator = &resourceManagedUserPoolClientAccessTokenValidityValidator{}

type resourceManagedUserPoolClientAccessTokenValidityValidator struct {
	resourceManagedUserPoolClientValidityValidator
}

func (v resourceManagedUserPoolClientAccessTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceManagedUserPoolClientData) types.Int64 {
			return rupcd.AccessTokenValidity
		},
		func(tvu *tokenValidityUnits) types.String {
			return tvu.AccessToken
		},
	)
}

var _ resource.ConfigValidator = &resourceManagedUserPoolClientIDTokenValidityValidator{}

type resourceManagedUserPoolClientIDTokenValidityValidator struct {
	resourceManagedUserPoolClientValidityValidator
}

func (v resourceManagedUserPoolClientIDTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceManagedUserPoolClientData) types.Int64 {
			return rupcd.IdTokenValidity
		},
		func(tvu *tokenValidityUnits) types.String {
			return tvu.IdToken
		},
	)
}

var _ resource.ConfigValidator = &resourceManagedUserPoolClientRefreshTokenValidityValidator{}

type resourceManagedUserPoolClientRefreshTokenValidityValidator struct {
	resourceManagedUserPoolClientValidityValidator
}

func (v resourceManagedUserPoolClientRefreshTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceManagedUserPoolClientData) types.Int64 {
			return rupcd.RefreshTokenValidity
		},
		func(tvu *tokenValidityUnits) types.String {
			return tvu.RefreshToken
		},
	)
}

type resourceManagedUserPoolClientValidityValidator struct {
	min         time.Duration
	max         time.Duration
	attr        string
	defaultUnit time.Duration
}

func (v resourceManagedUserPoolClientValidityValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v resourceManagedUserPoolClientValidityValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("must have a duration between %s and %s", v.min, v.max)
}

func (v resourceManagedUserPoolClientValidityValidator) validate(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, valF func(resourceManagedUserPoolClientData) types.Int64, unitF func(*tokenValidityUnits) types.String) {
	var config resourceManagedUserPoolClientData
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
