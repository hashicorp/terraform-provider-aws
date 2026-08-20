// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_application_grant", name="Application Grant")
func newApplicationGrantResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationGrantResource{}, nil
}

const (
	ResNameApplicationGrant = "Application Grant"

	applicationGrantIDPartCount = 2
)

type applicationGrantResource struct {
	framework.ResourceWithModel[applicationGrantResourceModel]
	framework.WithImportByID
}

func (r *applicationGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"grant_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.GrantType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"grant": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[grantModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"authorization_code": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authorizationCodeGrantModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"redirect_uris": schema.ListAttribute{
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Required:    true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"jwt_bearer": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[jwtBearerGrantModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"authorized_token_issuers": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[authorizedTokenIssuerModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtLeast(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"authorized_audiences": schema.ListAttribute{
													CustomType:  fwtypes.ListOfStringType,
													ElementType: types.StringType,
													Optional:    true,
												},
												"trusted_token_issuer_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Optional:   true,
												},
											},
										},
									},
								},
							},
						},
						"refresh_token": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[refreshTokenGrantModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
						"token_exchange": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[tokenExchangeGrantModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{},
						},
					},
				},
			},
		},
	}
}

func (r *applicationGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan applicationGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grantPtr, diags := plan.Grant.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant, d := grantPtr.expandGrant(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.PutApplicationGrantInput{
		ApplicationArn: plan.ApplicationARN.ValueStringPointer(),
		GrantType:      plan.GrantType.ValueEnum(),
		Grant:          grant,
	}

	_, err := conn.PutApplicationGrant(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationGrant, plan.ApplicationARN.String(), err),
			err.Error(),
		)
		return
	}

	idParts := []string{
		plan.ApplicationARN.ValueString(),
		string(plan.GrantType.ValueEnum()),
	}
	id, err := intflex.FlattenResourceId(idParts, applicationGrantIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationGrant, plan.ApplicationARN.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *applicationGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state applicationGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationGrantByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplicationGrant, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	parts, err := intflex.ExpandResourceId(state.ID.ValueString(), applicationGrantIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameApplicationGrant, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ApplicationARN = fwtypes.ARNValue(parts[0])
	state.GrantType = fwtypes.StringEnumValue[awstypes.GrantType](awstypes.GrantType(parts[1]))

	var grantData grantModel
	resp.Diagnostics.Append(grantData.flattenGrant(ctx, out.Grant)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Grant = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &grantData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan, state applicationGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Grant.Equal(state.Grant) {
		grantPtr, diags := plan.Grant.ToPtr(ctx)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		grant, d := grantPtr.expandGrant(ctx)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		in := &ssoadmin.PutApplicationGrantInput{
			ApplicationArn: plan.ApplicationARN.ValueStringPointer(),
			GrantType:      plan.GrantType.ValueEnum(),
			Grant:          grant,
		}

		_, err := conn.PutApplicationGrant(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameApplicationGrant, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *applicationGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state applicationGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.DeleteApplicationGrantInput{
		ApplicationArn: state.ApplicationARN.ValueStringPointer(),
		GrantType:      state.GrantType.ValueEnum(),
	}

	_, err := conn.DeleteApplicationGrant(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameApplicationGrant, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findApplicationGrantByID(ctx context.Context, conn *ssoadmin.Client, id string) (*ssoadmin.GetApplicationGrantOutput, error) {
	parts, err := intflex.ExpandResourceId(id, applicationGrantIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &ssoadmin.GetApplicationGrantInput{
		ApplicationArn: aws.String(parts[0]),
		GrantType:      awstypes.GrantType(parts[1]),
	}

	out, err := conn.GetApplicationGrant(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type applicationGrantResourceModel struct {
	framework.WithRegionModel
	ApplicationARN fwtypes.ARN                                    `tfsdk:"application_arn"`
	Grant          fwtypes.ListNestedObjectValueOf[grantModel]    `tfsdk:"grant"`
	GrantType      fwtypes.StringEnum[awstypes.GrantType]         `tfsdk:"grant_type"`
	ID             types.String                                   `tfsdk:"id"`
}

type grantModel struct {
	AuthorizationCode fwtypes.ListNestedObjectValueOf[authorizationCodeGrantModel] `tfsdk:"authorization_code"`
	JwtBearer         fwtypes.ListNestedObjectValueOf[jwtBearerGrantModel]         `tfsdk:"jwt_bearer"`
	RefreshToken      fwtypes.ListNestedObjectValueOf[refreshTokenGrantModel]      `tfsdk:"refresh_token"`
	TokenExchange     fwtypes.ListNestedObjectValueOf[tokenExchangeGrantModel]     `tfsdk:"token_exchange"`
}

var (
	_ flex.TypedExpander = grantModel{}
	_ flex.Flattener     = &grantModel{}
)

func (m grantModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if targetType == reflect.TypeFor[awstypes.Grant]() {
		grant, d := m.expandGrant(ctx)
		diags.Append(d...)
		return grant, diags
	}

	return nil, diags
}

func (m grantModel) expandGrant(ctx context.Context) (awstypes.Grant, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch {
	case !m.AuthorizationCode.IsNull():
		authCode, d := m.AuthorizationCode.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.GrantMemberAuthorizationCode
		diags.Append(flex.Expand(ctx, authCode, &r.Value)...)
		return &r, diags

	case !m.JwtBearer.IsNull():
		jwtBearer, d := m.JwtBearer.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.GrantMemberJwtBearer
		diags.Append(flex.Expand(ctx, jwtBearer, &r.Value)...)
		return &r, diags

	case !m.RefreshToken.IsNull():
		return &awstypes.GrantMemberRefreshToken{}, diags

	case !m.TokenExchange.IsNull():
		return &awstypes.GrantMemberTokenExchange{}, diags
	}

	return nil, diags
}

func (m *grantModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	return m.flattenGrant(ctx, v)
}

func (m *grantModel) flattenGrant(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Initialize all sub-blocks to null so Terraform sees a clean state.
	m.AuthorizationCode = fwtypes.NewListNestedObjectValueOfNull[authorizationCodeGrantModel](ctx)
	m.JwtBearer = fwtypes.NewListNestedObjectValueOfNull[jwtBearerGrantModel](ctx)
	m.RefreshToken = fwtypes.NewListNestedObjectValueOfNull[refreshTokenGrantModel](ctx)
	m.TokenExchange = fwtypes.NewListNestedObjectValueOfNull[tokenExchangeGrantModel](ctx)

	switch v := v.(type) {
	case *awstypes.GrantMemberAuthorizationCode:
		var authCode authorizationCodeGrantModel
		diags.Append(flex.Flatten(ctx, v.Value, &authCode)...)
		if diags.HasError() {
			return diags
		}
		m.AuthorizationCode = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &authCode)

	case *awstypes.GrantMemberJwtBearer:
		var jwtBearer jwtBearerGrantModel
		diags.Append(flex.Flatten(ctx, v.Value, &jwtBearer)...)
		if diags.HasError() {
			return diags
		}
		m.JwtBearer = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &jwtBearer)

	case *awstypes.GrantMemberRefreshToken:
		var refreshToken refreshTokenGrantModel
		m.RefreshToken = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &refreshToken)

	case *awstypes.GrantMemberTokenExchange:
		var tokenExchange tokenExchangeGrantModel
		m.TokenExchange = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tokenExchange)
	}

	return diags
}

type authorizationCodeGrantModel struct {
	RedirectURIs fwtypes.ListOfString `tfsdk:"redirect_uris"`
}

type jwtBearerGrantModel struct {
	AuthorizedTokenIssuers fwtypes.ListNestedObjectValueOf[authorizedTokenIssuerModel] `tfsdk:"authorized_token_issuers"`
}

type authorizedTokenIssuerModel struct {
	AuthorizedAudiences   fwtypes.ListOfString `tfsdk:"authorized_audiences"`
	TrustedTokenIssuerARN fwtypes.ARN          `tfsdk:"trusted_token_issuer_arn"`
}

type refreshTokenGrantModel struct{}

type tokenExchangeGrantModel struct{}
