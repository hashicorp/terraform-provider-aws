// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ssoadmin

import (
	"context"
	"fmt"
	"reflect"
	"strings"

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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_trusted_token_issuer", name="Trusted Token Issuer")
// @Tags
// @ArnIdentity(identityDuplicateAttributes="id")
// @ArnFormat(global=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ssoadmin;ssoadmin.DescribeTrustedTokenIssuerOutput")
// @Testing(preCheckWithRegion="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.PreCheckSSOAdminInstancesWithRegion")
// @Testing(serialize=true)
// @Testing(preIdentityVersion="v5.100.0")
func newTrustedTokenIssuerResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &trustedTokenIssuerResource{}, nil
}

type trustedTokenIssuerResource struct {
	framework.ResourceWithModel[trustedTokenIssuerResourceModel]
	framework.WithImportByIdentity
}

func (r *trustedTokenIssuerResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
				CustomType: fwtypes.StringEnumType[awstypes.TrustedTokenIssuerType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"trusted_token_issuer_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[trustedTokenIssuerConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"oidc_jwt_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[oidcJWTConfigurationModel](ctx),
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
										CustomType: fwtypes.StringEnumType[awstypes.JwksRetrievalOption](),
										Required:   true,
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

func (r *trustedTokenIssuerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data trustedTokenIssuerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input ssoadmin.CreateTrustedTokenIssuerInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateTrustedTokenIssuer(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SSO Trusted Token Issuer (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.TrustedTokenIssuerARN = fwflex.StringToFramework(ctx, output.TrustedTokenIssuerArn)
	data.ID = data.TrustedTokenIssuerARN

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *trustedTokenIssuerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data trustedTokenIssuerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	output, err := findTrustedTokenIssuerByARN(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSO Trusted Token Issuer (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	instanceARN := trustedTokenIssuerParseInstanceARN(ctx, r.Meta(), data.TrustedTokenIssuerARN.ValueString())
	data.InstanceARN = fwflex.StringToFrameworkARN(ctx, aws.String(instanceARN))

	// listTags requires both trusted token issuer and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	tags, err := listTags(ctx, conn, data.TrustedTokenIssuerARN.ValueString(), data.InstanceARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSO Trusted Token Issuer (%s) tags", data.ID.ValueString()), err.Error())

		return
	}

	setTagsOut(ctx, svcTags(tags))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *trustedTokenIssuerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old trustedTokenIssuerResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	if !new.Name.Equal(old.Name) ||
		!new.TrustedTokenIssuerConfiguration.Equal(old.TrustedTokenIssuerConfiguration) {
		var input ssoadmin.UpdateTrustedTokenIssuerInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateTrustedTokenIssuer(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SSO Trusted Token Issuer (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	// updateTags requires both trusted token issuer and instance ARN, so must be called
	// explicitly rather than with transparent tagging.
	if oldTagsAll, newTagsAll := old.TagsAll, new.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, new.ID.ValueString(), new.InstanceARN.ValueString(), oldTagsAll, newTagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating SSO Trusted Token Issuer (%s) tags", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *trustedTokenIssuerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data trustedTokenIssuerResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SSOAdminClient(ctx)

	input := ssoadmin.DeleteTrustedTokenIssuerInput{
		TrustedTokenIssuerArn: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeleteTrustedTokenIssuer(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting SSO Trusted Token Issuer (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findTrustedTokenIssuerByARN(ctx context.Context, conn *ssoadmin.Client, arn string) (*ssoadmin.DescribeTrustedTokenIssuerOutput, error) {
	input := ssoadmin.DescribeTrustedTokenIssuerInput{
		TrustedTokenIssuerArn: aws.String(arn),
	}
	output, err := conn.DescribeTrustedTokenIssuer(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// Instance ARN is not returned by DescribeTrustedTokenIssuer but is needed for schema consistency when importing and tagging.
// Instance ARN can be extracted from the Trusted Token Issuer ARN.
func trustedTokenIssuerParseInstanceARN(ctx context.Context, c *conns.AWSClient, id string) string {
	parts := strings.Split(id, "/")

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return c.GlobalARNNoAccount(ctx, "sso", "instance/"+parts[1])
	}

	return ""
}

type trustedTokenIssuerResourceModel struct {
	framework.WithRegionModel
	ClientToken                     types.String                                                          `tfsdk:"client_token"`
	ID                              types.String                                                          `tfsdk:"id"`
	InstanceARN                     fwtypes.ARN                                                           `tfsdk:"instance_arn"`
	Name                            types.String                                                          `tfsdk:"name"`
	TrustedTokenIssuerARN           types.String                                                          `tfsdk:"arn"`
	TrustedTokenIssuerConfiguration fwtypes.ListNestedObjectValueOf[trustedTokenIssuerConfigurationModel] `tfsdk:"trusted_token_issuer_configuration"`
	TrustedTokenIssuerType          fwtypes.StringEnum[awstypes.TrustedTokenIssuerType]                   `tfsdk:"trusted_token_issuer_type"`
	Tags                            tftags.Map                                                            `tfsdk:"tags"`
	TagsAll                         tftags.Map                                                            `tfsdk:"tags_all"`
}

type trustedTokenIssuerConfigurationModel struct {
	OIDCJWTConfiguration fwtypes.ListNestedObjectValueOf[oidcJWTConfigurationModel] `tfsdk:"oidc_jwt_configuration"`
}

var (
	_ fwflex.TypedExpander = trustedTokenIssuerConfigurationModel{}
	_ fwflex.Flattener     = &trustedTokenIssuerConfigurationModel{}
)

func (m trustedTokenIssuerConfigurationModel) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var result any
	var diags diag.Diagnostics

	switch targetType {
	case reflect.TypeFor[awstypes.TrustedTokenIssuerConfiguration]():
		r, d := m.expandToTrustedTokenIssuerConfiguration(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		result = r
	case reflect.TypeFor[awstypes.TrustedTokenIssuerUpdateConfiguration]():
		r, d := m.expandToTrustedTokenIssuerUpdateConfiguration(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		result = r
	}

	return result, diags
}

func (m trustedTokenIssuerConfigurationModel) expandToTrustedTokenIssuerConfiguration(ctx context.Context) (awstypes.TrustedTokenIssuerConfiguration, diag.Diagnostics) {
	var result awstypes.TrustedTokenIssuerConfiguration
	var diags diag.Diagnostics

	switch {
	case !m.OIDCJWTConfiguration.IsNull():
		var r awstypes.TrustedTokenIssuerConfigurationMemberOidcJwtConfiguration
		diags.Append(fwflex.Expand(ctx, m.OIDCJWTConfiguration, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m trustedTokenIssuerConfigurationModel) expandToTrustedTokenIssuerUpdateConfiguration(ctx context.Context) (awstypes.TrustedTokenIssuerUpdateConfiguration, diag.Diagnostics) {
	var result awstypes.TrustedTokenIssuerUpdateConfiguration
	var diags diag.Diagnostics

	switch {
	case !m.OIDCJWTConfiguration.IsNull():
		var r awstypes.TrustedTokenIssuerUpdateConfigurationMemberOidcJwtConfiguration
		diags.Append(fwflex.Expand(ctx, m.OIDCJWTConfiguration, &r.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		result = &r
	}

	return result, diags
}

func (m *trustedTokenIssuerConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch v := v.(type) {
	case awstypes.TrustedTokenIssuerConfigurationMemberOidcJwtConfiguration:
		var oidcJWTConfiguration oidcJWTConfigurationModel
		diags.Append(fwflex.Flatten(ctx, v.Value, &oidcJWTConfiguration)...)
		if diags.HasError() {
			return diags
		}

		m.OIDCJWTConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &oidcJWTConfiguration)
	}

	return diags
}

type oidcJWTConfigurationModel struct {
	ClaimAttributePath         types.String                                     `tfsdk:"claim_attribute_path"`
	IdentityStoreAttributePath types.String                                     `tfsdk:"identity_store_attribute_path"`
	IssuerURL                  types.String                                     `tfsdk:"issuer_url"`
	JWKSRetrievalOption        fwtypes.StringEnum[awstypes.JwksRetrievalOption] `tfsdk:"jwks_retrieval_option"`
}
