// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"
	"fmt"
	"log"
	"strings"

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Identity Source")
func newResourceIdentitySource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceIdentitySource{}

	return r, nil
}

const (
	ResNameIdentitySource = "Identity Source"
)

type resourceIdentitySource struct {
	framework.ResourceWithConfigure
}

func (r *resourceIdentitySource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_verifiedpermissions_identity_source"
}

func (r *resourceIdentitySource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
									"user_pool_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"client_ids": schema.ListAttribute{
										Computed:    true,
										CustomType:  fwtypes.ListOfStringType,
										ElementType: types.StringType,
										Optional:    true,
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
									names.AttrIssuer: schema.StringAttribute{
										Required: true,
									},
									"entity_id_prefix": schema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]schema.Block{
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

func (r *resourceIdentitySource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var plan resourceIdentitySourceData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &verifiedpermissions.CreateIdentitySourceInput{}
	response.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	clientToken := id.UniqueId()
	input.ClientToken = aws.String(clientToken)

	// Manually set as Union types are not supported by AutoFlex yet.
	if !plan.Configuration.IsNull() {
		var tfList []configuration
		response.Diagnostics.Append(plan.Configuration.ElementsAs(ctx, &tfList, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		configuration, d := expandConfiguration(ctx, tfList)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		input.Configuration = configuration
	}

	output, err := conn.CreateIdentitySource(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNameIdentitySource, clientToken, err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = flex.StringValueToFramework(ctx, aws.ToString(output.IdentitySourceId))

	response.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get call to retrieve computed values not included in create response.
	out, err := findIdentitySourceByIDAndPolicyStoreID(ctx, conn, state.ID.ValueString(), state.PolicyStoreID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNameIdentitySource, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.PrincipalEntityType = flex.StringToFramework(ctx, out.PrincipalEntityType)

	configuration, d := flattenConfiguration(ctx, out.Configuration)
	response.Diagnostics.Append(d...)
	state.Configuration = configuration

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceIdentitySource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state resourceIdentitySourceData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findIdentitySourceByIDAndPolicyStoreID(ctx, conn, state.ID.ValueString(), state.PolicyStoreID.ValueString())

	if tfresource.NotFound(err) {
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

	configuration, d := flattenConfiguration(ctx, output.Configuration)
	response.Diagnostics.Append(d...)
	state.Configuration = configuration

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceIdentitySource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state, plan resourceIdentitySourceData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Configuration.Equal(state.Configuration) || !plan.PolicyStoreID.Equal(state.PolicyStoreID) || !plan.PrincipalEntityType.Equal(state.PrincipalEntityType) {
		input := &verifiedpermissions.UpdateIdentitySourceInput{
			IdentitySourceId: flex.StringFromFramework(ctx, plan.ID),
		}

		response.Diagnostics.Append(flex.Expand(ctx, plan, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Manually set as Union types are not supported by AutoFlex yet.
		if !plan.Configuration.IsNull() {
			var tfList []configuration
			response.Diagnostics.Append(plan.Configuration.ElementsAs(ctx, &tfList, false)...)
			if response.Diagnostics.HasError() {
				return
			}

			configuration, d := expandUpdateConfiguration(ctx, tfList)
			response.Diagnostics.Append(d...)
			if response.Diagnostics.HasError() {
				return
			}
			input.UpdateConfiguration = configuration
		}

		output, err := conn.UpdateIdentitySource(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNameIdentitySource, state.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceIdentitySource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state resourceIdentitySourceData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Verified Permissions Identity Source", map[string]interface{}{
		names.AttrID: state.ID.ValueString(),
	})

	input := &verifiedpermissions.DeleteIdentitySourceInput{
		IdentitySourceId: flex.StringFromFramework(ctx, state.ID),
		PolicyStoreId:    flex.StringFromFramework(ctx, state.PolicyStoreID),
	}

	_, err := conn.DeleteIdentitySource(ctx, input)

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

func (r *resourceIdentitySource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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

func expandConfiguration(ctx context.Context, tfList []configuration) (awstypes.Configuration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	if !tfObj.CognitoUserPoolConfiguration.IsNull() {
		cognitoUserPoolConfigurationData, d := tfObj.CognitoUserPoolConfiguration.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.ConfigurationMemberCognitoUserPoolConfiguration{}
		diags.Append(flex.Expand(ctx, cognitoUserPoolConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return input, diags
	}

	if !tfObj.OpenIDConnectConfiguration.IsNull() {
		openIDConnectConfigurationData, d := tfObj.OpenIDConnectConfiguration.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.ConfigurationMemberOpenIdConnectConfiguration{}
		diags.Append(flex.Expand(ctx, openIDConnectConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		// Manually set TokenSelection as Union types are not supported by AutoFlex yet.
		var openIDConnectTokenSelectionData []openIDConnectTokenSelection
		diags.Append(openIDConnectConfigurationData.TokenSelection.ElementsAs(ctx, &openIDConnectTokenSelectionData, false)...)

		tokenSelection, d := expandTokenSelection(ctx, openIDConnectTokenSelectionData)
		diags.Append(d...)

		input.Value.TokenSelection = tokenSelection

		return input, diags
	}

	return nil, diags
}

func expandTokenSelection(ctx context.Context, tfList []openIDConnectTokenSelection) (awstypes.OpenIdConnectTokenSelection, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	if !tfObj.AccessTokenOnly.IsNull() {
		openIDConnectAccessTokenConfigurationData, d := tfObj.AccessTokenOnly.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.OpenIdConnectTokenSelectionMemberAccessTokenOnly{}
		diags.Append(flex.Expand(ctx, openIDConnectAccessTokenConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return input, diags
	}

	if !tfObj.IdentityTokenOnly.IsNull() {
		openIDConnectIdentityTokenConfigurationData, d := tfObj.IdentityTokenOnly.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.OpenIdConnectTokenSelectionMemberIdentityTokenOnly{}
		diags.Append(flex.Expand(ctx, openIDConnectIdentityTokenConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return input, diags
	}

	return nil, diags
}

func expandUpdateConfiguration(ctx context.Context, tfList []configuration) (awstypes.UpdateConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	if !tfObj.CognitoUserPoolConfiguration.IsNull() {
		cognitoUserPoolConfigurationData, d := tfObj.CognitoUserPoolConfiguration.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.UpdateConfigurationMemberCognitoUserPoolConfiguration{}
		diags.Append(flex.Expand(ctx, cognitoUserPoolConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		return input, diags
	}

	if !tfObj.OpenIDConnectConfiguration.IsNull() {
		openIDConnectConfigurationData, d := tfObj.OpenIDConnectConfiguration.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.UpdateConfigurationMemberOpenIdConnectConfiguration{}
		diags.Append(flex.Expand(ctx, openIDConnectConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}

		// Manually set TokenSelection as Union types are not supported by AutoFlex yet.
		var openIDConnectTokenSelectionData []openIDConnectTokenSelection
		diags.Append(openIDConnectConfigurationData.TokenSelection.ElementsAs(ctx, &openIDConnectTokenSelectionData, false)...)

		tokenSelection, d := expandUpdateTokenSelection(ctx, openIDConnectTokenSelectionData)
		diags.Append(d...)

		input.Value.TokenSelection = tokenSelection

		return input, diags
	}

	return nil, diags
}

func expandUpdateTokenSelection(ctx context.Context, tfList []openIDConnectTokenSelection) (awstypes.UpdateOpenIdConnectTokenSelection, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(tfList) == 0 {
		return nil, diags
	}
	tfObj := tfList[0]

	if !tfObj.AccessTokenOnly.IsNull() {
		openIDConnectAccessTokenConfigurationData, d := tfObj.AccessTokenOnly.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.UpdateOpenIdConnectTokenSelectionMemberAccessTokenOnly{}
		diags.Append(flex.Expand(ctx, openIDConnectAccessTokenConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return input, diags
	}

	if !tfObj.IdentityTokenOnly.IsNull() {
		openIDConnectIdentityTokenConfigurationData, d := tfObj.IdentityTokenOnly.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		input := &awstypes.UpdateOpenIdConnectTokenSelectionMemberIdentityTokenOnly{}
		diags.Append(flex.Expand(ctx, openIDConnectIdentityTokenConfigurationData, &input.Value)...)
		if diags.HasError() {
			return nil, diags
		}
		return input, diags
	}

	return nil, diags
}

func flattenConfiguration(ctx context.Context, apiObject awstypes.ConfigurationDetail) (fwtypes.ListNestedObjectValueOf[configuration], diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[configuration](ctx), diags
	}

	obj := &configuration{}

	switch v := apiObject.(type) {
	case *awstypes.ConfigurationDetailMemberCognitoUserPoolConfiguration:
		var cognitoUserPoolConfigurationData cognitoUserPoolConfiguration
		d := flex.Flatten(ctx, v.Value, &cognitoUserPoolConfigurationData)
		diags.Append(d...)

		obj.CognitoUserPoolConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &cognitoUserPoolConfigurationData)
		obj.OpenIDConnectConfiguration = fwtypes.NewListNestedObjectValueOfNull[openIDConnectConfiguration](ctx)
	case *awstypes.ConfigurationDetailMemberOpenIdConnectConfiguration:
		var openIDConnectConfigurationData openIDConnectConfiguration
		d := flex.Flatten(ctx, v.Value, &openIDConnectConfigurationData)
		diags.Append(d...)

		// Manually set as Union types are not supported by AutoFlex yet.
		tokenSelectionData, d := flattenTokenSelection(ctx, v.Value.TokenSelection)
		diags.Append(d...)
		openIDConnectConfigurationData.TokenSelection = tokenSelectionData

		obj.CognitoUserPoolConfiguration = fwtypes.NewListNestedObjectValueOfNull[cognitoUserPoolConfiguration](ctx)
		obj.OpenIDConnectConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &openIDConnectConfigurationData)
	default:
		log.Println("union is nil or unknown type")
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obj), diags
}

func flattenTokenSelection(ctx context.Context, apiObject awstypes.OpenIdConnectTokenSelectionDetail) (fwtypes.ListNestedObjectValueOf[openIDConnectTokenSelection], diag.Diagnostics) {
	var diags diag.Diagnostics

	if apiObject == nil {
		return fwtypes.NewListNestedObjectValueOfNull[openIDConnectTokenSelection](ctx), diags
	}

	obj := &openIDConnectTokenSelection{}

	switch v := apiObject.(type) {
	case *awstypes.OpenIdConnectTokenSelectionDetailMemberAccessTokenOnly:
		var openIDConnectAccessTokenConfigurationData openIDConnectAccessTokenConfiguration
		d := flex.Flatten(ctx, v.Value, &openIDConnectAccessTokenConfigurationData)
		diags.Append(d...)

		obj.AccessTokenOnly = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &openIDConnectAccessTokenConfigurationData)
		obj.IdentityTokenOnly = fwtypes.NewListNestedObjectValueOfNull[openIDConnectIdentityTokenConfiguration](ctx)
	case *awstypes.OpenIdConnectTokenSelectionDetailMemberIdentityTokenOnly:
		var openIDConnectIdentityTokenConfigurationData openIDConnectIdentityTokenConfiguration
		d := flex.Flatten(ctx, v.Value, &openIDConnectIdentityTokenConfigurationData)
		diags.Append(d...)

		obj.AccessTokenOnly = fwtypes.NewListNestedObjectValueOfNull[openIDConnectAccessTokenConfiguration](ctx)
		obj.IdentityTokenOnly = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &openIDConnectIdentityTokenConfigurationData)
	default:
		log.Println("union is nil or unknown type")
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, obj), diags
}

type resourceIdentitySourceData struct {
	Configuration       fwtypes.ListNestedObjectValueOf[configuration] `tfsdk:"configuration"`
	ID                  types.String                                   `tfsdk:"id"`
	PolicyStoreID       types.String                                   `tfsdk:"policy_store_id"`
	PrincipalEntityType types.String                                   `tfsdk:"principal_entity_type"`
}

type configuration struct {
	CognitoUserPoolConfiguration fwtypes.ListNestedObjectValueOf[cognitoUserPoolConfiguration] `tfsdk:"cognito_user_pool_configuration"`
	OpenIDConnectConfiguration   fwtypes.ListNestedObjectValueOf[openIDConnectConfiguration]   `tfsdk:"open_id_connect_configuration"`
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
	in := &verifiedpermissions.GetIdentitySourceInput{
		IdentitySourceId: aws.String(id),
		PolicyStoreId:    aws.String(policyStoreID),
	}

	out, err := conn.GetIdentitySource(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.IdentitySourceId == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
