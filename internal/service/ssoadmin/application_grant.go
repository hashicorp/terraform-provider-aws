// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_application_grant", name="Application Grant")
// @IdentityAttribute("application_arn")
// @IdentityAttribute("grant_type")
// @ArnFormat(global=true)
// @ImportIDHandler("applicationGrantImportID")
// @Testing(preCheck="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstances")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="application_arn;grant_type", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newApplicationGrantResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &applicationGrantResource{}, nil
}

const (
	ResNameApplicationGrant = "Application Grant"

	applicationGrantIDPartCount = 2
)

type applicationGrantResource struct {
	framework.ResourceWithModel[applicationGrantResourceModel]
	framework.WithImportByIdentity
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
					listvalidator.SizeAtMost(1),
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
										Optional:    true,
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

	applicationARN := plan.ApplicationARN.ValueString()
	grantType := plan.GrantType.ValueString()

	grant, diags := expandApplicationGrant(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.PutApplicationGrantInput{
		ApplicationArn: aws.String(applicationARN),
		Grant:          grant,
		GrantType:      awstypes.GrantType(grantType),
	}

	_, err := conn.PutApplicationGrant(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationGrant, applicationARN, err),
			err.Error(),
		)
		return
	}

	idParts := []string{applicationARN, grantType}
	id, err := intflex.FlattenResourceId(idParts, applicationGrantIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameApplicationGrant, applicationARN, err),
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
	state.GrantType = fwtypes.StringEnumValue(awstypes.GrantType(parts[1]))

	grantValue, diags := flattenApplicationGrant(ctx, out.Grant)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Grant = grantValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan applicationGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	grant, diags := expandApplicationGrant(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.PutApplicationGrantInput{
		ApplicationArn: aws.String(plan.ApplicationARN.ValueString()),
		Grant:          grant,
		GrantType:      awstypes.GrantType(plan.GrantType.ValueString()),
	}

	_, err := conn.PutApplicationGrant(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameApplicationGrant, plan.ID.String(), err),
			err.Error(),
		)
		return
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

	parts, err := intflex.ExpandResourceId(state.ID.ValueString(), applicationGrantIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameApplicationGrant, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	in := &ssoadmin.DeleteApplicationGrantInput{
		ApplicationArn: aws.String(parts[0]),
		GrantType:      awstypes.GrantType(parts[1]),
	}

	_, err = conn.DeleteApplicationGrant(ctx, in)
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

func expandApplicationGrant(ctx context.Context, data *applicationGrantResourceModel) (awstypes.Grant, diag.Diagnostics) {
	var diags diag.Diagnostics
	grantType := awstypes.GrantType(data.GrantType.ValueString())

	switch grantType {
	case awstypes.GrantTypeRefreshToken:
		return &awstypes.GrantMemberRefreshToken{}, diags
	case awstypes.GrantTypeTokenExchange:
		return &awstypes.GrantMemberTokenExchange{}, diags
	}

	if data.Grant.IsNull() || data.Grant.IsUnknown() {
		diags.AddError(
			"Missing grant configuration",
			"The grant block is required when grant_type is "+string(grantType),
		)
		return nil, diags
	}

	grantData, d := data.Grant.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	switch grantType {
	case awstypes.GrantTypeAuthorizationCode:
		if grantData.AuthorizationCode.IsNull() || grantData.AuthorizationCode.IsUnknown() {
			diags.AddError(
				"Missing authorization_code configuration",
				"The authorization_code block is required when grant_type is authorization_code",
			)
			return nil, diags
		}
		acData, d := grantData.AuthorizationCode.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return &awstypes.GrantMemberAuthorizationCode{
			Value: awstypes.AuthorizationCodeGrant{
				RedirectUris: flex.ExpandFrameworkStringValueList(ctx, acData.RedirectUris),
			},
		}, diags

	case awstypes.GrantTypeJwtBearer:
		if grantData.JwtBearer.IsNull() || grantData.JwtBearer.IsUnknown() {
			diags.AddError(
				"Missing jwt_bearer configuration",
				"The jwt_bearer block is required when grant_type is "+string(awstypes.GrantTypeJwtBearer),
			)
			return nil, diags
		}
		jbData, d := grantData.JwtBearer.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		issuers, d := jbData.AuthorizedTokenIssuers.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var tokenIssuers []awstypes.AuthorizedTokenIssuer
		for _, issuer := range issuers {
			ti := awstypes.AuthorizedTokenIssuer{
				TrustedTokenIssuerArn: issuer.TrustedTokenIssuerArn.ValueStringPointer(),
			}
			if !issuer.AuthorizedAudiences.IsNull() {
				ti.AuthorizedAudiences = flex.ExpandFrameworkStringValueList(ctx, issuer.AuthorizedAudiences)
			}
			tokenIssuers = append(tokenIssuers, ti)
		}
		return &awstypes.GrantMemberJwtBearer{
			Value: awstypes.JwtBearerGrant{
				AuthorizedTokenIssuers: tokenIssuers,
			},
		}, diags
	}

	diags.AddError(
		"Unsupported grant type",
		"Unsupported grant_type: "+string(grantType),
	)
	return nil, diags
}

func flattenApplicationGrant(ctx context.Context, grant awstypes.Grant) (fwtypes.ListNestedObjectValueOf[grantModel], diag.Diagnostics) {
	var diags diag.Diagnostics

	switch v := grant.(type) {
	case *awstypes.GrantMemberAuthorizationCode:
		acModel := &authorizationCodeGrantModel{
			RedirectUris: flex.FlattenFrameworkStringValueListOfString(ctx, v.Value.RedirectUris),
		}
		gm := &grantModel{
			AuthorizationCode: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, acModel),
			JwtBearer:         fwtypes.NewListNestedObjectValueOfNull[jwtBearerGrantModel](ctx),
		}
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, gm), diags

	case *awstypes.GrantMemberJwtBearer:
		var issuers []*authorizedTokenIssuerModel
		for _, issuer := range v.Value.AuthorizedTokenIssuers {
			issuers = append(issuers, &authorizedTokenIssuerModel{
				AuthorizedAudiences:   flex.FlattenFrameworkStringValueListOfString(ctx, issuer.AuthorizedAudiences),
				TrustedTokenIssuerArn: flex.StringToFrameworkARN(ctx, issuer.TrustedTokenIssuerArn),
			})
		}
		jbModel := &jwtBearerGrantModel{
			AuthorizedTokenIssuers: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, issuers),
		}
		gm := &grantModel{
			AuthorizationCode: fwtypes.NewListNestedObjectValueOfNull[authorizationCodeGrantModel](ctx),
			JwtBearer:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, jbModel),
		}
		return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, gm), diags

	case *awstypes.GrantMemberRefreshToken, *awstypes.GrantMemberTokenExchange:
		return fwtypes.NewListNestedObjectValueOfNull[grantModel](ctx), diags

	default:
		return fwtypes.NewListNestedObjectValueOfNull[grantModel](ctx), diags
	}
}

type applicationGrantResourceModel struct {
	framework.WithRegionModel
	ApplicationARN fwtypes.ARN                                 `tfsdk:"application_arn"`
	Grant          fwtypes.ListNestedObjectValueOf[grantModel] `tfsdk:"grant"`
	GrantType      fwtypes.StringEnum[awstypes.GrantType]      `tfsdk:"grant_type"`
	ID             types.String                                `tfsdk:"id"`
}

type grantModel struct {
	AuthorizationCode fwtypes.ListNestedObjectValueOf[authorizationCodeGrantModel] `tfsdk:"authorization_code"`
	JwtBearer         fwtypes.ListNestedObjectValueOf[jwtBearerGrantModel]         `tfsdk:"jwt_bearer"`
}

type authorizationCodeGrantModel struct {
	RedirectUris fwtypes.ListOfString `tfsdk:"redirect_uris"`
}

type jwtBearerGrantModel struct {
	AuthorizedTokenIssuers fwtypes.ListNestedObjectValueOf[authorizedTokenIssuerModel] `tfsdk:"authorized_token_issuers"`
}

type authorizedTokenIssuerModel struct {
	AuthorizedAudiences   fwtypes.ListOfString `tfsdk:"authorized_audiences"`
	TrustedTokenIssuerArn fwtypes.ARN          `tfsdk:"trusted_token_issuer_arn"`
}

var _ inttypes.ImportIDParser = applicationGrantImportID{}

type applicationGrantImportID struct{}

func (applicationGrantImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, applicationGrantIDPartCount, false)
	if err != nil {
		return "", nil, fmt.Errorf("id %q should be in the format <application-arn>%s<grant-type>: %w", id, intflex.ResourceIdSeparator, err)
	}

	result := map[string]any{
		"application_arn": parts[0],
		"grant_type":      parts[1],
	}

	return id, result, nil
}
