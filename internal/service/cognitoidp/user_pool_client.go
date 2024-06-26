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

// @FrameworkResource(name="User Pool Client")
func newUserPoolClientResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &userPoolClientResource{}

	return r, nil
}

type userPoolClientResource struct {
	framework.ResourceWithConfigure
	userPoolClientResourceWithImport
	userPoolClientResourceWithConfigValidators
}

func (*userPoolClientResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_cognito_user_pool_client"
}

func (r *userPoolClientResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = userPoolClientResourceSchema(ctx)
}

func (r *userPoolClientResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data userPoolClientResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	name := data.ClientName.ValueString()
	input := &cognitoidentityprovider.CreateUserPoolClientInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateUserPoolClient(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Cognito User Pool Client (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.UserPoolClient, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *userPoolClientResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data userPoolClientResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	output, err := findUserPoolClientByTwoPartKey(ctx, conn, data.UserPoolID.ValueString(), data.ClientID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User Pool Client (%s)", data.ClientID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	// if isDefaultTokenValidityUnits(output.TokenValidityUnits) {
	// 	output.TokenValidityUnits = nil
	// }
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *userPoolClientResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new userPoolClientResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	input := &cognitoidentityprovider.UpdateUserPoolClientInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// If removing `token_validity_units`, reset to defaults.
	// if !old.TokenValidityUnits.IsNull() && new.TokenValidityUnits.IsNull() {
	// 	input.TokenValidityUnits.AccessToken = awstypes.TimeUnitsTypeHours
	// 	input.TokenValidityUnits.IdToken = awstypes.TimeUnitsTypeHours
	// 	input.TokenValidityUnits.RefreshToken = awstypes.TimeUnitsTypeDays
	// }

	const (
		timeout = 2 * time.Minute
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ConcurrentModificationException](ctx, timeout, func() (interface{}, error) {
		return conn.UpdateUserPoolClient(ctx, input)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Cognito User Pool Client (%s)", new.ClientID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputRaw.(*cognitoidentityprovider.UpdateUserPoolClientOutput).UserPoolClient, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *userPoolClientResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data userPoolClientResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CognitoIDPClient(ctx)

	tflog.Debug(ctx, "deleting Cognito User Pool Client", map[string]interface{}{
		names.AttrID:         data.ClientID.ValueString(),
		names.AttrUserPoolID: data.UserPoolID.ValueString(),
	})

	_, err := conn.DeleteUserPoolClient(ctx, &cognitoidentityprovider.DeleteUserPoolClientInput{
		ClientId:   fwflex.StringFromFramework(ctx, data.ClientID),
		UserPoolId: fwflex.StringFromFramework(ctx, data.UserPoolID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Cognito User Pool Client (%s)", data.ClientID.ValueString()), err.Error())

		return
	}
}

type userPoolClientResourceWithImport struct{}

func (r *userPoolClientResourceWithImport) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	const (
		separator = "/"
	)
	parts := strings.Split(request.ID, separator)

	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf("unexpected format for ID (%[1]s), expected UserPoolID%[2]sClientID", request.ID, separator))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), parts[1])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrUserPoolID), parts[0])...)
}

type userPoolClientResourceWithConfigValidators struct{}

func (r *userPoolClientResourceWithConfigValidators) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		userPoolClientResourceAccessTokenValidityValidator{
			userPoolClientResourceValidityValidator{
				attr:        "access_token_validity",
				min:         5 * time.Minute,
				max:         24 * time.Hour,
				defaultUnit: time.Hour,
			},
		},
		userPoolClientResourceIDTokenValidityValidator{
			userPoolClientResourceValidityValidator{
				attr:        "id_token_validity",
				min:         5 * time.Minute,
				max:         24 * time.Hour,
				defaultUnit: time.Hour,
			},
		},
		userPoolClientResourceRefreshTokenValidityValidator{
			userPoolClientResourceValidityValidator{
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

func userPoolClientResourceSchema(ctx context.Context) schema.Schema {
	timeUnitsType := fwtypes.StringEnumType[awstypes.TimeUnitsType]()

	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_token_validity": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"allowed_oauth_flows": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[analyticsConfigurationTypeModel](ctx),
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[tokenValidityUnitsTypeModel](ctx),
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
}

type userPoolClientResourceModel struct {
	AccessTokenValidity                      types.Int64                                                      `tfsdk:"access_token_validity"`
	AllowedOAuthFlows                        fwtypes.SetValueOf[types.String]                                 `tfsdk:"allowed_oauth_flows"`
	AllowedOAuthFlowsUserPoolClient          types.Bool                                                       `tfsdk:"allowed_oauth_flows_user_pool_client"`
	AllowedOAuthScopes                       fwtypes.SetValueOf[types.String]                                 `tfsdk:"allowed_oauth_scopes"`
	AnalyticsConfiguration                   fwtypes.ListNestedObjectValueOf[analyticsConfigurationTypeModel] `tfsdk:"analytics_configuration"`
	AuthSessionValidity                      types.Int64                                                      `tfsdk:"auth_session_validity"`
	CallbackURLs                             fwtypes.SetValueOf[types.String]                                 `tfsdk:"callback_urls"`
	ClientID                                 types.String                                                     `tfsdk:"id"`
	ClientName                               types.String                                                     `tfsdk:"name"`
	ClientSecret                             types.String                                                     `tfsdk:"client_secret"`
	DefaultRedirectURI                       types.String                                                     `tfsdk:"default_redirect_uri"`
	EnablePropagateAdditionalUserContextData types.Bool                                                       `tfsdk:"enable_propagate_additional_user_context_data"`
	EnableTokenRevocation                    types.Bool                                                       `tfsdk:"enable_token_revocation"`
	ExplicitAuthFlows                        fwtypes.SetValueOf[types.String]                                 `tfsdk:"explicit_auth_flows"`
	GenerateSecret                           types.Bool                                                       `tfsdk:"generate_secret"`
	IDTokenValidity                          types.Int64                                                      `tfsdk:"id_token_validity"`
	LogoutURLs                               fwtypes.SetValueOf[types.String]                                 `tfsdk:"logout_urls"`
	PreventUserExistenceErrors               fwtypes.StringEnum[awstypes.PreventUserExistenceErrorTypes]      `tfsdk:"prevent_user_existence_errors"`
	ReadAttributes                           fwtypes.SetValueOf[types.String]                                 `tfsdk:"read_attributes"`
	RefreshTokenValidity                     types.Int64                                                      `tfsdk:"refresh_token_validity"`
	SupportedIdentityProviders               fwtypes.SetValueOf[types.String]                                 `tfsdk:"supported_identity_providers"`
	TokenValidityUnits                       fwtypes.ListNestedObjectValueOf[tokenValidityUnitsTypeModel]     `tfsdk:"token_validity_units"`
	UserPoolID                               types.String                                                     `tfsdk:"user_pool_id"`
	WriteAttributes                          fwtypes.SetValueOf[types.String]                                 `tfsdk:"write_attributes"`
}

type analyticsConfigurationTypeModel struct {
	ApplicationARN fwtypes.ARN  `tfsdk:"application_arn"`
	ApplicationID  types.String `tfsdk:"application_id"`
	ExternalID     types.String `tfsdk:"external_id"`
	RoleARN        fwtypes.ARN  `tfsdk:"role_arn"`
	UserDataShared types.Bool   `tfsdk:"user_data_shared"`
}

type tokenValidityUnitsTypeModel struct {
	AccessToken  fwtypes.StringEnum[awstypes.TimeUnitsType] `tfsdk:"access_token"`
	IdToken      fwtypes.StringEnum[awstypes.TimeUnitsType] `tfsdk:"id_token"`
	RefreshToken fwtypes.StringEnum[awstypes.TimeUnitsType] `tfsdk:"refresh_token"`
}

func isDefaultTokenValidityUnits(apiObject *awstypes.TokenValidityUnitsType) bool {
	if apiObject == nil {
		return false
	}

	return apiObject.AccessToken == awstypes.TimeUnitsTypeHours &&
		apiObject.IdToken == awstypes.TimeUnitsTypeHours &&
		apiObject.RefreshToken == awstypes.TimeUnitsTypeDays
}

var _ resource.ConfigValidator = &userPoolClientResourceAccessTokenValidityValidator{}

type userPoolClientResourceAccessTokenValidityValidator struct {
	userPoolClientResourceValidityValidator
}

func (v userPoolClientResourceAccessTokenValidityValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	v.validate(ctx, request, response,
		func(v userPoolClientResourceModel) types.Int64 {
			return v.AccessTokenValidity
		},
		func(v *tokenValidityUnitsTypeModel) awstypes.TimeUnitsType {
			return v.AccessToken.ValueEnum()
		},
	)
}

var _ resource.ConfigValidator = &userPoolClientResourceIDTokenValidityValidator{}

type userPoolClientResourceIDTokenValidityValidator struct {
	userPoolClientResourceValidityValidator
}

func (v userPoolClientResourceIDTokenValidityValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	v.validate(ctx, request, response,
		func(v userPoolClientResourceModel) types.Int64 {
			return v.IDTokenValidity
		},
		func(v *tokenValidityUnitsTypeModel) awstypes.TimeUnitsType {
			return v.IdToken.ValueEnum()
		},
	)
}

var _ resource.ConfigValidator = &userPoolClientResourceRefreshTokenValidityValidator{}

type userPoolClientResourceRefreshTokenValidityValidator struct {
	userPoolClientResourceValidityValidator
}

func (v userPoolClientResourceRefreshTokenValidityValidator) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	v.validate(ctx, request, response,
		func(v userPoolClientResourceModel) types.Int64 {
			return v.RefreshTokenValidity
		},
		func(v *tokenValidityUnitsTypeModel) awstypes.TimeUnitsType {
			return v.RefreshToken.ValueEnum()
		},
	)
}

type userPoolClientResourceValidityValidator struct {
	min         time.Duration
	max         time.Duration
	attr        string
	defaultUnit time.Duration
}

func (v userPoolClientResourceValidityValidator) Description(ctx context.Context) string {
	return v.MarkdownDescription(ctx)
}

func (v userPoolClientResourceValidityValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("must have a duration between %s and %s", v.min, v.max)
}

func (v userPoolClientResourceValidityValidator) validate(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse, valF func(userPoolClientResourceModel) types.Int64, unitF func(*tokenValidityUnitsTypeModel) awstypes.TimeUnitsType) {
	var data userPoolClientResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	x := valF(data)

	if x.IsUnknown() || x.IsNull() {
		return
	}

	units, diags := data.TokenValidityUnits.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var duration time.Duration
	if val := x.ValueInt64(); units == nil {
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
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			path.Root(v.attr),
			v.Description(ctx),
			duration.String(),
		))
	}
}
