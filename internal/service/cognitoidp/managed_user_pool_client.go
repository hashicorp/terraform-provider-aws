// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_cognito_managed_user_pool_client", name="Managed User Pool Client")
func newManagedUserPoolClientResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &managedUserPoolClientResource{}, nil
}

type managedUserPoolClientResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

// Schema returns the schema for this resource.
func (r *managedUserPoolClientResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
						enum.FrameworkValidate[awstypes.OAuthFlowType](),
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
			names.AttrClientSecret: schema.StringAttribute{
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
						enum.FrameworkValidate[awstypes.ExplicitAuthFlowsType](),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
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
			names.AttrName: schema.StringAttribute{
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
						path.MatchRelative().AtParent().AtName(names.AttrNamePrefix),
						path.MatchRelative().AtParent().AtName("name_pattern"),
					),
				),
			},
			names.AttrNamePrefix: schema.StringAttribute{
				Optional:   true,
				Validators: userPoolClientNameValidator,
			},
			"prevent_user_existence_errors": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PreventUserExistenceErrorTypes](),
				Optional:   true,
				Computed:   true,
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
			names.AttrUserPoolID: schema.StringAttribute{
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[analyticsConfigurationModel](ctx),
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
									path.MatchRelative().AtParent().AtName(names.AttrApplicationID),
								),
								stringvalidator.ConflictsWith(
									path.MatchRelative().AtParent().AtName(names.AttrExternalID),
									path.MatchRelative().AtParent().AtName(names.AttrRoleARN),
								),
							},
						},
						names.AttrApplicationID: schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName(names.AttrExternalID),
									path.MatchRelative().AtParent().AtName(names.AttrRoleARN),
								),
							},
						},
						names.AttrExternalID: schema.StringAttribute{
							Optional: true,
						},
						names.AttrRoleARN: schema.StringAttribute{
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[tokenValidityUnitsModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"access_token": schema.StringAttribute{
							CustomType: timeUnitsType,
							Optional:   true,
							Computed:   true,
							Default:    timeUnitsType.AttributeDefault(awstypes.TimeUnitsTypeHours),
						},
						"id_token": schema.StringAttribute{
							CustomType: timeUnitsType,
							Optional:   true,
							Computed:   true,
							Default:    timeUnitsType.AttributeDefault(awstypes.TimeUnitsTypeHours),
						},
						"refresh_token": schema.StringAttribute{
							CustomType: timeUnitsType,
							Optional:   true,
							Computed:   true,
							Default:    timeUnitsType.AttributeDefault(awstypes.TimeUnitsTypeDays),
						},
					},
				},
			},
		},
	}

	response.Schema = s
}

func (r *managedUserPoolClientResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var config resourceManagedUserPoolClientModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceManagedUserPoolClientModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	filter := tfslices.PredicateTrue[*awstypes.UserPoolClientDescription]()
	if namePattern := plan.NamePattern; !namePattern.IsUnknown() && !namePattern.IsNull() {
		filter = func(v *awstypes.UserPoolClientDescription) bool {
			return namePattern.ValueRegexp().MatchString(aws.ToString(v.ClientName))
		}
	}
	if namePrefix := plan.NamePrefix; !namePrefix.IsUnknown() && !namePrefix.IsNull() {
		filter = func(v *awstypes.UserPoolClientDescription) bool {
			return strings.HasPrefix(aws.ToString(v.ClientName), namePrefix.ValueString())
		}
	}
	userPoolID := plan.UserPoolID.ValueString()

	poolClient, err := findUserPoolClientByName(ctx, conn, userPoolID, filter)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed User Pool Client (%s)", userPoolID), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, poolClient, &config, fwflex.WithFieldNamePrefix("Client"))...)
	config.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits, &response.Diagnostics)
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
		var input cognitoidentityprovider.UpdateUserPoolClientInput
		response.Diagnostics.Append(fwflex.Expand(ctx, config, &input, fwflex.WithFieldNamePrefix("Client"))...)
		if response.Diagnostics.HasError() {
			return
		}

		const (
			timeout = 2 * time.Minute
		)
		output, err := tfresource.RetryWhenIsA[*awstypes.ConcurrentModificationException](ctx, timeout, func() (any, error) {
			return conn.UpdateUserPoolClient(ctx, &input)
		})
		if err != nil {
			response.Diagnostics.AddError(
				fmt.Sprintf("updating Cognito Managed User Pool Client (%s)", plan.ID.ValueString()),
				err.Error(),
			)
			return
		}

		poolClient := output.(*cognitoidentityprovider.UpdateUserPoolClientOutput).UserPoolClient

		response.Diagnostics.Append(fwflex.Flatten(ctx, poolClient, &config, fwflex.WithFieldNamePrefix("Client"))...)
		config.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &config)...)
}

func (r *managedUserPoolClientResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resourceManagedUserPoolClientModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	poolClient, err := findUserPoolClientByTwoPartKey(ctx, conn, state.UserPoolID.ValueString(), state.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito Managed User Pool Client (%s)", state.ID.ValueString()), err.Error())

		return
	}

	tokenValidityUnitsNull := state.TokenValidityUnits.IsNull()

	response.Diagnostics.Append(fwflex.Flatten(ctx, poolClient, &state, fwflex.WithFieldNamePrefix("Client"))...)
	if tokenValidityUnitsNull && isDefaultTokenValidityUnits(poolClient.TokenValidityUnits) {
		state.TokenValidityUnits = fwtypes.NewListNestedObjectValueOfNull[tokenValidityUnitsModel](ctx)
	} else {
		state.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits, &response.Diagnostics)
	}

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *managedUserPoolClientResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var config resourceManagedUserPoolClientModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceManagedUserPoolClientModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state resourceManagedUserPoolClientModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	var input cognitoidentityprovider.UpdateUserPoolClientInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Client"))...)
	if response.Diagnostics.HasError() {
		return
	}

	// If removing `token_validity_units`, reset to defaults
	if !state.TokenValidityUnits.IsNull() && plan.TokenValidityUnits.IsNull() {
		input.TokenValidityUnits = &awstypes.TokenValidityUnitsType{
			AccessToken:  awstypes.TimeUnitsTypeHours,
			IdToken:      awstypes.TimeUnitsTypeHours,
			RefreshToken: awstypes.TimeUnitsTypeDays,
		}
	}

	const (
		timeout = 2 * time.Minute
	)
	output, err := tfresource.RetryWhenIsA[*awstypes.ConcurrentModificationException](ctx, timeout, func() (any, error) {
		return conn.UpdateUserPoolClient(ctx, &input)
	})
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("updating Cognito Managed User Pool Client (%s)", plan.ID.ValueString()),
			err.Error(),
		)
		return
	}

	poolClient := output.(*cognitoidentityprovider.UpdateUserPoolClientOutput).UserPoolClient

	response.Diagnostics.Append(fwflex.Flatten(ctx, poolClient, &config, fwflex.WithFieldNamePrefix("Client"))...)
	if !state.TokenValidityUnits.IsNull() && plan.TokenValidityUnits.IsNull() && isDefaultTokenValidityUnits(poolClient.TokenValidityUnits) {
		config.TokenValidityUnits = fwtypes.NewListNestedObjectValueOfNull[tokenValidityUnitsModel](ctx)
	} else {
		config.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits, &response.Diagnostics)
	}

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &config)...)
}

func (r *managedUserPoolClientResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, "/")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf("wrong format of import ID (%s), use: 'user-pool-id/client-id'", request.ID))
		return
	}
	userPoolId := parts[0]
	clientId := parts[1]
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), clientId)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrUserPoolID), userPoolId)...)
}

func (r *managedUserPoolClientResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
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

func findUserPoolClientByName(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID string, filter tfslices.Predicate[*awstypes.UserPoolClientDescription]) (*awstypes.UserPoolClientType, error) {
	input := &cognitoidentityprovider.ListUserPoolClientsInput{
		UserPoolId: aws.String(userPoolID),
	}
	var userPoolClients []awstypes.UserPoolClientDescription

	pages := cognitoidentityprovider.NewListUserPoolClientsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.UserPoolClients {
			if filter(&v) {
				userPoolClients = append(userPoolClients, v)
			}
		}
	}

	userPoolClient, err := tfresource.AssertSingleValueResult(userPoolClients)

	if err != nil {
		return nil, err
	}

	return findUserPoolClientByTwoPartKey(ctx, conn, userPoolID, aws.ToString(userPoolClient.ClientId))
}

type resourceManagedUserPoolClientModel struct {
	AccessTokenValidity                      types.Int64                                                  `tfsdk:"access_token_validity" autoflex:",legacy"`
	AllowedOauthFlows                        types.Set                                                    `tfsdk:"allowed_oauth_flows" autoflex:",legacy"`
	AllowedOauthFlowsUserPoolClient          types.Bool                                                   `tfsdk:"allowed_oauth_flows_user_pool_client"`
	AllowedOauthScopes                       types.Set                                                    `tfsdk:"allowed_oauth_scopes" autoflex:",legacy"`
	AnalyticsConfiguration                   fwtypes.ListNestedObjectValueOf[analyticsConfigurationModel] `tfsdk:"analytics_configuration"`
	AuthSessionValidity                      types.Int64                                                  `tfsdk:"auth_session_validity"`
	CallbackUrls                             types.Set                                                    `tfsdk:"callback_urls" autoflex:",legacy"`
	ClientSecret                             types.String                                                 `tfsdk:"client_secret" autoflex:",legacy"`
	DefaultRedirectUri                       types.String                                                 `tfsdk:"default_redirect_uri" autoflex:",legacy"`
	EnablePropagateAdditionalUserContextData types.Bool                                                   `tfsdk:"enable_propagate_additional_user_context_data"`
	EnableTokenRevocation                    types.Bool                                                   `tfsdk:"enable_token_revocation"`
	ExplicitAuthFlows                        types.Set                                                    `tfsdk:"explicit_auth_flows" autoflex:",legacy"`
	ID                                       types.String                                                 `tfsdk:"id"`
	IdTokenValidity                          types.Int64                                                  `tfsdk:"id_token_validity" autoflex:",legacy"`
	LogoutUrls                               types.Set                                                    `tfsdk:"logout_urls" autoflex:",legacy"`
	Name                                     types.String                                                 `tfsdk:"name"`
	NamePattern                              fwtypes.Regexp                                               `tfsdk:"name_pattern"`
	NamePrefix                               types.String                                                 `tfsdk:"name_prefix"`
	PreventUserExistenceErrors               fwtypes.StringEnum[awstypes.PreventUserExistenceErrorTypes]  `tfsdk:"prevent_user_existence_errors" autoflex:",legacy"`
	ReadAttributes                           types.Set                                                    `tfsdk:"read_attributes" autoflex:",legacy"`
	RefreshTokenValidity                     types.Int64                                                  `tfsdk:"refresh_token_validity"`
	SupportedIdentityProviders               types.Set                                                    `tfsdk:"supported_identity_providers" autoflex:",legacy"`
	TokenValidityUnits                       fwtypes.ListNestedObjectValueOf[tokenValidityUnitsModel]     `tfsdk:"token_validity_units"`
	UserPoolID                               types.String                                                 `tfsdk:"user_pool_id"`
	WriteAttributes                          types.Set                                                    `tfsdk:"write_attributes" autoflex:",legacy"`
}

var _ resource.ConfigValidator = &resourceManagedUserPoolClientAccessTokenValidityValidator{}

type resourceManagedUserPoolClientAccessTokenValidityValidator struct {
	resourceManagedUserPoolClientValidityValidator
}

func (v resourceManagedUserPoolClientAccessTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceManagedUserPoolClientModel) types.Int64 {
			return rupcd.AccessTokenValidity
		},
		func(tvu *tokenValidityUnitsModel) awstypes.TimeUnitsType {
			return tvu.AccessToken.ValueEnum()
		},
	)
}

var _ resource.ConfigValidator = &resourceManagedUserPoolClientIDTokenValidityValidator{}

type resourceManagedUserPoolClientIDTokenValidityValidator struct {
	resourceManagedUserPoolClientValidityValidator
}

func (v resourceManagedUserPoolClientIDTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceManagedUserPoolClientModel) types.Int64 {
			return rupcd.IdTokenValidity
		},
		func(tvu *tokenValidityUnitsModel) awstypes.TimeUnitsType {
			return tvu.IdToken.ValueEnum()
		},
	)
}

var _ resource.ConfigValidator = &resourceManagedUserPoolClientRefreshTokenValidityValidator{}

type resourceManagedUserPoolClientRefreshTokenValidityValidator struct {
	resourceManagedUserPoolClientValidityValidator
}

func (v resourceManagedUserPoolClientRefreshTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceManagedUserPoolClientModel) types.Int64 {
			return rupcd.RefreshTokenValidity
		},
		func(tvu *tokenValidityUnitsModel) awstypes.TimeUnitsType {
			return tvu.RefreshToken.ValueEnum()
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

func (v resourceManagedUserPoolClientValidityValidator) validate(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, valF func(resourceManagedUserPoolClientModel) types.Int64, unitF func(*tokenValidityUnitsModel) awstypes.TimeUnitsType) {
	var config resourceManagedUserPoolClientModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	x := valF(config)

	if x.IsUnknown() || x.IsNull() {
		return
	}

	var duration time.Duration

	units := resolveTokenValidityUnits(ctx, config.TokenValidityUnits, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if val := aws.ToInt64(fwflex.Int64FromFramework(ctx, x)); units == nil {
		duration = time.Duration(val * int64(v.defaultUnit))
	} else {
		switch unitF(units) {
		case awstypes.TimeUnitsTypeSeconds:
			duration = time.Duration(val * int64(time.Second))
		case awstypes.TimeUnitsTypeMinutes:
			duration = time.Duration(val * int64(time.Minute))
		case awstypes.TimeUnitsTypeHours:
			duration = time.Duration(val * int64(time.Hour))
		case awstypes.TimeUnitsTypeDays:
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
