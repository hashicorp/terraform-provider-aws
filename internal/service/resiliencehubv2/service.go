// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehubv2_service", name="Service")
// @Tags(identifierAttribute="arn")
// @ArnIdentity
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.Service")
// @Testing(hasNoPreExistingResource=true)
func newServiceResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &serviceResource{}, nil
}

type serviceResource struct {
	framework.ResourceWithModel[serviceResourceModel]
	framework.WithImportByIdentity
}

// ModifyPlan marks the computed-only dev_ops_agent_status block as unknown whenever it may
// change on apply: on create, and on update when the writable dev_ops_agent_configuration
// changes (connect/reconnect/disconnect). The DevOps Agent connection status is resolved by
// the service (CONNECTING synchronously, then CONNECTED/FAILED out-of-band, DISCONNECTED on
// opt-out), so a stale planned value would otherwise trigger "inconsistent result after apply".
// When the configuration is unchanged, the prior value is retained to keep plans quiet.
func (r *serviceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// No plan on destroy.
	if req.Plan.Raw.IsNull() {
		return
	}

	// Create: no prior state. Status is resolved by the service, so mark it unknown.
	if req.State.Raw.IsNull() {
		setDevOpsAgentStatusUnknown(ctx, resp)
		return
	}

	var plan, state serviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update: if the writable configuration changes, the status will be recomputed on apply, so
	// mark it unknown. Otherwise pin the status attributes to their prior state values (like
	// UseStateForUnknown) so an unrelated update (e.g. tags-only) does not leave them unknown —
	// Update only re-flattens when the diff has changes, so unresolved unknowns would otherwise
	// persist into state. Only the DevOps Agent status attributes are touched here; everything
	// else (including the framework-managed region) is left as planned.
	if !plan.DevOpsAgentConfiguration.Equal(state.DevOpsAgentConfiguration) {
		setDevOpsAgentStatusUnknown(ctx, resp)
	} else {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_status"), state.DevOpsAgentStatus)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_space_id"), state.DevOpsAgentSpaceID)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_region"), state.DevOpsAgentRegion)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_operator_app_url"), state.DevOpsAgentOperatorAppURL)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_connected_at"), state.DevOpsAgentConnectedAt)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_status_reason"), state.DevOpsAgentStatusReason)...)
	}
}

// setDevOpsAgentStatusUnknown marks every computed DevOps Agent status attribute unknown so the
// value resolved on apply (which the service controls) does not conflict with the plan. It sets
// only those attributes, leaving the rest of the plan intact.
func setDevOpsAgentStatusUnknown(ctx context.Context, resp *resource.ModifyPlanResponse) {
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_status"), fwtypes.StringEnumUnknown[awstypes.DevOpsAgentConnectionStatus]())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_space_id"), types.StringUnknown())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_region"), types.StringUnknown())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_operator_app_url"), types.StringUnknown())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_connected_at"), timetypes.NewRFC3339Unknown())...)
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("dev_ops_agent_status_reason"), types.StringUnknown())...)
}

func (r *serviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"dependency_discovery": fwschema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DependencyDiscoveryInput](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 615),
				},
			},
			// Computed-only, read-side view of the DevOps Agent connection. The provider is
			// served over Terraform protocol v5 (via the SDKv2 mux), which does not support
			// nested attributes, and blocks cannot be marked unknown at plan time. These are
			// therefore flattened to top-level computed scalar attributes, which ModifyPlan can
			// mark unknown when the connection may change (status transitions out-of-band:
			// CONNECTING -> CONNECTED/FAILED, and on disconnect -> DISCONNECTED).
			"dev_ops_agent_status": fwschema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DevOpsAgentConnectionStatus](),
				Computed:   true,
			},
			"dev_ops_agent_space_id": fwschema.StringAttribute{
				Computed: true,
			},
			"dev_ops_agent_region": fwschema.StringAttribute{
				Computed: true,
			},
			"dev_ops_agent_operator_app_url": fwschema.StringAttribute{
				Computed: true,
			},
			"dev_ops_agent_connected_at": fwschema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"dev_ops_agent_status_reason": fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyID: fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validResourceName,
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_arn": fwschema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"regions": fwschema.SetAttribute{
				CustomType: fwtypes.SetOfStringType,
				Required:   true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 5),
					setvalidator.ValueStringsAre(fwvalidators.AWSRegion()),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"dev_ops_agent_configuration": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[devOpsAgentConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"agent_space_arn": fwschema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
			"associated_system": fwschema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[associatedSystemModel](ctx),
				Validators: []validator.Set{
					setvalidator.SizeBetween(0, 20),
				},
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"system_arn": fwschema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
						"user_journey_ids": fwschema.SetAttribute{
							CustomType: fwtypes.SetOfStringType,
							Optional:   true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 20),
							},
						},
					},
				},
			},
			"permission_model": fwschema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[permissionModelModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"invoker_role_name": fwschema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"cross_account_role": fwschema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[crossAccountRoleModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeBetween(0, 5),
							},
							NestedObject: fwschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									"cross_account_role_arn": fwschema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									names.AttrExternalID: fwschema.StringAttribute{
										Optional: true,
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

func (r *serviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	name := fwflex.StringValueFromFramework(ctx, plan.Name)
	var input resiliencehubv2.CreateServiceInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := tfresource.RetryWhen(ctx, propagationTimeout, func(ctx context.Context) (*resiliencehubv2.CreateServiceOutput, error) {
		return conn.CreateService(ctx, &input)
	}, isServiceRoleNotReady)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.Name, name)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output.Service, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *serviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ServiceARN)
	svc, err := findServiceByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, svc, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *serviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, plan.ServiceARN)
		var input resiliencehubv2.UpdateServiceInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		// UpdateService treats an omitted AssociatedSystems as "no change", leaving the
		// existing associations in place. Removing the last associated_system block
		// therefore requires sending an explicitly empty list.
		if input.AssociatedSystems == nil && state.AssociatedSystems.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0 {
			input.AssociatedSystems = []awstypes.AssociatedSystem{}
		}

		// UpdateService treats an omitted DevOpsAgentConfiguration as "no change". To
		// disconnect, the API expects the block present but without an agentSpaceArn.
		// So when the plan removes dev_ops_agent_configuration that previously existed,
		// send an explicit empty configuration to clear the linkage.
		if input.DevOpsAgentConfiguration == nil && state.DevOpsAgentConfiguration.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0 {
			input.DevOpsAgentConfiguration = &awstypes.DevOpsAgentConfiguration{}
		}

		output, err := tfresource.RetryWhen(ctx, propagationTimeout, func(ctx context.Context) (*resiliencehubv2.UpdateServiceOutput, error) {
			return conn.UpdateService(ctx, &input)
		}, isServiceRoleNotReady)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		// Re-flatten the updated service so computed values (e.g. dev_ops_agent_status)
		// reflect the server's post-update state rather than the prior plan.
		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, output.Service, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *serviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serviceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	arn := fwflex.StringValueFromFramework(ctx, state.ServiceARN)
	input := resiliencehubv2.DeleteServiceInput{
		ServiceArn: aws.String(arn),
	}
	_, err := conn.DeleteService(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
	}
}

func (r *serviceResource) flatten(ctx context.Context, svc *awstypes.Service, data *serviceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, svc, data)...)
	if diags.HasError() {
		return diags
	}

	data.DependencyDiscovery = fwtypes.StringEnumNull[awstypes.DependencyDiscoveryInput]()
	if v := svc.DependencyDiscovery; v != nil {
		switch v.Status {
		case awstypes.DependencyDiscoveryStatusEnabled, awstypes.DependencyDiscoveryStatusInitializing:
			data.DependencyDiscovery = fwtypes.StringEnumValue(awstypes.DependencyDiscoveryInputEnabled)
		case awstypes.DependencyDiscoveryStatusDisabled:
			data.DependencyDiscovery = fwtypes.StringEnumValue(awstypes.DependencyDiscoveryInputDisabled)
		}
	}

	// The DevOps Agent integration surfaces as the writable dev_ops_agent_configuration
	// block (agent_space_arn only) plus the computed-only, flattened dev_ops_agent_* status
	// attributes. Both are derived from the single read-side view (svc.DevOpsAgentConfiguration).
	//
	// The service returns the view whenever the service has ever been linked: while
	// CONNECTING/CONNECTED the view carries an agent_space_arn, and after a disconnect the view
	// is still returned with status=DISCONNECTED but a nil agent_space_arn. The view is omitted
	// entirely only for a service that was never linked. So key the writable
	// dev_ops_agent_configuration on a non-nil agent_space_arn (null once disconnected, keeping
	// it in sync with a config that no longer sets it), and populate the status attributes
	// whenever the view is present (so a disconnected service still reports DISCONNECTED).
	data.DevOpsAgentConfiguration = fwtypes.NewListNestedObjectValueOfNull[devOpsAgentConfigurationModel](ctx)
	data.DevOpsAgentStatus = fwtypes.StringEnumNull[awstypes.DevOpsAgentConnectionStatus]()
	data.DevOpsAgentSpaceID = types.StringNull()
	data.DevOpsAgentRegion = types.StringNull()
	data.DevOpsAgentOperatorAppURL = types.StringNull()
	data.DevOpsAgentConnectedAt = timetypes.NewRFC3339Null()
	data.DevOpsAgentStatusReason = types.StringNull()
	if v := svc.DevOpsAgentConfiguration; v != nil {
		if v.AgentSpaceArn != nil {
			data.DevOpsAgentConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &devOpsAgentConfigurationModel{
				AgentSpaceARN: fwtypes.ARNValue(aws.ToString(v.AgentSpaceArn)),
			})
		}
		if v.Status != "" {
			data.DevOpsAgentStatus = fwtypes.StringEnumValue(v.Status)
		}
		data.DevOpsAgentSpaceID = fwflex.StringToFramework(ctx, v.AgentSpaceId)
		data.DevOpsAgentRegion = fwflex.StringToFramework(ctx, v.Region)
		data.DevOpsAgentOperatorAppURL = fwflex.StringToFramework(ctx, v.OperatorAppUrl)
		data.DevOpsAgentStatusReason = fwflex.StringToFramework(ctx, v.StatusReason)
		if v.ConnectedAt != nil {
			data.DevOpsAgentConnectedAt = timetypes.NewRFC3339TimeValue(aws.ToTime(v.ConnectedAt))
		}
	}

	setTagsOut(ctx, svc.Tags)

	return diags
}

// isServiceRoleNotReady marks the IAM-propagation races that CreateService/UpdateService can hit
// as retryable. The invoker role (and, when linking a DevOps Agent Agent Space, its aidevops
// permissions) may have just been created, so Resilience Hub's cross-account assume/validation
// can transiently fail until IAM propagates.
func isServiceRoleNotReady(err error) (bool, error) {
	if err == nil {
		return false, nil
	}
	// Invoker role / trust policy not yet usable.
	if errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Ensure the role exists and its trust policy allows access") {
		return true, err
	}
	// Role lacks (freshly granted) DevOps Agent permissions.
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "not authorized to call DevOps Agent") {
		return true, err
	}
	return false, err
}

func findServiceByARN(ctx context.Context, conn *resiliencehubv2.Client, arn string) (*awstypes.Service, error) {
	input := resiliencehubv2.GetServiceInput{
		ServiceArn: aws.String(arn),
	}

	return findService(ctx, conn, &input)
}

func findService(ctx context.Context, conn *resiliencehubv2.Client, input *resiliencehubv2.GetServiceInput) (*awstypes.Service, error) {
	output, err := conn.GetService(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil || output.Service == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output.Service, nil
}

type serviceResourceModel struct {
	framework.WithRegionModel
	AssociatedSystems         fwtypes.SetNestedObjectValueOf[associatedSystemModel]          `tfsdk:"associated_system"`
	DependencyDiscovery       fwtypes.StringEnum[awstypes.DependencyDiscoveryInput]          `tfsdk:"dependency_discovery" autoflex:",noflatten"`
	Description               types.String                                                   `tfsdk:"description"`
	DevOpsAgentConfiguration  fwtypes.ListNestedObjectValueOf[devOpsAgentConfigurationModel] `tfsdk:"dev_ops_agent_configuration" autoflex:",noflatten"`
	DevOpsAgentConnectedAt    timetypes.RFC3339                                              `tfsdk:"dev_ops_agent_connected_at" autoflex:"-"`
	DevOpsAgentOperatorAppURL types.String                                                   `tfsdk:"dev_ops_agent_operator_app_url" autoflex:"-"`
	DevOpsAgentRegion         types.String                                                   `tfsdk:"dev_ops_agent_region" autoflex:"-"`
	DevOpsAgentSpaceID        types.String                                                   `tfsdk:"dev_ops_agent_space_id" autoflex:"-"`
	DevOpsAgentStatus         fwtypes.StringEnum[awstypes.DevOpsAgentConnectionStatus]       `tfsdk:"dev_ops_agent_status" autoflex:"-"`
	DevOpsAgentStatusReason   types.String                                                   `tfsdk:"dev_ops_agent_status_reason" autoflex:"-"`
	KMSKeyID                  fwtypes.ARN                                                    `tfsdk:"kms_key_id"`
	Name                      types.String                                                   `tfsdk:"name"`
	PermissionModel           fwtypes.ListNestedObjectValueOf[permissionModelModel]          `tfsdk:"permission_model"`
	PolicyARN                 fwtypes.ARN                                                    `tfsdk:"policy_arn"`
	Regions                   fwtypes.SetOfString                                            `tfsdk:"regions"`
	ServiceARN                types.String                                                   `tfsdk:"arn"`
	Tags                      tftags.Map                                                     `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                     `tfsdk:"tags_all"`
}

type devOpsAgentConfigurationModel struct {
	AgentSpaceARN fwtypes.ARN `tfsdk:"agent_space_arn"`
}

type associatedSystemModel struct {
	SystemARN      fwtypes.ARN         `tfsdk:"system_arn"`
	UserJourneyIDs fwtypes.SetOfString `tfsdk:"user_journey_ids"`
}

type permissionModelModel struct {
	CrossAccountRole fwtypes.ListNestedObjectValueOf[crossAccountRoleModel] `tfsdk:"cross_account_role"`
	InvokerRoleName  types.String                                           `tfsdk:"invoker_role_name"`
}

type crossAccountRoleModel struct {
	CrossAccountRoleARN fwtypes.ARN  `tfsdk:"cross_account_role_arn"`
	ExternalID          types.String `tfsdk:"external_id"`
}
