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
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	timeUnitsType = fwtypes.StringEnumType[awstypes.TimeUnitsType]()
)

// @FrameworkResource("aws_cognito_user_pool_client", name="User Pool Client")
func newUserPoolClientResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &userPoolClientResource{}

	return r, nil
}

type userPoolClientResource struct {
	framework.ResourceWithConfigure
}

// Schema returns the schema for this resource.
func (r *userPoolClientResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
			"generate_secret": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
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
				Required:   true,
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

func (r *userPoolClientResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var config resourceUserPoolClientModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceUserPoolClientModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var input cognitoidentityprovider.CreateUserPoolClientInput
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input, fwflex.WithFieldNamePrefix("Client"))...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := conn.CreateUserPoolClient(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("creating Cognito User Pool Client (%s)", plan.Name.ValueString()),
			err.Error(),
		)
		return
	}

	poolClient := resp.UserPoolClient

	response.Diagnostics.Append(fwflex.Flatten(ctx, poolClient, &config, fwflex.WithFieldNamePrefix("Client"))...)
	config.TokenValidityUnits = flattenTokenValidityUnits(ctx, poolClient.TokenValidityUnits, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &config)...)
}

func (r *userPoolClientResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resourceUserPoolClientModel
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
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User Pool Client (%s)", state.ID.ValueString()), err.Error())
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

func (r *userPoolClientResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var config resourceUserPoolClientModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan resourceUserPoolClientModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var state resourceUserPoolClientModel
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
			fmt.Sprintf("updating Cognito User Pool Client (%s)", plan.ID.ValueString()),
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

func (r *userPoolClientResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resourceUserPoolClientModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := state.deleteInput(ctx)

	tflog.Debug(ctx, "deleting Cognito User Pool Client", map[string]any{
		names.AttrID:         state.ID.ValueString(),
		names.AttrUserPoolID: state.UserPoolID.ValueString(),
	})

	conn := r.Meta().CognitoIDPClient(ctx)

	_, err := conn.DeleteUserPoolClient(ctx, params)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func (r *userPoolClientResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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

func (r *userPoolClientResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
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

func findUserPoolClientByTwoPartKey(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, clientID string) (*awstypes.UserPoolClientType, error) {
	input := &cognitoidentityprovider.DescribeUserPoolClientInput{
		ClientId:   aws.String(clientID),
		UserPoolId: aws.String(userPoolID),
	}

	output, err := conn.DescribeUserPoolClient(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.UserPoolClient == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.UserPoolClient, nil
}

type resourceUserPoolClientModel struct {
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
	GenerateSecret                           types.Bool                                                   `tfsdk:"generate_secret"`
	ID                                       types.String                                                 `tfsdk:"id"`
	IdTokenValidity                          types.Int64                                                  `tfsdk:"id_token_validity" autoflex:",legacy"`
	LogoutUrls                               types.Set                                                    `tfsdk:"logout_urls" autoflex:",legacy"`
	Name                                     types.String                                                 `tfsdk:"name"`
	PreventUserExistenceErrors               fwtypes.StringEnum[awstypes.PreventUserExistenceErrorTypes]  `tfsdk:"prevent_user_existence_errors" autoflex:",legacy"`
	ReadAttributes                           types.Set                                                    `tfsdk:"read_attributes" autoflex:",legacy"`
	RefreshTokenValidity                     types.Int64                                                  `tfsdk:"refresh_token_validity"`
	SupportedIdentityProviders               types.Set                                                    `tfsdk:"supported_identity_providers" autoflex:",legacy"`
	TokenValidityUnits                       fwtypes.ListNestedObjectValueOf[tokenValidityUnitsModel]     `tfsdk:"token_validity_units"`
	UserPoolID                               types.String                                                 `tfsdk:"user_pool_id"`
	WriteAttributes                          types.Set                                                    `tfsdk:"write_attributes" autoflex:",legacy"`
}

func (data resourceUserPoolClientModel) deleteInput(ctx context.Context) *cognitoidentityprovider.DeleteUserPoolClientInput {
	return &cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   fwflex.StringFromFramework(ctx, data.ID),
		UserPoolId: fwflex.StringFromFramework(ctx, data.UserPoolID),
	}
}

type analyticsConfigurationModel struct {
	ApplicationARN fwtypes.ARN  `tfsdk:"application_arn"`
	ApplicationID  types.String `tfsdk:"application_id"`
	ExternalID     types.String `tfsdk:"external_id"`
	RoleARN        fwtypes.ARN  `tfsdk:"role_arn"`
	UserDataShared types.Bool   `tfsdk:"user_data_shared"`
}

type tokenValidityUnitsModel struct {
	AccessToken  fwtypes.StringEnum[awstypes.TimeUnitsType] `tfsdk:"access_token"`
	IdToken      fwtypes.StringEnum[awstypes.TimeUnitsType] `tfsdk:"id_token"`
	RefreshToken fwtypes.StringEnum[awstypes.TimeUnitsType] `tfsdk:"refresh_token"`
}

func isDefaultTokenValidityUnits(tvu *awstypes.TokenValidityUnitsType) bool {
	if tvu == nil {
		return false
	}
	return tvu.AccessToken == awstypes.TimeUnitsTypeHours &&
		tvu.IdToken == awstypes.TimeUnitsTypeHours &&
		tvu.RefreshToken == awstypes.TimeUnitsTypeDays
}

func resolveTokenValidityUnits(ctx context.Context, list fwtypes.ListNestedObjectValueOf[tokenValidityUnitsModel], diags *diag.Diagnostics) *tokenValidityUnitsModel {
	var units []tokenValidityUnitsModel
	diags.Append(list.ElementsAs(ctx, &units, false)...)
	if diags.HasError() {
		return nil
	}

	if len(units) == 1 {
		return &units[0]
	}
	return nil
}

func flattenTokenValidityUnits(ctx context.Context, tvu *awstypes.TokenValidityUnitsType, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[tokenValidityUnitsModel] {
	if tvu == nil || (tvu.AccessToken == "" && tvu.IdToken == "" && tvu.RefreshToken == "") {
		return fwtypes.NewListNestedObjectValueOfNull[tokenValidityUnitsModel](ctx)
	}

	var result tokenValidityUnitsModel
	diags.Append(fwflex.Flatten(ctx, tvu, &result)...)

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &result)
}

var _ resource.ConfigValidator = &resourceUserPoolClientAccessTokenValidityValidator{}

type resourceUserPoolClientAccessTokenValidityValidator struct {
	resourceUserPoolClientValidityValidator
}

func (v resourceUserPoolClientAccessTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceUserPoolClientModel) types.Int64 {
			return rupcd.AccessTokenValidity
		},
		func(tvu *tokenValidityUnitsModel) awstypes.TimeUnitsType {
			return tvu.AccessToken.ValueEnum()
		},
	)
}

var _ resource.ConfigValidator = &resourceUserPoolClientIDTokenValidityValidator{}

type resourceUserPoolClientIDTokenValidityValidator struct {
	resourceUserPoolClientValidityValidator
}

func (v resourceUserPoolClientIDTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceUserPoolClientModel) types.Int64 {
			return rupcd.IdTokenValidity
		},
		func(tvu *tokenValidityUnitsModel) awstypes.TimeUnitsType {
			return tvu.IdToken.ValueEnum()
		},
	)
}

var _ resource.ConfigValidator = &resourceUserPoolClientRefreshTokenValidityValidator{}

type resourceUserPoolClientRefreshTokenValidityValidator struct {
	resourceUserPoolClientValidityValidator
}

func (v resourceUserPoolClientRefreshTokenValidityValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	v.validate(ctx, req, resp,
		func(rupcd resourceUserPoolClientModel) types.Int64 {
			return rupcd.RefreshTokenValidity
		},
		func(tvu *tokenValidityUnitsModel) awstypes.TimeUnitsType {
			return tvu.RefreshToken.ValueEnum()
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

func (v resourceUserPoolClientValidityValidator) validate(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse, valF func(resourceUserPoolClientModel) types.Int64, unitF func(*tokenValidityUnitsModel) awstypes.TimeUnitsType) {
	var config resourceUserPoolClientModel
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
