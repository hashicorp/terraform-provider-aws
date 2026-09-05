// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfobjectvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/objectvalidator"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_agentregistry_registry", name="Registry")
// @Tags(identifierAttribute="registry_arn")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="registry_id")
// @IdentityAttribute("registry_id")
func newRegistryResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &registryResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type registryResource struct {
	framework.ResourceWithModel[registryResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *registryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4096),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\-\.\/]{0,63}$`),
						"Must start with a letter or digit. Valid characters are a-z, A-Z, 0-9, _ (underscore), - (hyphen), . (dot), and / (forward slash). The name can have up to 64 characters.",
					),
				},
			},
			"registry_arn":    framework.ARNAttributeComputedOnly(),
			"registry_id":     framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"approval_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[approvalConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"auto_approval_rules": schema.SetAttribute{
							CustomType: fwtypes.SetOfStringEnumType[awstypes.AutoApprovalRule](),
							Optional:   true,
						},
					},
				},
			},
			"discovery_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[discoveryConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"authorizer_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.RegistryAuthorizerType](),
							Required:   true,
							Validators: []validator.String{
								tfstringvalidator.AlsoRequiresWhenEquals(
									awstypes.RegistryAuthorizerTypeCustomJwt,
									path.MatchRelative().AtParent().AtName("authorizer_configuration").AtListIndex(0).AtName("custom_jwt_authorizer"),
								),
								tfstringvalidator.ConflictsWithWhenEquals(
									awstypes.RegistryAuthorizerTypeAwsIam,
									path.MatchRelative().AtParent().AtName("authorizer_configuration"),
								),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"authorizer_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[authorizerConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Validators: []validator.Object{
									tfobjectvalidator.AtLeastOneOfChildren(
										path.MatchRelative().AtName("custom_jwt_authorizer"),
									),
								},
								Blocks: map[string]schema.Block{
									"custom_jwt_authorizer": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[customJWTAuthorizerConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"allowed_audience": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Optional:   true,
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
												},
												"allowed_clients": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Optional:   true,
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
												},
												"allowed_scopes": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Optional:   true,
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
												},
												"discovery_url": schema.StringAttribute{
													Required: true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"custom_claim": schema.SetNestedBlock{
													CustomType: fwtypes.NewSetNestedObjectTypeOf[customClaimValidationTypeModel](ctx),
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"inbound_token_claim_name": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 255),
																	stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.:-]+$`), "must contain only letters, numbers, and the characters _ . - :"),
																},
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
															"inbound_token_claim_value_type": schema.StringAttribute{
																CustomType: fwtypes.StringEnumType[awstypes.InboundTokenClaimValueType](),
																Required:   true,
																PlanModifiers: []planmodifier.String{
																	stringplanmodifier.RequiresReplace(),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"authorizing_claim_match_value": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[authorizingClaimMatchValueTypeModel](ctx),
																Validators: []validator.List{
																	listvalidator.IsRequired(),
																	listvalidator.SizeAtLeast(1),
																	listvalidator.SizeAtMost(1),
																},
																PlanModifiers: []planmodifier.List{
																	listplanmodifier.RequiresReplace(),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"claim_match_operator": schema.StringAttribute{
																			CustomType: fwtypes.StringEnumType[awstypes.ClaimMatchOperatorType](),
																			Required:   true,
																			PlanModifiers: []planmodifier.String{
																				stringplanmodifier.RequiresReplace(),
																			},
																		},
																	},
																	Blocks: map[string]schema.Block{
																		"claim_match_value": schema.ListNestedBlock{
																			CustomType: fwtypes.NewListNestedObjectTypeOf[claimMatchValueTypeModel](ctx),
																			Validators: []validator.List{
																				listvalidator.IsRequired(),
																				listvalidator.SizeAtLeast(1),
																				listvalidator.SizeAtMost(1),
																			},
																			PlanModifiers: []planmodifier.List{
																				listplanmodifier.RequiresReplace(),
																			},
																			NestedObject: schema.NestedBlockObject{
																				Validators: []validator.Object{
																					tfobjectvalidator.AtLeastOneOfChildren(
																						path.MatchRelative().AtName("match_value_string"),
																						path.MatchRelative().AtName("match_value_string_list"),
																					),
																				},
																				Attributes: map[string]schema.Attribute{
																					"match_value_string": schema.StringAttribute{
																						Optional: true,
																						Validators: []validator.String{
																							stringvalidator.LengthBetween(1, 255),
																							stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.:-]+$`), "must contain only letters, numbers, and the characters _ . - :"),
																						},
																						PlanModifiers: []planmodifier.String{
																							stringplanmodifier.RequiresReplace(),
																						},
																					},
																					"match_value_string_list": schema.SetAttribute{
																						Optional:    true,
																						ElementType: types.StringType,
																						Validators: []validator.Set{
																							setvalidator.ValueStringsAre(
																								stringvalidator.LengthBetween(1, 255),
																								stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_.:-]+$`), "must contain only letters, numbers, and the characters _ . - :"),
																							),
																						},
																						PlanModifiers: []planmodifier.Set{
																							setplanmodifier.RequiresReplace(),
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
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *registryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input agentregistrycontrol.CreateRegistryInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateRegistry(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	registryARN := aws.ToString(out.RegistryArn)
	registryID, err := registryIDFromARN(registryARN)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
	}

	if _, err := waitRegistryCreated(ctx, conn, registryID, r.CreateTimeout(ctx, plan.Timeouts)); err != nil {
		// Taint the resource.
		resp.State.SetAttribute(ctx, path.Root("registry_id"), registryID)
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
		return
	}

	// Set values for unknowns.
	plan.RegistryARN = fwflex.StringValueToFramework(ctx, registryARN)
	plan.RegistryID = fwflex.StringValueToFramework(ctx, registryID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *registryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, state.RegistryID)
	out, err := findRegistryByID(ctx, conn, registryID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *registryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state registryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		registryID := fwflex.StringValueFromFramework(ctx, plan.RegistryID)
		optFns := []fwflex.AutoFlexOptionsFunc{
			fwflex.WithIgnoredFieldNamesAppend("ApprovalConfiguration"),
			fwflex.WithIgnoredFieldNamesAppend("Description"),
		}
		var input agentregistrycontrol.UpdateRegistryInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input, optFns...))
		if resp.Diagnostics.HasError() {
			return
		}

		if !plan.ApprovalConfiguration.Equal(state.ApprovalConfiguration) {
			input.ApprovalConfiguration = &awstypes.UpdatedApprovalConfiguration{}
			if !plan.ApprovalConfiguration.IsNull() {
				smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan.ApprovalConfiguration, &input.ApprovalConfiguration.OptionalValue))
				if resp.Diagnostics.HasError() {
					return
				}
			}
		}

		if !plan.Description.Equal(state.Description) {
			input.Description = &awstypes.UpdatedDescription{}
			if !plan.Description.IsNull() {
				input.Description.OptionalValue = fwflex.StringFromFramework(ctx, plan.Description)
			}
		}

		_, err := conn.UpdateRegistry(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
			return
		}

		if _, err := waitRegistryUpdated(ctx, conn, registryID, r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AgentRegistryClient(ctx)

	registryID := fwflex.StringValueFromFramework(ctx, state.RegistryID)
	input := agentregistrycontrol.DeleteRegistryInput{
		RegistryId: aws.String(registryID),
	}
	_, err := conn.DeleteRegistry(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
		return
	}

	if _, err := waitRegistryDeleted(ctx, conn, registryID, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, registryID)
		return
	}
}

func (r *registryResource) flatten(ctx context.Context, out *agentregistrycontrol.GetRegistryOutput, data *registryResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	// Normalize approval_configuration "approvalConfiguration": {"autoApprovalRules": []} to null.
	if out.ApprovalConfiguration != nil && len(out.ApprovalConfiguration.AutoApprovalRules) == 0 {
		out.ApprovalConfiguration = nil
	}
	diags.Append(fwflex.Flatten(ctx, out, data)...)
	return diags
}

func findRegistryByID(ctx context.Context, conn *agentregistrycontrol.Client, id string) (*agentregistrycontrol.GetRegistryOutput, error) {
	input := agentregistrycontrol.GetRegistryInput{
		RegistryId: aws.String(id),
	}
	return findRegistry(ctx, conn, &input)
}

func findRegistry(ctx context.Context, conn *agentregistrycontrol.Client, input *agentregistrycontrol.GetRegistryInput) (*agentregistrycontrol.GetRegistryOutput, error) {
	out, err := conn.GetRegistry(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func waitRegistryCreated(ctx context.Context, conn *agentregistrycontrol.Client, id string, timeout time.Duration) (*agentregistrycontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryStatusCreating),
		Target:                    enum.Slice(awstypes.RegistryStatusReady),
		Refresh:                   statusRegistry(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryOutput); ok {
		if out.Status == awstypes.RegistryStatusCreateFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryUpdated(ctx context.Context, conn *agentregistrycontrol.Client, id string, timeout time.Duration) (*agentregistrycontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.RegistryStatusUpdating),
		Target:                    enum.Slice(awstypes.RegistryStatusReady),
		Refresh:                   statusRegistry(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryOutput); ok {
		if out.Status == awstypes.RegistryStatusUpdateFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitRegistryDeleted(ctx context.Context, conn *agentregistrycontrol.Client, id string, timeout time.Duration) (*agentregistrycontrol.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistryStatusDeleting, awstypes.RegistryStatusReady),
		Target:  []string{},
		Refresh: statusRegistry(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*agentregistrycontrol.GetRegistryOutput); ok {
		if out.Status == awstypes.RegistryStatusDeleteFailed {
			retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		}
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusRegistry(conn *agentregistrycontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findRegistryByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func registryIDFromARN(registryyARN string) (string, error) {
	parsedARN, err := arn.Parse(registryyARN)
	if err != nil {
		return "", fmt.Errorf("parsing registry ARN (%s): %w", registryyARN, err)
	}
	memoryID := strings.TrimPrefix(parsedARN.Resource, "registry/")
	return memoryID, nil
}

type registryResourceModel struct {
	framework.WithRegionModel
	ApprovalConfiguration fwtypes.ListNestedObjectValueOf[approvalConfigurationModel] `tfsdk:"approval_configuration"`
	// AutoDetectionConfiguration fwtypes.ListNestedObjectValueOf[autoDetectionConfigurationModel] `tfsdk:"auto_detection_configuration"`
	Description            types.String                                                 `tfsdk:"description"`
	DiscoveryConfiguration fwtypes.ListNestedObjectValueOf[discoveryConfigurationModel] `tfsdk:"discovery_configuration"`
	// EncryptionConfiguration fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	Name        types.String   `tfsdk:"name"`
	RegistryARN types.String   `tfsdk:"registry_arn"`
	RegistryID  types.String   `tfsdk:"registry_id"`
	Tags        tftags.Map     `tfsdk:"tags"`
	TagsAll     tftags.Map     `tfsdk:"tags_all"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

type approvalConfigurationModel struct {
	AutoApprovalRules fwtypes.SetOfStringEnum[awstypes.AutoApprovalRule] `tfsdk:"auto_approval_rules"`
}

type discoveryConfigurationModel struct {
	AuthorizerConfiguration fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel] `tfsdk:"authorizer_configuration"`
	AuthorizerType          fwtypes.StringEnum[awstypes.RegistryAuthorizerType]           `tfsdk:"authorizer_type"`
}

type authorizerConfigurationModel struct {
	CustomJWTAuthorizer fwtypes.ListNestedObjectValueOf[customJWTAuthorizerConfigurationModel] `tfsdk:"custom_jwt_authorizer"`
}

var (
	_ fwflex.Expander  = authorizerConfigurationModel{}
	_ fwflex.Flattener = &authorizerConfigurationModel{}
)

func (m *authorizerConfigurationModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer:
		var model customJWTAuthorizerConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &model))
		if diags.HasError() {
			return diags
		}
		var d diag.Diagnostics
		m.CustomJWTAuthorizer, d = fwtypes.NewListNestedObjectValueOfPtr(ctx, &model)
		smerr.AddEnrich(ctx, &diags, d)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("authorizerConfigurationModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m authorizerConfigurationModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.CustomJWTAuthorizer.IsNull():
		model, d := m.CustomJWTAuthorizer.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.AuthorizerConfigurationMemberCustomJWTAuthorizer
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, model, &r.Value))
		return &r, diags
	}

	return nil, diags
}

type customJWTAuthorizerConfigurationModel struct {
	AllowedAudience fwtypes.ListOfString                                           `tfsdk:"allowed_audience"`
	AllowedClients  fwtypes.ListOfString                                           `tfsdk:"allowed_clients"`
	AllowedScopes   fwtypes.ListOfString                                           `tfsdk:"allowed_scopes"`
	CustomClaims    fwtypes.SetNestedObjectValueOf[customClaimValidationTypeModel] `tfsdk:"custom_claim"`
	DiscoveryURL    types.String                                                   `tfsdk:"discovery_url"`
	// TODO
	// PrivateEndpoint          fwtypes.SetNestedObjectValueOf[privateEndpointModel]         `tfsdk:"private_endpoint"`
	// PrivateEndpointOverrides fwtypes.SetNestedObjectValueOf[privateEndpointOverrideModel] `tfsdk:"private_endpoint_override"`
}

type customClaimValidationTypeModel struct {
	AuthorizingClaimMatchValue fwtypes.ListNestedObjectValueOf[authorizingClaimMatchValueTypeModel] `tfsdk:"authorizing_claim_match_value"`
	InboundTokenClaimName      types.String                                                         `tfsdk:"inbound_token_claim_name"`
	InboundTokenClaimValueType fwtypes.StringEnum[awstypes.InboundTokenClaimValueType]              `tfsdk:"inbound_token_claim_value_type"`
}

type authorizingClaimMatchValueTypeModel struct {
	ClaimMatchOperator fwtypes.StringEnum[awstypes.ClaimMatchOperatorType]       `tfsdk:"claim_match_operator"`
	ClaimMatchValue    fwtypes.ListNestedObjectValueOf[claimMatchValueTypeModel] `tfsdk:"claim_match_value"`
}

type claimMatchValueTypeModel struct {
	MatchValueString     types.String        `tfsdk:"match_value_string"`
	MatchValueStringList fwtypes.SetOfString `tfsdk:"match_value_string_list"`
}

var (
	_ fwflex.Expander  = claimMatchValueTypeModel{}
	_ fwflex.Flattener = &claimMatchValueTypeModel{}
)

func (m *claimMatchValueTypeModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.ClaimMatchValueTypeMemberMatchValueString:
		m.MatchValueString = types.StringValue(t.Value)

	case awstypes.ClaimMatchValueTypeMemberMatchValueStringList:
		m.MatchValueStringList = fwflex.FlattenFrameworkStringValueSetOfString(ctx, t.Value)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("claimMatchValueTypeModel.Flatten: %T", v),
		)
	}

	return diags
}

func (m claimMatchValueTypeModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.MatchValueString.IsNull():
		var r awstypes.ClaimMatchValueTypeMemberMatchValueString
		r.Value = fwflex.StringValueFromFramework(ctx, m.MatchValueString)
		return &r, diags

	case !m.MatchValueStringList.IsNull():
		var r awstypes.ClaimMatchValueTypeMemberMatchValueStringList
		r.Value = fwflex.ExpandFrameworkStringValueSet(ctx, m.MatchValueStringList)
		return &r, diags
	}

	return nil, diags
}
