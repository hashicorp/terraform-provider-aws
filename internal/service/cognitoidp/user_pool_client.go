package cognitoidp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

// @FrameworkResource
func newResourceUserPoolClient(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceUserPoolClient{}
	r.SetMigratedFromPluginSDK(true)

	return r, nil
}

type resourceUserPoolClient struct {
	framework.ResourceWithConfigure
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
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
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"allowed_oauth_flows": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
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
				Required:   true,
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

func (r *resourceUserPoolClient) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPConn()

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

	data := newUserPoolClientData(ctx, plan, poolClient, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceUserPoolClient) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceUserPoolClientData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPConn()

	userPoolClient, err := FindCognitoUserPoolClientByID(ctx, conn, data.UserPoolID.ValueString(), data.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolClient, data.ID.ValueString())
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.Append(create.DiagErrorFramework(names.CognitoIDP, create.ErrActionReading, ResNameUserPoolClient, data.ID.ValueString(), err))
		return
	}

	data = newUserPoolClientData(ctx, data, userPoolClient, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceUserPoolClient) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan resourceUserPoolClientData
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

	data := newUserPoolClientData(ctx, plan, poolClient, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
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

	conn := r.Meta().CognitoIDPConn()

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
	}
	userPoolId := parts[0]
	clientId := parts[1]
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), clientId)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_pool_id"), userPoolId)...)
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

func newUserPoolClientData(ctx context.Context, plan resourceUserPoolClientData, in *cognitoidentityprovider.UserPoolClientType, diags *diag.Diagnostics) resourceUserPoolClientData {
	return resourceUserPoolClientData{
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
		GenerateSecret:                           plan.GenerateSecret,
		LogoutUrls:                               flex.FlattenFrameworkStringSet(ctx, in.LogoutURLs),
		Name:                                     flex.StringToFramework(ctx, in.ClientName),
		PreventUserExistenceErrors:               flex.StringToFrameworkLegacy(ctx, in.PreventUserExistenceErrors),
		ReadAttributes:                           flex.FlattenFrameworkStringSet(ctx, in.ReadAttributes),
		RefreshTokenValidity:                     flex.Int64ToFramework(ctx, in.RefreshTokenValidity),
		SupportedIdentityProviders:               flex.FlattenFrameworkStringSet(ctx, in.SupportedIdentityProviders),
		TokenValidityUnits:                       flattenTokenValidityUnits(ctx, in.TokenValidityUnits),
		UserPoolID:                               flex.StringToFramework(ctx, in.UserPoolId),
		WriteAttributes:                          flex.FlattenFrameworkStringSet(ctx, in.WriteAttributes),
	}
}

func (data resourceUserPoolClientData) createInput(ctx context.Context, diags *diag.Diagnostics) *cognitoidentityprovider.CreateUserPoolClientInput {
	return &cognitoidentityprovider.CreateUserPoolClientInput{
		AccessTokenValidity:                      flex.Int64FromFramework(ctx, data.AccessTokenValidity),
		AllowedOAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthFlows),
		AllowedOAuthFlowsUserPoolClient:          flex.BoolFromFramework(ctx, data.AllowedOauthFlowsUserPoolClient),
		AllowedOAuthScopes:                       flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthScopes),
		AnalyticsConfiguration:                   expandAnaylticsConfiguration(ctx, data.AnalyticsConfiguration, diags),
		AuthSessionValidity:                      flex.Int64FromFramework(ctx, data.AuthSessionValidity),
		CallbackURLs:                             flex.ExpandFrameworkStringSet(ctx, data.CallbackUrls),
		ClientName:                               flex.StringFromFramework(ctx, data.Name),
		DefaultRedirectURI:                       flex.StringFromFramework(ctx, data.DefaultRedirectUri),
		EnablePropagateAdditionalUserContextData: flex.BoolFromFramework(ctx, data.EnablePropagateAdditionalUserContextData),
		EnableTokenRevocation:                    flex.BoolFromFramework(ctx, data.EnableTokenRevocation),
		ExplicitAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.ExplicitAuthFlows),
		GenerateSecret:                           flex.BoolFromFramework(ctx, data.GenerateSecret),
		IdTokenValidity:                          flex.Int64FromFramework(ctx, data.IdTokenValidity),
		LogoutURLs:                               flex.ExpandFrameworkStringSet(ctx, data.LogoutUrls),
		PreventUserExistenceErrors:               flex.StringFromFramework(ctx, data.PreventUserExistenceErrors),
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
		AccessTokenValidity:                      flex.Int64FromFramework(ctx, data.AccessTokenValidity),
		AllowedOAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthFlows),
		AllowedOAuthFlowsUserPoolClient:          flex.BoolFromFramework(ctx, data.AllowedOauthFlowsUserPoolClient),
		AllowedOAuthScopes:                       flex.ExpandFrameworkStringSet(ctx, data.AllowedOauthScopes),
		AnalyticsConfiguration:                   expandAnaylticsConfiguration(ctx, data.AnalyticsConfiguration, diags),
		AuthSessionValidity:                      flex.Int64FromFramework(ctx, data.AuthSessionValidity),
		CallbackURLs:                             flex.ExpandFrameworkStringSet(ctx, data.CallbackUrls),
		ClientId:                                 flex.StringFromFramework(ctx, data.ID),
		ClientName:                               flex.StringFromFramework(ctx, data.Name),
		DefaultRedirectURI:                       flex.StringFromFramework(ctx, data.DefaultRedirectUri),
		EnablePropagateAdditionalUserContextData: flex.BoolFromFramework(ctx, data.EnablePropagateAdditionalUserContextData),
		EnableTokenRevocation:                    flex.BoolFromFramework(ctx, data.EnableTokenRevocation),
		ExplicitAuthFlows:                        flex.ExpandFrameworkStringSet(ctx, data.ExplicitAuthFlows),
		IdTokenValidity:                          flex.Int64FromFramework(ctx, data.IdTokenValidity),
		LogoutURLs:                               flex.ExpandFrameworkStringSet(ctx, data.LogoutUrls),
		PreventUserExistenceErrors:               flex.StringFromFramework(ctx, data.PreventUserExistenceErrors),
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
		ApplicationArn: flex.ARNStringFromFramework(ctx, ac.ApplicationARN),
		ApplicationId:  flex.StringFromFramework(ctx, ac.ApplicationID),
		ExternalId:     flex.StringFromFramework(ctx, ac.ExternalID),
		RoleArn:        flex.ARNStringFromFramework(ctx, ac.RoleARN),
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

func flattenAnaylticsConfiguration(ctx context.Context, ac *cognitoidentityprovider.AnalyticsConfigurationType, diags *diag.Diagnostics) types.List {
	attributeTypes := framework.AttributeTypesMust[analyticsConfiguration](ctx)
	elemType := types.ObjectType{AttrTypes: attributeTypes}

	if ac == nil {
		return types.ListNull(elemType)
	}

	attrs := map[string]attr.Value{}
	attrs["application_arn"] = flex.StringToFrameworkARN(ctx, ac.ApplicationArn, diags)
	attrs["application_id"] = flex.StringToFramework(ctx, ac.ApplicationId)
	attrs["external_id"] = flex.StringToFramework(ctx, ac.ExternalId)
	attrs["role_arn"] = flex.StringToFrameworkARN(ctx, ac.RoleArn, diags)
	attrs["user_data_shared"] = flex.BoolToFramework(ctx, ac.UserDataShared)

	val := types.ObjectValueMust(attributeTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{val})
}

type tokenValidityUnits struct {
	AccessToken  types.String `tfsdk:"access_token"`
	IdToken      types.String `tfsdk:"id_token"`
	RefreshToken types.String `tfsdk:"refresh_token"`
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

func expandTokenValidityUnits(ctx context.Context, list types.List, diags *diag.Diagnostics) *cognitoidentityprovider.TokenValidityUnitsType {
	var units []tokenValidityUnits
	diags.Append(list.ElementsAs(ctx, &units, false)...)
	if diags.HasError() {
		return nil
	}

	if len(units) == 1 {
		return units[0].expand(ctx)
	}
	return nil
}

func flattenTokenValidityUnits(ctx context.Context, tvu *cognitoidentityprovider.TokenValidityUnitsType) types.List {
	attributeTypes := framework.AttributeTypesMust[tokenValidityUnits](ctx)
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
