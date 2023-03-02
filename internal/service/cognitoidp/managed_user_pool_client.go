package cognitoidp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwstringplanmodifier "github.com/hashicorp/terraform-provider-aws/internal/framework/stringplanmodifier"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	_sp.registerFrameworkResourceFactory(newResourceManagedUserPoolClient)
}

func newResourceManagedUserPoolClient(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceManagedUserPoolClient{}
	r.SetMigratedFromPluginSDK(true)

	return r, nil
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
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
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
			},
			"allowed_oauth_flows_user_pool_client": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"allowed_oauth_scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(50),
				},
			},
			"auth_session_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(3, 15),
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
			},
			"client_secret": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"default_redirect_uri": schema.StringAttribute{
				Optional:   true,
				Computed:   true,
				Validators: userPoolClientURLValidator,
			},
			"enable_propagate_additional_user_context_data": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"enable_token_revocation": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"explicit_auth_flows": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.OneOf(cognitoidentityprovider.ExplicitAuthFlowsType_Values()...),
					),
				},
			},
			"id": framework.IDAttribute(),
			"id_token_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
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
			},
			"name": schema.StringAttribute{
				Computed: true,
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
			},
			"read_attributes": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"refresh_token_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(0, 315360000),
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
							PlanModifiers: []planmodifier.String{
								fwstringplanmodifier.DefaultValue(cognitoidentityprovider.TimeUnitsTypeHours),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(cognitoidentityprovider.TimeUnitsType_Values()...),
							},
						},
						"id_token": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								fwstringplanmodifier.DefaultValue(cognitoidentityprovider.TimeUnitsTypeHours),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(cognitoidentityprovider.TimeUnitsType_Values()...),
							},
						},
						"refresh_token": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								fwstringplanmodifier.DefaultValue(cognitoidentityprovider.TimeUnitsTypeDays),
							},
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
	conn := r.Meta().CognitoIDPConn()

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

	userPoolClient, err := FindCognitoUserPoolClientByName(ctx, conn, userPoolId, nameMatcher)
	if err != nil {
		response.Diagnostics.AddError(
			"acquiring Cognito User Pool Client",
			err.Error(),
		)
		return
	}

	data := newManagedUserPoolClientData(ctx, plan, userPoolClient, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	needsUpdate := false

	if !plan.AccessTokenValidity.IsUnknown() && !plan.AccessTokenValidity.Equal(data.AccessTokenValidity) {
		needsUpdate = true
		data.AccessTokenValidity = plan.AccessTokenValidity
	}
	if !plan.AllowedOauthFlows.IsUnknown() && !plan.AllowedOauthFlows.Equal(data.AllowedOauthFlows) {
		needsUpdate = true
		data.AllowedOauthFlows = plan.AllowedOauthFlows
	}
	if !plan.AllowedOauthFlowsUserPoolClient.IsUnknown() && !plan.AllowedOauthFlowsUserPoolClient.Equal(data.AllowedOauthFlowsUserPoolClient) {
		needsUpdate = true
		data.AllowedOauthFlowsUserPoolClient = plan.AllowedOauthFlowsUserPoolClient
	}
	if !plan.AllowedOauthScopes.IsUnknown() && !plan.AllowedOauthScopes.Equal(data.AllowedOauthScopes) {
		needsUpdate = true
		data.AllowedOauthScopes = plan.AllowedOauthScopes
	}
	if !plan.AnalyticsConfiguration.IsUnknown() && !plan.AnalyticsConfiguration.Equal(data.AnalyticsConfiguration) {
		needsUpdate = true
		data.AnalyticsConfiguration = plan.AnalyticsConfiguration
	}
	if !plan.AuthSessionValidity.IsUnknown() && !plan.AuthSessionValidity.Equal(data.AuthSessionValidity) {
		needsUpdate = true
		data.AuthSessionValidity = plan.AuthSessionValidity
	}
	if !plan.CallbackUrls.IsUnknown() && !plan.CallbackUrls.Equal(data.CallbackUrls) {
		needsUpdate = true
		data.CallbackUrls = plan.CallbackUrls
	}
	if !plan.DefaultRedirectUri.IsUnknown() && !plan.DefaultRedirectUri.Equal(data.DefaultRedirectUri) {
		needsUpdate = true
		data.DefaultRedirectUri = plan.DefaultRedirectUri
	}
	if !plan.EnablePropagateAdditionalUserContextData.IsUnknown() && !plan.EnablePropagateAdditionalUserContextData.Equal(data.EnablePropagateAdditionalUserContextData) {
		needsUpdate = true
		data.EnablePropagateAdditionalUserContextData = plan.EnablePropagateAdditionalUserContextData
	}
	if !plan.EnableTokenRevocation.IsUnknown() && !plan.EnableTokenRevocation.Equal(data.EnableTokenRevocation) {
		needsUpdate = true
		data.EnableTokenRevocation = plan.EnableTokenRevocation
	}
	if !plan.ExplicitAuthFlows.IsUnknown() && !plan.ExplicitAuthFlows.Equal(data.ExplicitAuthFlows) {
		needsUpdate = true
		data.ExplicitAuthFlows = plan.ExplicitAuthFlows
	}
	if !plan.IdTokenValidity.IsUnknown() && !plan.IdTokenValidity.Equal(data.IdTokenValidity) {
		needsUpdate = true
		data.IdTokenValidity = plan.IdTokenValidity
	}
	if !plan.LogoutUrls.IsUnknown() && !plan.LogoutUrls.Equal(data.LogoutUrls) {
		needsUpdate = true
		data.LogoutUrls = plan.LogoutUrls
	}
	if !plan.PreventUserExistenceErrors.IsUnknown() && !plan.PreventUserExistenceErrors.Equal(data.PreventUserExistenceErrors) {
		needsUpdate = true
		data.PreventUserExistenceErrors = plan.PreventUserExistenceErrors
	}
	if !plan.ReadAttributes.IsUnknown() && !plan.ReadAttributes.Equal(data.ReadAttributes) {
		needsUpdate = true
		data.ReadAttributes = plan.ReadAttributes
	}
	if !plan.RefreshTokenValidity.IsUnknown() && !plan.RefreshTokenValidity.Equal(data.RefreshTokenValidity) {
		needsUpdate = true
		data.RefreshTokenValidity = plan.RefreshTokenValidity
	}
	if !plan.SupportedIdentityProviders.IsUnknown() && !plan.SupportedIdentityProviders.Equal(data.SupportedIdentityProviders) {
		needsUpdate = true
		data.SupportedIdentityProviders = plan.SupportedIdentityProviders
	}
	if !plan.TokenValidityUnits.IsUnknown() && !plan.TokenValidityUnits.Equal(data.TokenValidityUnits) {
		needsUpdate = true
		data.TokenValidityUnits = plan.TokenValidityUnits
	}
	if !plan.WriteAttributes.IsUnknown() && !plan.WriteAttributes.Equal(data.WriteAttributes) {
		needsUpdate = true
		data.WriteAttributes = plan.WriteAttributes
	}

	if needsUpdate {
		params := data.updateInput(ctx, &response.Diagnostics)
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

		data = newManagedUserPoolClientData(ctx, data, poolClient, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceManagedUserPoolClient) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resourceManagedUserPoolClientData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPConn()

	userPoolClient, err := FindCognitoUserPoolClientByID(ctx, conn, state.UserPoolID.ValueString(), state.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolClient, state.ID.ValueString())
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.Append(create.DiagErrorFramework(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolClient, state.ID.ValueString(), err))
		return
	}

	state = newManagedUserPoolClientData(ctx, state, userPoolClient, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceManagedUserPoolClient) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, config, plan resourceManagedUserPoolClientData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPConn()

	params := plan.updateInput(ctx, &response.Diagnostics)
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

	data := newManagedUserPoolClientData(ctx, plan, poolClient, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceManagedUserPoolClient) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceManagedUserPoolClientData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting TODO", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	response.Diagnostics.AddWarning(
		"Cognito User Pool Client (%s) not deleted",
		"User Pool Client is managed by another service and will be deleted when that resource is deleted. Removed from Terraform state.",
	)
}

func (r *resourceManagedUserPoolClient) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, "/")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf("wrong format of import ID (%s), use: 'user-pool-id/client-id'", request.ID))
	}
	userPoolId := parts[0]
	clientId := parts[1]
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), clientId)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_pool_id"), userPoolId)...)
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

func newManagedUserPoolClientData(ctx context.Context, plan resourceManagedUserPoolClientData, in *cognitoidentityprovider.UserPoolClientType, diags *diag.Diagnostics) resourceManagedUserPoolClientData {
	return resourceManagedUserPoolClientData{
		AccessTokenValidity:                      flex.Int64ToFrameworkLegacy(ctx, in.AccessTokenValidity),
		AllowedOauthFlows:                        flex.FlattenFrameworkStringSet(ctx, in.AllowedOAuthFlows),
		AllowedOauthFlowsUserPoolClient:          flex.BoolToFramework(ctx, in.AllowedOAuthFlowsUserPoolClient),
		AllowedOauthScopes:                       flex.FlattenFrameworkStringSet(ctx, in.AllowedOAuthScopes),
		AnalyticsConfiguration:                   flattenAnaylticsConfiguration(ctx, in.AnalyticsConfiguration, diags),
		AuthSessionValidity:                      flex.Int64ToFramework(ctx, in.AuthSessionValidity),
		CallbackUrls:                             flex.FlattenFrameworkStringSet(ctx, in.CallbackURLs),
		ClientSecret:                             flex.StringToFrameworkLegacy(ctx, in.ClientSecret),
		DefaultRedirectUri:                       flex.StringToFrameworkLegacy(ctx, in.DefaultRedirectURI),
		EnablePropagateAdditionalUserContextData: flex.BoolToFramework(ctx, in.EnablePropagateAdditionalUserContextData),
		EnableTokenRevocation:                    flex.BoolToFramework(ctx, in.EnableTokenRevocation),
		ExplicitAuthFlows:                        flex.FlattenFrameworkStringSet(ctx, in.ExplicitAuthFlows),
		ID:                                       flex.StringToFramework(ctx, in.ClientId),
		IdTokenValidity:                          flex.Int64ToFrameworkLegacy(ctx, in.IdTokenValidity),
		LogoutUrls:                               flex.FlattenFrameworkStringSet(ctx, in.LogoutURLs),
		Name:                                     flex.StringToFramework(ctx, in.ClientName),
		NamePattern:                              plan.NamePattern,
		NamePrefix:                               plan.NamePrefix,
		PreventUserExistenceErrors:               flex.StringToFrameworkLegacy(ctx, in.PreventUserExistenceErrors),
		ReadAttributes:                           flex.FlattenFrameworkStringSet(ctx, in.ReadAttributes),
		RefreshTokenValidity:                     flex.Int64ToFramework(ctx, in.RefreshTokenValidity),
		SupportedIdentityProviders:               flex.FlattenFrameworkStringSet(ctx, in.SupportedIdentityProviders),
		TokenValidityUnits:                       flattenTokenValidityUnits(ctx, in.TokenValidityUnits),
		UserPoolID:                               flex.StringToFramework(ctx, in.UserPoolId),
		WriteAttributes:                          flex.FlattenFrameworkStringSet(ctx, in.WriteAttributes),
	}
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
