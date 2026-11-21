// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	lazystringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator/lazy"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Cognito can reject concurrent writes for the terms and privacy policy on one client.
const userPoolClientTermsConflictTimeout = 2 * time.Minute

// @FrameworkResource("aws_cognito_user_pool_client_terms", name="User Pool Client Terms")
// @IdentityAttribute("user_pool_id")
// @IdentityAttribute("terms_id")
// @ImportIDHandler("userPoolClientTermsImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types;awstypes;awstypes.TermsType")
// @Testing(importStateIdAttribute="terms_id")
// @Testing(importStateIdAttributes="user_pool_id;terms_id", importStateIdAttributesSep="flex.ResourceIdSeparator")
func newUserPoolClientTermsResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &userPoolClientTermsResource{}, nil
}

type userPoolClientTermsResource struct {
	framework.ResourceWithModel[userPoolClientTermsResourceModel]
	framework.WithImportByIdentity
}

func (r *userPoolClientTermsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrClientID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"links": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     mapdefault.StaticValue(types.MapValueMust(types.StringType, nil)),
				Validators: []validator.Map{
					mapvalidator.SizeAtMost(13),
					mapvalidator.KeysAre(lazystringvalidator.RegexMatches(
						`^cognito:(default|dutch|english|french|spanish|german|bahasa-indonesia|italian|japanese|korean|portuguese-brazil|chinese-(simplified|traditional))$`,
						"must be a supported Cognito language key",
					)),
					mapvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 1024)),
				},
			},
			"terms_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"terms_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("privacy-policy", "terms-of-use"),
				},
			},
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *userPoolClientTermsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan userPoolClientTermsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cognitoidentityprovider.CreateTermsInput{
		Enforcement: awstypes.TermsEnforcementTypeNone,
		TermsSource: awstypes.TermsSourceTypeLink,
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := tfresource.RetryWhenIsA[*cognitoidentityprovider.CreateTermsOutput, *awstypes.ConcurrentModificationException](ctx, userPoolClientTermsConflictTimeout, func(ctx context.Context) (*cognitoidentityprovider.CreateTermsOutput, error) {
		return conn.CreateTerms(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.TermsName.ValueString())
		return
	}
	if out == nil || out.Terms == nil || aws.ToString(out.Terms.TermsId) == "" {
		smerr.AddError(ctx, &resp.Diagnostics, tfresource.NewEmptyResultError(), smerr.ID, plan.TermsName.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, plan.flatten(ctx, out.Terms))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *userPoolClientTermsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state userPoolClientTermsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserPoolClientTermsByID(ctx, conn, state.UserPoolID.ValueString(), state.TermsID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.TermsID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, state.flatten(ctx, out))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userPoolClientTermsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan, state userPoolClientTermsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, diags := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, diags)
	if resp.Diagnostics.HasError() {
		return
	}
	if !diff.HasChanges() {
		smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
		return
	}

	input := cognitoidentityprovider.UpdateTermsInput{
		TermsId: state.TermsID.ValueStringPointer(),
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}
	// Include an empty map explicitly when the configuration removes all links.
	if input.Links == nil {
		input.Links = map[string]string{}
	}

	out, err := tfresource.RetryWhenIsA[*cognitoidentityprovider.UpdateTermsOutput, *awstypes.ConcurrentModificationException](ctx, userPoolClientTermsConflictTimeout, func(ctx context.Context) (*cognitoidentityprovider.UpdateTermsOutput, error) {
		return conn.UpdateTerms(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.TermsID.ValueString())
		return
	}
	if out == nil || out.Terms == nil || aws.ToString(out.Terms.TermsId) == "" {
		smerr.AddError(ctx, &resp.Diagnostics, tfresource.NewEmptyResultError(), smerr.ID, state.TermsID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, plan.flatten(ctx, out.Terms))
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *userPoolClientTermsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state userPoolClientTermsResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cognitoidentityprovider.DeleteTermsInput{
		TermsId:    state.TermsID.ValueStringPointer(),
		UserPoolId: state.UserPoolID.ValueStringPointer(),
	}

	_, err := tfresource.RetryWhenIsA[*cognitoidentityprovider.DeleteTermsOutput, *awstypes.ConcurrentModificationException](ctx, userPoolClientTermsConflictTimeout, func(ctx context.Context) (*cognitoidentityprovider.DeleteTermsOutput, error) {
		return conn.DeleteTerms(ctx, &input)
	})
	if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.TermsID.ValueString())
	}
}

func findUserPoolClientTermsByID(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, termsID string) (*awstypes.TermsType, error) {
	input := cognitoidentityprovider.DescribeTermsInput{
		TermsId:    aws.String(termsID),
		UserPoolId: aws.String(userPoolID),
	}

	out, err := conn.DescribeTerms(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if out == nil || out.Terms == nil || aws.ToString(out.Terms.TermsId) == "" {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out.Terms, nil
}

type userPoolClientTermsResourceModel struct {
	framework.WithRegionModel
	ClientID   types.String `tfsdk:"client_id"`
	Links      types.Map    `tfsdk:"links"`
	TermsID    types.String `tfsdk:"terms_id"`
	TermsName  types.String `tfsdk:"terms_name"`
	UserPoolID types.String `tfsdk:"user_pool_id"`
}

func (m *userPoolClientTermsResourceModel) flatten(ctx context.Context, terms *awstypes.TermsType) diag.Diagnostics {
	diags := flex.Flatten(ctx, terms, m)
	// Cognito can omit Links when empty. Match the Terraform default in state.
	if len(terms.Links) == 0 {
		m.Links = types.MapValueMust(types.StringType, nil)
	}
	return diags
}

var _ inttypes.ImportIDParser = userPoolClientTermsImportID{}

type userPoolClientTermsImportID struct{}

func (userPoolClientTermsImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := intflex.ExpandResourceId(id, 2, false)
	if err != nil {
		return "", nil, smarterr.Errorf("unexpected import ID %q: expected format <user_pool_id>,<terms_id>: %w", id, err)
	}

	return id, map[string]any{
		names.AttrUserPoolID: parts[0],
		"terms_id":           parts[1],
	}, nil
}
