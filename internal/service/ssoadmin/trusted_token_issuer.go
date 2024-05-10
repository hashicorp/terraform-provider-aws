// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Trusted Token Issuer")
// @Tags
func newResourceTrustedTokenIssuer(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceTrustedTokenIssuer{}, nil
}

const (
	ResNameTrustedTokenIssuer = "Trusted Token Issuer"
)

type resourceTrustedTokenIssuer struct {
	framework.ResourceWithConfigure
}

func (r *resourceTrustedTokenIssuer) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ssoadmin_trusted_token_issuer"
}

func (r *resourceTrustedTokenIssuer) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"client_token": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"trusted_token_issuer_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.TrustedTokenIssuerType](),
				},
			},

			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"trusted_token_issuer_configuration": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"oidc_jwt_configuration": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.IsRequired(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"claim_attribute_path": schema.StringAttribute{
										Required: true,
									},
									"identity_store_attribute_path": schema.StringAttribute{
										Required: true,
									},
									"issuer_url": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{ // Not part of OidcJwtUpdateConfiguration struct, have to recreate at change
											stringplanmodifier.RequiresReplace(),
										},
									},
									"jwks_retrieval_option": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											enum.FrameworkValidate[awstypes.JwksRetrievalOption](),
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

func (r *resourceTrustedTokenIssuer) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan resourceTrustedTokenIssuerData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.CreateTrustedTokenIssuerInput{
		InstanceArn:            aws.String(plan.InstanceARN.ValueString()),
		Name:                   aws.String(plan.Name.ValueString()),
		TrustedTokenIssuerType: awstypes.TrustedTokenIssuerType(plan.TrustedTokenIssuerType.ValueString()),
		Tags:                   getTagsIn(ctx),
	}

	if !plan.ClientToken.IsNull() {
		in.ClientToken = aws.String(plan.ClientToken.ValueString())
	}

	if !plan.TrustedTokenIssuerConfiguration.IsNull() {
		var tfList []TrustedTokenIssuerConfigurationData
		resp.Diagnostics.Append(plan.TrustedTokenIssuerConfiguration.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		trustedTokenIssuerConfiguration, d := expandTrustedTokenIssuerConfiguration(ctx, tfList)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.TrustedTokenIssuerConfiguration = trustedTokenIssuerConfiguration
	}

	out, err := conn.CreateTrustedTokenIssuer(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameTrustedTokenIssuer, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionCreating, ResNameTrustedTokenIssuer, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.TrustedTokenIssuerArn)
	plan.ID = flex.StringToFramework(ctx, out.TrustedTokenIssuerArn)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceTrustedTokenIssuer) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceTrustedTokenIssuerData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findTrustedTokenIssuerByARN(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameTrustedTokenIssuer, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	instanceARN, _ := TrustedTokenIssuerParseInstanceARN(r.Meta(), aws.ToString(out.TrustedTokenIssuerArn))

	state.ARN = flex.StringToFramework(ctx, out.TrustedTokenIssuerArn)
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.ID = flex.StringToFramework(ctx, out.TrustedTokenIssuerArn)
	state.InstanceARN = flex.StringToFrameworkARN(ctx, aws.String(instanceARN))
	state.TrustedTokenIssuerType = flex.StringValueToFramework(ctx, out.TrustedTokenIssuerType)

	trustedTokenIssuerConfiguration, d := flattenTrustedTokenIssuerConfiguration(ctx, out.TrustedTokenIssuerConfiguration)
	resp.Diagnostics.Append(d...)
	state.TrustedTokenIssuerConfiguration = trustedTokenIssuerConfiguration

	// listTags requires both trusted token issuer and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	tags, err := listTags(ctx, conn, state.ARN.ValueString(), state.InstanceARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionSetting, ResNameTrustedTokenIssuer, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	setTagsOut(ctx, Tags(tags))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceTrustedTokenIssuer) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var plan, state resourceTrustedTokenIssuerData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) || !plan.TrustedTokenIssuerConfiguration.Equal(state.TrustedTokenIssuerConfiguration) {
		in := &ssoadmin.UpdateTrustedTokenIssuerInput{
			TrustedTokenIssuerArn: aws.String(plan.ID.ValueString()),
		}

		if !plan.Name.IsNull() {
			in.Name = aws.String(plan.Name.ValueString())
		}

		if !plan.TrustedTokenIssuerConfiguration.IsNull() {
			var tfList []TrustedTokenIssuerConfigurationData
			resp.Diagnostics.Append(plan.TrustedTokenIssuerConfiguration.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			trustedTokenIssuerUpdateConfiguration, d := expandTrustedTokenIssuerUpdateConfiguration(ctx, tfList)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			in.TrustedTokenIssuerConfiguration = trustedTokenIssuerUpdateConfiguration
		}

		out, err := conn.UpdateTrustedTokenIssuer(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameTrustedTokenIssuer, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameTrustedTokenIssuer, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	// updateTags requires both trusted token issuer and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	if oldTagsAll, newTagsAll := state.TagsAll, plan.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, plan.ID.ValueString(), plan.InstanceARN.ValueString(), oldTagsAll, newTagsAll); err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionUpdating, ResNameTrustedTokenIssuer, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceTrustedTokenIssuer) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SSOAdminClient(ctx)

	var state resourceTrustedTokenIssuerData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ssoadmin.DeleteTrustedTokenIssuerInput{
		TrustedTokenIssuerArn: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteTrustedTokenIssuer(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionDeleting, ResNameTrustedTokenIssuer, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceTrustedTokenIssuer) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceTrustedTokenIssuer) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func findTrustedTokenIssuerByARN(ctx context.Context, conn *ssoadmin.Client, arn string) (*ssoadmin.DescribeTrustedTokenIssuerOutput, error) {
	in := &ssoadmin.DescribeTrustedTokenIssuerInput{
		TrustedTokenIssuerArn: aws.String(arn),
	}

	out, err := conn.DescribeTrustedTokenIssuer(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandTrustedTokenIssuerConfiguration(ctx context.Context, tfList []TrustedTokenIssuerConfigurationData) (awstypes.TrustedTokenIssuerConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	var OIDCJWTConfigurationData []OIDCJWTConfigurationData
	diags.Append(tfObj.OIDCJWTConfiguration.ElementsAs(ctx, &OIDCJWTConfigurationData, false)...)

	apiObject := &awstypes.TrustedTokenIssuerConfigurationMemberOidcJwtConfiguration{
		Value: *expandOIDCJWTConfiguration(OIDCJWTConfigurationData),
	}

	return apiObject, diags
}

func expandOIDCJWTConfiguration(tfList []OIDCJWTConfigurationData) *awstypes.OidcJwtConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]

	apiObject := &awstypes.OidcJwtConfiguration{
		ClaimAttributePath:         aws.String(tfObj.ClaimAttributePath.ValueString()),
		IdentityStoreAttributePath: aws.String(tfObj.IdentityStoreAttributePath.ValueString()),
		IssuerUrl:                  aws.String(tfObj.IssuerUrl.ValueString()),
		JwksRetrievalOption:        awstypes.JwksRetrievalOption(tfObj.JWKSRetrievalOption.ValueString()),
	}

	return apiObject
}

func expandTrustedTokenIssuerUpdateConfiguration(ctx context.Context, tfList []TrustedTokenIssuerConfigurationData) (awstypes.TrustedTokenIssuerUpdateConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	var OIDCJWTConfigurationData []OIDCJWTConfigurationData
	diags.Append(tfObj.OIDCJWTConfiguration.ElementsAs(ctx, &OIDCJWTConfigurationData, false)...)

	apiObject := &awstypes.TrustedTokenIssuerUpdateConfigurationMemberOidcJwtConfiguration{
		Value: *expandOIDCJWTUpdateConfiguration(OIDCJWTConfigurationData),
	}

	return apiObject, diags
}

func expandOIDCJWTUpdateConfiguration(tfList []OIDCJWTConfigurationData) *awstypes.OidcJwtUpdateConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]

	apiObject := &awstypes.OidcJwtUpdateConfiguration{
		ClaimAttributePath:         aws.String(tfObj.ClaimAttributePath.ValueString()),
		IdentityStoreAttributePath: aws.String(tfObj.IdentityStoreAttributePath.ValueString()),
		JwksRetrievalOption:        awstypes.JwksRetrievalOption(tfObj.JWKSRetrievalOption.ValueString()),
	}

	return apiObject
}

func flattenTrustedTokenIssuerConfiguration(ctx context.Context, apiObject awstypes.TrustedTokenIssuerConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: TrustedTokenIssuerConfigurationAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{}

	switch v := apiObject.(type) {
	case *awstypes.TrustedTokenIssuerConfigurationMemberOidcJwtConfiguration:
		oidcJWTConfiguration, d := flattenOIDCJWTConfiguration(ctx, &v.Value)
		obj["oidc_jwt_configuration"] = oidcJWTConfiguration
		diags.Append(d...)
	default:
		log.Println("union is nil or unknown type")
	}

	objVal, d := types.ObjectValue(TrustedTokenIssuerConfigurationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenOIDCJWTConfiguration(ctx context.Context, apiObject *awstypes.OidcJwtConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: OIDCJWTConfigurationAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"claim_attribute_path":          flex.StringToFramework(ctx, apiObject.ClaimAttributePath),
		"identity_store_attribute_path": flex.StringToFramework(ctx, apiObject.IdentityStoreAttributePath),
		"issuer_url":                    flex.StringToFramework(ctx, apiObject.IssuerUrl),
		"jwks_retrieval_option":         flex.StringValueToFramework(ctx, apiObject.JwksRetrievalOption),
	}

	objVal, d := types.ObjectValue(OIDCJWTConfigurationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

// Instance ARN is not returned by DescribeTrustedTokenIssuer but is needed for schema consistency when importing and tagging.
// Instance ARN can be extracted from the Trusted Token Issuer ARN.
func TrustedTokenIssuerParseInstanceARN(conn *conns.AWSClient, id string) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	parts := strings.Split(id, "/")

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return fmt.Sprintf("arn:%s:sso:::instance/%s", conn.Partition, parts[1]), diags
	}

	return "", diags
}

type resourceTrustedTokenIssuerData struct {
	ARN                             types.String `tfsdk:"arn"`
	ClientToken                     types.String `tfsdk:"client_token"`
	ID                              types.String `tfsdk:"id"`
	InstanceARN                     fwtypes.ARN  `tfsdk:"instance_arn"`
	Name                            types.String `tfsdk:"name"`
	TrustedTokenIssuerConfiguration types.List   `tfsdk:"trusted_token_issuer_configuration"`
	TrustedTokenIssuerType          types.String `tfsdk:"trusted_token_issuer_type"`
	Tags                            types.Map    `tfsdk:"tags"`
	TagsAll                         types.Map    `tfsdk:"tags_all"`
}

type TrustedTokenIssuerConfigurationData struct {
	OIDCJWTConfiguration types.List `tfsdk:"oidc_jwt_configuration"`
}

type OIDCJWTConfigurationData struct {
	ClaimAttributePath         types.String `tfsdk:"claim_attribute_path"`
	IdentityStoreAttributePath types.String `tfsdk:"identity_store_attribute_path"`
	IssuerUrl                  types.String `tfsdk:"issuer_url"`
	JWKSRetrievalOption        types.String `tfsdk:"jwks_retrieval_option"`
}

var TrustedTokenIssuerConfigurationAttrTypes = map[string]attr.Type{
	"oidc_jwt_configuration": types.ListType{ElemType: types.ObjectType{AttrTypes: OIDCJWTConfigurationAttrTypes}},
}

var OIDCJWTConfigurationAttrTypes = map[string]attr.Type{
	"claim_attribute_path":          types.StringType,
	"identity_store_attribute_path": types.StringType,
	"issuer_url":                    types.StringType,
	"jwks_retrieval_option":         types.StringType,
}
