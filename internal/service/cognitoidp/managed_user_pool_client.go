package cognitoidp

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
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
			"name_prefix": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexp.MustCompile(`[\w\s+=,.@-]+`), "XXX"),
				},
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

// Create is called when the provider must create a new resource.
// Config and planned state values should be read from the CreateRequest and new state values set on the CreateResponse.
func (r *resourceManagedUserPoolClient) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPConn()

	var plan resourceManagedUserPoolClientData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	userPoolId := plan.UserPoolID.ValueString()

	var nameMatcher cognitoUserPoolClientDescriptionNameFilter
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

	data := plan
	data.ID = types.StringValue(aws.StringValue(userPoolClient.ClientId))

	data.AccessTokenValidity = flex.Int64ToFrameworkLegacy(ctx, userPoolClient.AccessTokenValidity)
	data.AllowedOauthFlows = flex.FlattenFrameworkStringSet(ctx, userPoolClient.AllowedOAuthFlows)
	data.AllowedOauthFlowsUserPoolClient = flex.BoolToFramework(ctx, userPoolClient.AllowedOAuthFlowsUserPoolClient)
	data.AllowedOauthScopes = flex.FlattenFrameworkStringSet(ctx, userPoolClient.AllowedOAuthScopes)
	data.AnalyticsConfiguration = flattenAnaylticsConfiguration(ctx, userPoolClient.AnalyticsConfiguration, &response.Diagnostics)
	data.AuthSessionValidity = flex.Int64ToFramework(ctx, userPoolClient.AuthSessionValidity)
	data.CallbackUrls = flex.FlattenFrameworkStringSet(ctx, userPoolClient.CallbackURLs)
	data.ClientSecret = flex.StringToFrameworkLegacy(ctx, userPoolClient.ClientSecret)
	data.DefaultRedirectUri = flex.StringToFrameworkLegacy(ctx, userPoolClient.DefaultRedirectURI)
	data.EnablePropagateAdditionalUserContextData = flex.BoolToFramework(ctx, userPoolClient.EnablePropagateAdditionalUserContextData)
	data.EnableTokenRevocation = flex.BoolToFramework(ctx, userPoolClient.EnableTokenRevocation)
	data.ExplicitAuthFlows = flex.FlattenFrameworkStringSet(ctx, userPoolClient.ExplicitAuthFlows)
	data.IdTokenValidity = flex.Int64ToFrameworkLegacy(ctx, userPoolClient.IdTokenValidity)
	data.LogoutUrls = flex.FlattenFrameworkStringSet(ctx, userPoolClient.LogoutURLs)
	data.Name = flex.StringToFramework(ctx, userPoolClient.ClientName)
	data.PreventUserExistenceErrors = flex.StringToFrameworkLegacy(ctx, userPoolClient.PreventUserExistenceErrors)
	data.ReadAttributes = flex.FlattenFrameworkStringSet(ctx, userPoolClient.ReadAttributes)
	data.RefreshTokenValidity = flex.Int64ToFramework(ctx, userPoolClient.RefreshTokenValidity)
	data.SupportedIdentityProviders = flex.FlattenFrameworkStringSet(ctx, userPoolClient.SupportedIdentityProviders)
	data.TokenValidityUnits = flattenTokenValidityUnits(ctx, userPoolClient.TokenValidityUnits)
	data.UserPoolID = flex.StringToFramework(ctx, userPoolClient.UserPoolId)
	data.WriteAttributes = flex.FlattenFrameworkStringSet(ctx, userPoolClient.WriteAttributes)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Read is called when the provider must read resource values in order to update state.
// Planned state values should be read from the ReadRequest and new state values set on the ReadResponse.
func (r *resourceManagedUserPoolClient) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceManagedUserPoolClientData

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

	data.AccessTokenValidity = flex.Int64ToFrameworkLegacy(ctx, userPoolClient.AccessTokenValidity)
	data.AllowedOauthFlows = flex.FlattenFrameworkStringSet(ctx, userPoolClient.AllowedOAuthFlows)
	data.AllowedOauthFlowsUserPoolClient = flex.BoolToFramework(ctx, userPoolClient.AllowedOAuthFlowsUserPoolClient)
	data.AllowedOauthScopes = flex.FlattenFrameworkStringSet(ctx, userPoolClient.AllowedOAuthScopes)
	data.AnalyticsConfiguration = flattenAnaylticsConfiguration(ctx, userPoolClient.AnalyticsConfiguration, &response.Diagnostics)
	data.AuthSessionValidity = flex.Int64ToFramework(ctx, userPoolClient.AuthSessionValidity)
	data.CallbackUrls = flex.FlattenFrameworkStringSet(ctx, userPoolClient.CallbackURLs)
	data.ClientSecret = flex.StringToFrameworkLegacy(ctx, userPoolClient.ClientSecret)
	data.DefaultRedirectUri = flex.StringToFrameworkLegacy(ctx, userPoolClient.DefaultRedirectURI)
	data.EnablePropagateAdditionalUserContextData = flex.BoolToFramework(ctx, userPoolClient.EnablePropagateAdditionalUserContextData)
	data.EnableTokenRevocation = flex.BoolToFramework(ctx, userPoolClient.EnableTokenRevocation)
	data.ExplicitAuthFlows = flex.FlattenFrameworkStringSet(ctx, userPoolClient.ExplicitAuthFlows)
	data.IdTokenValidity = flex.Int64ToFrameworkLegacy(ctx, userPoolClient.IdTokenValidity)
	data.LogoutUrls = flex.FlattenFrameworkStringSet(ctx, userPoolClient.LogoutURLs)
	data.Name = flex.StringToFramework(ctx, userPoolClient.ClientName)
	data.PreventUserExistenceErrors = flex.StringToFrameworkLegacy(ctx, userPoolClient.PreventUserExistenceErrors)
	data.ReadAttributes = flex.FlattenFrameworkStringSet(ctx, userPoolClient.ReadAttributes)
	data.RefreshTokenValidity = flex.Int64ToFramework(ctx, userPoolClient.RefreshTokenValidity)
	data.SupportedIdentityProviders = flex.FlattenFrameworkStringSet(ctx, userPoolClient.SupportedIdentityProviders)
	data.TokenValidityUnits = flattenTokenValidityUnits(ctx, userPoolClient.TokenValidityUnits)
	data.UserPoolID = flex.StringToFramework(ctx, userPoolClient.UserPoolId)
	data.WriteAttributes = flex.FlattenFrameworkStringSet(ctx, userPoolClient.WriteAttributes)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

// Update is called to update the state of the resource.
// Config, planned state, and prior state values should be read from the UpdateRequest and new state values set on the UpdateResponse.
func (r *resourceManagedUserPoolClient) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourceManagedUserPoolClientData

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

// Delete is called when the provider must delete the resource.
// Config values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically call DeleteResponse.State.RemoveResource(),
// so it can be omitted from provider logic.
func (r *resourceManagedUserPoolClient) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceManagedUserPoolClientData

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting TODO", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

// ImportState is called when the provider must import the state of a resource instance.
// This method must return enough state so the Read method can properly refresh the full resource.
//
// If setting an attribute with the import identifier, it is recommended to use the ImportStatePassthroughID() call in this method.
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
	AccessTokenValidity                      types.Int64               `tfsdk:"access_token_validity"`
	AllowedOauthFlows                        types.Set                 `tfsdk:"allowed_oauth_flows"`
	AllowedOauthFlowsUserPoolClient          types.Bool                `tfsdk:"allowed_oauth_flows_user_pool_client"`
	AllowedOauthScopes                       types.Set                 `tfsdk:"allowed_oauth_scopes"`
	AnalyticsConfiguration                   []*analyticsConfiguration `tfsdk:"analytics_configuration"`
	AuthSessionValidity                      types.Int64               `tfsdk:"auth_session_validity"`
	CallbackUrls                             types.Set                 `tfsdk:"callback_urls"`
	ClientSecret                             types.String              `tfsdk:"client_secret"`
	DefaultRedirectUri                       types.String              `tfsdk:"default_redirect_uri"`
	EnablePropagateAdditionalUserContextData types.Bool                `tfsdk:"enable_propagate_additional_user_context_data"`
	EnableTokenRevocation                    types.Bool                `tfsdk:"enable_token_revocation"`
	ExplicitAuthFlows                        types.Set                 `tfsdk:"explicit_auth_flows"`
	ID                                       types.String              `tfsdk:"id"`
	IdTokenValidity                          types.Int64               `tfsdk:"id_token_validity"`
	LogoutUrls                               types.Set                 `tfsdk:"logout_urls"`
	Name                                     types.String              `tfsdk:"name"`
	NamePrefix                               types.String              `tfsdk:"name_prefix"`
	PreventUserExistenceErrors               types.String              `tfsdk:"prevent_user_existence_errors"`
	ReadAttributes                           types.Set                 `tfsdk:"read_attributes"`
	RefreshTokenValidity                     types.Int64               `tfsdk:"refresh_token_validity"`
	SupportedIdentityProviders               types.Set                 `tfsdk:"supported_identity_providers"`
	TokenValidityUnits                       []*tokenValidityUnits     `tfsdk:"token_validity_units"`
	UserPoolID                               types.String              `tfsdk:"user_pool_id"`
	WriteAttributes                          types.Set                 `tfsdk:"write_attributes"`
}
