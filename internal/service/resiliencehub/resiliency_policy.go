// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_resiliencehub_resiliency_policy", name="Resiliency Policy")
// @Tags(identifierAttribute="arn")
func newResourceResiliencyPolicy(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResiliencyPolicy{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameResiliencyPolicy = "Resiliency Policy"
)

type resourceResiliencyPolicy struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceResiliencyPolicy) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_resiliencehub_resiliency_policy"
}

func (r *resourceResiliencyPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiredObjAttrs := map[string]schema.Attribute{
		"rto_in_secs": schema.Int32Attribute{
			Description: "Recovery Time Objective (RTO) in seconds.",
			Required:    true,
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
				int32planmodifier.RequiresReplace(),
			},
			Validators: []validator.Int32{
				int32validator.AtLeast(0),
			},
		},
		"rpo_in_secs": schema.Int32Attribute{
			Description: "Recovery Point Objective (RPO) in seconds.",
			Required:    true,
			PlanModifiers: []planmodifier.Int32{
				int32planmodifier.UseStateForUnknown(),
				int32planmodifier.RequiresReplace(),
			},
			Validators: []validator.Int32{
				int32validator.AtLeast(0),
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:  framework.IDAttribute(),
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"policy_name": schema.StringAttribute{
				Description: "The name of the policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(
						"^[A-Za-z0-9][A-Za-z0-9_\\-]{1,59}$"), "Must match ^[A-Za-z0-9][A-Za-z0-9_\\-]{1,59}$"),
				},
			},
			"policy_description": schema.StringAttribute{
				Description: "The description for the policy.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(500),
				},
			},
			"data_location_constraint": schema.StringAttribute{
				Description: "Specifies a high-level geographical location constraint for where resilience policy data can be stored.",
				CustomType:  fwtypes.StringEnumType[awstypes.DataLocationConstraint](),
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tier": schema.StringAttribute{
				Description: "The tier for the resiliency policy, ranging from the highest severity (MissionCritical) to lowest (NonCritical).",
				CustomType:  fwtypes.StringEnumType[awstypes.ResiliencyPolicyTier](),
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"estimated_cost_tier": schema.StringAttribute{
				Description: "Specifies the estimated cost tier of the resiliency policy.",
				CustomType:  fwtypes.StringEnumType[awstypes.EstimatedCostTier](),
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"policy": schema.SingleNestedBlock{
				Description: "The resiliency failure policy.",
				CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyPolicyModel](ctx),
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Blocks: map[string]schema.Block{
					"az": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential availability zone disruptions.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.RequiresReplace(),
						},
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Attributes: requiredObjAttrs,
					},
					"hardware": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential infrastructure disruptions.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.RequiresReplace(),
						},
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Attributes: requiredObjAttrs,
					},
					"software": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential application disruptions.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.RequiresReplace(),
						},
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Attributes: requiredObjAttrs,
					},
					"region": schema.SingleNestedBlock{
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential region disruptions.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.RequiresReplace(),
						},
						Validators: []validator.Object{
							objectvalidator.AlsoRequires(
								path.MatchRelative().AtName("rto_in_secs"),
								path.MatchRelative().AtName("rpo_in_secs"),
							),
						},
						Attributes: map[string]schema.Attribute{
							"rto_in_secs": schema.Int32Attribute{
								Description: "Recovery Time Objective (RTO) in seconds.",
								Optional:    true,
								PlanModifiers: []planmodifier.Int32{
									int32planmodifier.UseStateForUnknown(),
									int32planmodifier.RequiresReplace(),
								},
								Validators: []validator.Int32{
									int32validator.AtLeast(0),
								},
							},
							"rpo_in_secs": schema.Int32Attribute{
								Description: "Recovery Point Objective (RPO) in seconds.",
								Optional:    true,
								PlanModifiers: []planmodifier.Int32{
									int32planmodifier.UseStateForUnknown(),
									int32planmodifier.RequiresReplace(),
								},
								Validators: []validator.Int32{
									int32validator.AtLeast(0),
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

func (r *resourceResiliencyPolicy) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceResiliencyPolicyData

	conn := r.Meta().ResilienceHubClient(ctx)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planPolicyModel, diags := plan.Policy.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	in := &resiliencehub.CreateResiliencyPolicyInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in, flex.WithIgnoredFieldNamesAppend("Policy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(planPolicyModel.expandPolicy(ctx, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.Tags = getTagsIn(ctx)

	out, err := conn.CreateResiliencyPolicy(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionCreating, ResNameResiliencyPolicy, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Policy == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionCreating, ResNameResiliencyPolicy, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.PolicyArn = flex.StringToFramework(ctx, out.Policy.PolicyArn)
	plan.setId()

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitResiliencyPolicyCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForCreation, ResNameResiliencyPolicy, plan.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, created, &plan, flex.WithIgnoredFieldNamesAppend("Policy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(plan.flattenPolicy(ctx, created.Policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResiliencyPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceResiliencyPolicyData

	conn := r.Meta().ResilienceHubClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResiliencyPolicyByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionSetting, ResNameResiliencyPolicy, state.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.PolicyArn)

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithIgnoredFieldNamesAppend("Policy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(state.flattenPolicy(ctx, out.Policy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceResiliencyPolicy) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceResiliencyPolicyData

	conn := r.Meta().ResilienceHubClient(ctx)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.PolicyDescription.Equal(state.PolicyDescription) ||
		!plan.DataLocationConstraint.Equal(state.DataLocationConstraint) ||
		!plan.Tier.Equal(state.Tier) ||
		!plan.PolicyName.Equal(state.PolicyName) {

		in := &resiliencehub.UpdateResiliencyPolicyInput{
			PolicyArn: flex.StringFromFramework(ctx, plan.ID),
		}

		if !plan.PolicyName.Equal(state.PolicyName) {
			in.PolicyName = flex.StringFromFramework(ctx, plan.PolicyName)
		}

		if !plan.PolicyDescription.Equal(state.PolicyDescription) {
			in.PolicyDescription = flex.StringFromFramework(ctx, plan.PolicyDescription)
		}

		if !plan.DataLocationConstraint.Equal(state.DataLocationConstraint) {
			in.DataLocationConstraint = plan.DataLocationConstraint.ValueEnum()
		}

		if !plan.Tier.Equal(state.Tier) {
			in.Tier = plan.Tier.ValueEnum()
		}

		out, err := conn.UpdateResiliencyPolicy(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("reading Resilience Hub policy ID (%s)", plan.PolicyArn.String()), err.Error())
			return
		}

		plan.ID = flex.StringToFramework(ctx, out.Policy.PolicyArn)
		plan.setId()

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		updated, err := waitResiliencyPolicyUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForUpdate, ResNameResiliencyPolicy, plan.PolicyArn.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, updated, &plan, flex.WithIgnoredFieldNamesAppend("Policy"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(state.flattenPolicy(ctx, updated.Policy)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceResiliencyPolicy) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceResiliencyPolicyData

	conn := r.Meta().ResilienceHubClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteResiliencyPolicy(ctx, &resiliencehub.DeleteResiliencyPolicyInput{
		PolicyArn: flex.StringFromFramework(ctx, state.ID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting Resilience Hub policy name (%s)", state.PolicyName.String()), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitResiliencyPolicyDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForDeletion, ResNameResiliencyPolicy, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResiliencyPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceResiliencyPolicy) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

const (
	resiliencyPolicyTypeAZ       = "AZ"
	resiliencyPolicyTypeHardware = "Hardware"
	resiliencyPolicyTypeSoftware = "Software"
	resiliencyPolicyTypeRegion   = "Region"
	statusChangePending          = "Pending"
	statusDeleting               = "Deleting"
	statusNormal                 = "Normal"
	statusUpdated                = "Updated"
	statusCompleted              = "Completed"
)

func waitResiliencyPolicyCreated(ctx context.Context, conn *resiliencehub.Client, id string, timeout time.Duration) (*awstypes.ResiliencyPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusCompleted},
		Refresh:                   statusResiliencyPolicy(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ResiliencyPolicy); ok {
		return out, err
	}

	return nil, err
}

func waitResiliencyPolicyUpdated(ctx context.Context, conn *resiliencehub.Client, id string, timeout time.Duration) (*awstypes.ResiliencyPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusCompleted},
		Refresh:                   statusResiliencyPolicy(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ResiliencyPolicy); ok {
		return out, err
	}

	return nil, err
}

func waitResiliencyPolicyDeleted(ctx context.Context, conn *resiliencehub.Client, id string, timeout time.Duration) (*awstypes.ResiliencyPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{},
		Refresh: statusResiliencyPolicy(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ResiliencyPolicy); ok {
		return out, err
	}

	return nil, err
}

func statusResiliencyPolicy(ctx context.Context, conn *resiliencehub.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findResiliencyPolicyByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusCompleted, nil
	}
}

func findResiliencyPolicyByID(ctx context.Context, conn *resiliencehub.Client, id string) (*awstypes.ResiliencyPolicy, error) {
	in := &resiliencehub.DescribeResiliencyPolicyInput{
		PolicyArn: aws.String(id),
	}

	out, err := conn.DescribeResiliencyPolicy(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Policy == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Policy, nil
}

func (r *resourceResiliencyPolicyData) setId() {
	r.ID = r.PolicyArn
}

func (m *resourceResiliencyPolicyModel) expandPolicy(ctx context.Context, in *resiliencehub.CreateResiliencyPolicyInput) (diags diag.Diagnostics) {

	if in == nil {
		return
	}

	failurePolicy := make(map[string]awstypes.FailurePolicy)

	// Policy key case must be modified to align with key case in CreateResiliencyPolicy API documentation.
	// See https://docs.aws.amazon.com/resilience-hub/latest/APIReference/API_CreateResiliencyPolicy.html
	failurePolicyKeyMap := map[string]fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel]{
		resiliencyPolicyTypeAZ:       m.AZ,
		resiliencyPolicyTypeHardware: m.Hardware,
		resiliencyPolicyTypeSoftware: m.Software,
		resiliencyPolicyTypeRegion:   m.Region}

	for k, v := range failurePolicyKeyMap {
		if !v.IsNull() {
			resObjModel, d := v.ToPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return
			}
			failurePolicy[k] = awstypes.FailurePolicy{
				RpoInSecs: resObjModel.RpoInSecs.ValueInt32(),
				RtoInSecs: resObjModel.RtoInSecs.ValueInt32(),
			}
		}
	}

	in.Policy = failurePolicy

	return
}

func (m *resourceResiliencyPolicyData) flattenPolicy(ctx context.Context, failurePolicy map[string]awstypes.FailurePolicy) (diags diag.Diagnostics) {

	if len(failurePolicy) == 0 {
		m.Policy = fwtypes.NewObjectValueOfNull[resourceResiliencyPolicyModel](ctx)
		return
	}

	newResObjModel := func(policyType string, failurePolicy map[string]awstypes.FailurePolicy) fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] {
		if pv, exists := failurePolicy[policyType]; exists {
			return fwtypes.NewObjectValueOfMust(ctx, &resourceResiliencyObjectiveModel{
				RpoInSecs: types.Int32Value(pv.RpoInSecs),
				RtoInSecs: types.Int32Value(pv.RtoInSecs)})
		} else {
			return fwtypes.NewObjectValueOfNull[resourceResiliencyObjectiveModel](ctx)
		}
	}

	m.Policy = fwtypes.NewObjectValueOfMust(ctx, &resourceResiliencyPolicyModel{
		newResObjModel(resiliencyPolicyTypeAZ, failurePolicy),
		newResObjModel(resiliencyPolicyTypeHardware, failurePolicy),
		newResObjModel(resiliencyPolicyTypeSoftware, failurePolicy),
		newResObjModel(resiliencyPolicyTypeRegion, failurePolicy),
	})

	return
}

type resourceResiliencyPolicyData struct {
	DataLocationConstraint fwtypes.StringEnum[awstypes.DataLocationConstraint]  `tfsdk:"data_location_constraint"`
	EstimatedCostTier      fwtypes.StringEnum[awstypes.EstimatedCostTier]       `tfsdk:"estimated_cost_tier"`
	ID                     types.String                                         `tfsdk:"id"`
	Policy                 fwtypes.ObjectValueOf[resourceResiliencyPolicyModel] `tfsdk:"policy"`
	PolicyArn              types.String                                         `tfsdk:"arn"`
	PolicyDescription      types.String                                         `tfsdk:"policy_description"`
	PolicyName             types.String                                         `tfsdk:"policy_name"`
	Tier                   fwtypes.StringEnum[awstypes.ResiliencyPolicyTier]    `tfsdk:"tier"`
	Tags                   types.Map                                            `tfsdk:"tags"`
	TagsAll                types.Map                                            `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                       `tfsdk:"timeouts"`
}

type resourceResiliencyPolicyModel struct {
	AZ       fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"az"`
	Hardware fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"hardware"`
	Software fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"software"`
	Region   fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"region"`
}

type resourceResiliencyObjectiveModel struct {
	RpoInSecs types.Int32 `tfsdk:"rpo_in_secs"`
	RtoInSecs types.Int32 `tfsdk:"rto_in_secs"`
}
