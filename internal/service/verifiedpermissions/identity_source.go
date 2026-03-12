// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package verifiedpermissions

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_verifiedpermissions_identity_source", name="Identity Source")
func newIdentitySourceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &identitySourceResource{}

	return r, nil
}

const (
	ResNameIdentitySource = "Identity Source"
)

type identitySourceResource struct {
	framework.ResourceWithModel[identitySourceResourceModel]
}

func (r *identitySourceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"policy_store_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal_entity_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[configuration](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"cognito_user_pool_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cognitoUserPoolConfiguration](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"client_ids": schema.ListAttribute{
										Computed:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
									"user_pool_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"group_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[cognitoGroupConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"group_entity_type": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"open_id_connect_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[openIDConnectConfiguration](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"entity_id_prefix": schema.StringAttribute{
										Optional: true,
									},
									names.AttrIssuer: schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"group_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[openIDConnectGroupConfiguration](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"group_claim": schema.StringAttribute{
													Required: true,
												},
												"group_entity_type": schema.StringAttribute{
													Required: true,
												},
											},
										},
									},
									"token_selection": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[openIDConnectTokenSelection](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"access_token_only": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[openIDConnectAccessTokenConfiguration](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"audiences": schema.ListAttribute{
																CustomType:  fwtypes.ListOfStringType,
																ElementType: types.StringType,
																Optional:    true,
															},
															"principal_id_claim": schema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"identity_token_only": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[openIDConnectIdentityTokenConfiguration](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"client_ids": schema.ListAttribute{
																CustomType:  fwtypes.ListOfStringType,
																ElementType: types.StringType,
																Optional:    true,
															},
															"principal_id_claim": schema.StringAttribute{
																Optional: true,
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
				},
			},
		},
	}

	response.Schema = s
}

func (r *identitySourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var plan identitySourceResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var input verifiedpermissions.CreateIdentitySourceInput
	response.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientToken := sdkid.UniqueId()
	input.ClientToken = aws.String(clientToken)

	output, err := tfresource.RetryWhenIsA[*verifiedpermissions.CreateIdentitySourceOutput, *awstypes.ResourceNotFoundException](ctx, 1*time.Minute, func(ctx context.Context) (*verifiedpermissions.CreateIdentitySourceOutput, error) {
		return conn.CreateIdentitySource(ctx, &input)
	})
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNameIdentitySource, clientToken, err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = flex.StringValueToFramework(ctx, aws.ToString(output.IdentitySourceId))

	// Get call to retrieve computed values not included in create response.
	out, err := findIdentitySourceByIDAndPolicyStoreID(ctx, conn, state.ID.ValueString(), state.PolicyStoreID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNameIdentitySource, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *identitySourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state identitySourceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findIdentitySourceByIDAndPolicyStoreID(ctx, conn, state.ID.ValueString(), state.PolicyStoreID.ValueString())

	if retry.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, ResNameIdentitySource, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *identitySourceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state, plan identitySourceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Configuration.Equal(state.Configuration) || !plan.PolicyStoreID.Equal(state.PolicyStoreID) || !plan.PrincipalEntityType.Equal(state.PrincipalEntityType) {
		input := verifiedpermissions.UpdateIdentitySourceInput{
			IdentitySourceId: flex.StringFromFramework(ctx, plan.ID),
		}

		response.Diagnostics.Append(flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Update"))...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateIdentitySource(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNameIdentitySource, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		// Retrieve values not included in update response.
		out, err := findIdentitySourceByIDAndPolicyStoreID(ctx, conn, state.ID.ValueString(), state.PolicyStoreID.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNameIdentitySource, plan.ID.String(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *identitySourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state identitySourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Verified Permissions Identity Source", map[string]any{
		names.AttrID: state.ID.ValueString(),
	})

	input := verifiedpermissions.DeleteIdentitySourceInput{
		IdentitySourceId: flex.StringFromFramework(ctx, state.ID),
		PolicyStoreId:    flex.StringFromFramework(ctx, state.PolicyStoreID),
	}

	_, err := conn.DeleteIdentitySource(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionDeleting, ResNameIdentitySource, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *identitySourceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ":")
	if len(parts) != 2 {
		response.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf("unexpected format of import ID (%s), expected: 'POLICY_STORE_ID:IDENTITY-SOURCE-ID'", request.ID))
		return
	}
	policyStoreID := parts[0]
	identitySourceID := parts[1]
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), identitySourceID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("policy_store_id"), policyStoreID)...)
}

type identitySourceResourceModel struct {
	framework.WithRegionModel
	Configuration       fwtypes.ListNestedObjectValueOf[configuration] `tfsdk:"configuration"`
	ID                  types.String                                   `tfsdk:"id"`
	PolicyStoreID       types.String                                   `tfsdk:"policy_store_id"`
	PrincipalEntityType types.String                                   `tfsdk:"principal_entity_type"`
}

type configuration struct {
	CognitoUserPoolConfiguration fwtypes.ListNestedObjectValueOf[cognitoUserPoolConfiguration] `tfsdk:"cognito_user_pool_configuration"`
	OpenIDConnectConfiguration   fwtypes.ListNestedObjectValueOf[openIDConnectConfiguration]   `tfsdk:"open_id_connect_configuration"`
}

var (
	_ flex.TypedExpander = configuration{}
	_ flex.Flattener     = &configuration{}
)

func (m configuration) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awstypes.Configuration]():
		return m.expandToConfiguration(ctx)

	case reflect.TypeFor[awstypes.UpdateConfiguration]():
		return m.expandToUpdateConfiguration(ctx)
	}
	return nil, diags
}

func (m configuration) expandToConfiguration(ctx context.Context) (result awstypes.Configuration, diags diag.Diagnostics) {
	switch {
	case !m.CognitoUserPoolConfiguration.IsNull():
		var result awstypes.ConfigurationMemberCognitoUserPoolConfiguration
		diags.Append(flex.Expand(ctx, m.CognitoUserPoolConfiguration, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags

	case !m.OpenIDConnectConfiguration.IsNull():
		var result awstypes.ConfigurationMemberOpenIdConnectConfiguration
		diags.Append(flex.Expand(ctx, m.OpenIDConnectConfiguration, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	}

	return nil, diags
}

func (m configuration) expandToUpdateConfiguration(ctx context.Context) (result awstypes.UpdateConfiguration, diags diag.Diagnostics) {
	switch {
	case !m.CognitoUserPoolConfiguration.IsNull():
		var result awstypes.UpdateConfigurationMemberCognitoUserPoolConfiguration
		diags.Append(flex.Expand(ctx, m.CognitoUserPoolConfiguration, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags

	case !m.OpenIDConnectConfiguration.IsNull():
		var result awstypes.UpdateConfigurationMemberOpenIdConnectConfiguration
		diags.Append(flex.Expand(ctx, m.OpenIDConnectConfiguration, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	}

	return nil, diags
}

func (m *configuration) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.ConfigurationDetailMemberCognitoUserPoolConfiguration:
		var model cognitoUserPoolConfiguration
		di := flex.Flatten(ctx, t.Value, &model)
		diags.Append(di...)
		if diags.HasError() {
			return diags
		}

		m.CognitoUserPoolConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	case awstypes.ConfigurationDetailMemberOpenIdConnectConfiguration:
		var model openIDConnectConfiguration
		di := flex.Flatten(ctx, t.Value, &model)
		diags.Append(di...)
		if diags.HasError() {
			return diags
		}

		m.OpenIDConnectConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	default:
		return diags
	}
}

type cognitoUserPoolConfiguration struct {
	UserPoolARN        fwtypes.ARN                                                `tfsdk:"user_pool_arn"`
	ClientIds          fwtypes.ListValueOf[types.String]                          `tfsdk:"client_ids"`
	GroupConfiguration fwtypes.ListNestedObjectValueOf[cognitoGroupConfiguration] `tfsdk:"group_configuration"`
}

type cognitoGroupConfiguration struct {
	GroupEntityType types.String `tfsdk:"group_entity_type"`
}

type openIDConnectConfiguration struct {
	Issuer             types.String                                                     `tfsdk:"issuer"`
	TokenSelection     fwtypes.ListNestedObjectValueOf[openIDConnectTokenSelection]     `tfsdk:"token_selection"`
	EntityIDPrefix     types.String                                                     `tfsdk:"entity_id_prefix"`
	GroupConfiguration fwtypes.ListNestedObjectValueOf[openIDConnectGroupConfiguration] `tfsdk:"group_configuration"`
}

type openIDConnectTokenSelection struct {
	AccessTokenOnly   fwtypes.ListNestedObjectValueOf[openIDConnectAccessTokenConfiguration]   `tfsdk:"access_token_only"`
	IdentityTokenOnly fwtypes.ListNestedObjectValueOf[openIDConnectIdentityTokenConfiguration] `tfsdk:"identity_token_only"`
}

var (
	_ flex.TypedExpander = openIDConnectTokenSelection{}
	_ flex.Flattener     = &openIDConnectTokenSelection{}
)

func (m openIDConnectTokenSelection) ExpandTo(ctx context.Context, targetType reflect.Type) (result any, diags diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[awstypes.OpenIdConnectTokenSelection]():
		return m.expandToOpenIDConnectTokenSelection(ctx)

	case reflect.TypeFor[awstypes.UpdateOpenIdConnectTokenSelection]():
		return m.expandToUpdateOpenIDConnectTokenSelection(ctx)
	}
	return nil, diags
}

func (m openIDConnectTokenSelection) expandToOpenIDConnectTokenSelection(ctx context.Context) (result awstypes.OpenIdConnectTokenSelection, diags diag.Diagnostics) {
	switch {
	case !m.AccessTokenOnly.IsNull():
		var result awstypes.OpenIdConnectTokenSelectionMemberAccessTokenOnly
		diags.Append(flex.Expand(ctx, m.AccessTokenOnly, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags

	case !m.IdentityTokenOnly.IsNull():
		var result awstypes.OpenIdConnectTokenSelectionMemberIdentityTokenOnly
		diags.Append(flex.Expand(ctx, m.IdentityTokenOnly, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	}

	return nil, diags
}

func (m openIDConnectTokenSelection) expandToUpdateOpenIDConnectTokenSelection(ctx context.Context) (result awstypes.UpdateOpenIdConnectTokenSelection, diags diag.Diagnostics) {
	switch {
	case !m.AccessTokenOnly.IsNull():
		var result awstypes.UpdateOpenIdConnectTokenSelectionMemberAccessTokenOnly
		diags.Append(flex.Expand(ctx, m.AccessTokenOnly, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags

	case !m.IdentityTokenOnly.IsNull():
		var result awstypes.UpdateOpenIdConnectTokenSelectionMemberIdentityTokenOnly
		diags.Append(flex.Expand(ctx, m.IdentityTokenOnly, &result.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return &result, diags
	}

	return nil, diags
}

func (m *openIDConnectTokenSelection) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch t := v.(type) {
	case awstypes.OpenIdConnectTokenSelectionDetailMemberAccessTokenOnly:
		var model openIDConnectAccessTokenConfiguration
		di := flex.Flatten(ctx, t.Value, &model)
		diags.Append(di...)
		if diags.HasError() {
			return diags
		}

		m.AccessTokenOnly = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	case awstypes.OpenIdConnectTokenSelectionDetailMemberIdentityTokenOnly:
		var model openIDConnectIdentityTokenConfiguration
		di := flex.Flatten(ctx, t.Value, &model)
		diags.Append(di...)
		if diags.HasError() {
			return diags
		}

		m.IdentityTokenOnly = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

		return diags

	default:
		return diags
	}
}

type openIDConnectAccessTokenConfiguration struct {
	Audiences        fwtypes.ListValueOf[types.String] `tfsdk:"audiences"`
	PrincipalIdClaim types.String                      `tfsdk:"principal_id_claim"`
}

type openIDConnectIdentityTokenConfiguration struct {
	ClientIds        fwtypes.ListValueOf[types.String] `tfsdk:"client_ids"`
	PrincipalIdClaim types.String                      `tfsdk:"principal_id_claim"`
}

type openIDConnectGroupConfiguration struct {
	GroupClaim      types.String `tfsdk:"group_claim"`
	GroupEntityType types.String `tfsdk:"group_entity_type"`
}

func findIdentitySourceByIDAndPolicyStoreID(ctx context.Context, conn *verifiedpermissions.Client, id string, policyStoreID string) (*verifiedpermissions.GetIdentitySourceOutput, error) {
	in := verifiedpermissions.GetIdentitySourceInput{
		IdentitySourceId: aws.String(id),
		PolicyStoreId:    aws.String(policyStoreID),
	}

	out, err := conn.GetIdentitySource(ctx, &in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.IdentitySourceId == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}
