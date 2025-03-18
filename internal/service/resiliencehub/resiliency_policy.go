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
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehub;resiliencehub.DescribeResiliencyPolicyOutput")
// @Testing(importStateIdAttribute="arn")
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
	framework.WithTimeouts
}

func (r *resourceResiliencyPolicy) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiredObjAttrs := map[string]schema.Attribute{
		"rto": schema.StringAttribute{
			Description: "Recovery Time Objective (RTO) as a Go duration.",
			CustomType:  timetypes.GoDurationType{},
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"rpo": schema.StringAttribute{
			Description: "Recovery Point Objective (RPO) as a Go duration.",
			CustomType:  timetypes.GoDurationType{},
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"data_location_constraint": schema.StringAttribute{
				Description: "Specifies a high-level geographical location constraint for where resilience policy data can be stored.",
				CustomType:  fwtypes.StringEnumType[awstypes.DataLocationConstraint](),
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Description: "The description for the policy.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtMost(500),
				},
			},
			names.AttrName: schema.StringAttribute{
				Description: "The name of the policy.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 60),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]+$`), "Must start with an alphanumeric character and contain alphanumeric characters, underscores, or hyphens"),
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
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrPolicy: schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				Description: "The resiliency failure policy.",
				CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyPolicyModel](ctx),
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Blocks: map[string]schema.Block{
					"az": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential availability zone disruptions.",
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Attributes: requiredObjAttrs,
					},
					"hardware": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential infrastructure disruptions.",
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Attributes: requiredObjAttrs,
					},
					"software": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential application disruptions.",
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Attributes: requiredObjAttrs,
					},
					names.AttrRegion: schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType:  fwtypes.NewObjectTypeOf[resourceResiliencyObjectiveModel](ctx),
						Description: "The RTO and RPO target to measure resiliency for potential region disruptions.",
						Validators: []validator.Object{
							objectvalidator.AlsoRequires(
								path.MatchRelative().AtName("rto"),
								path.MatchRelative().AtName("rpo"),
							),
						},
						Attributes: map[string]schema.Attribute{
							"rto": schema.StringAttribute{
								Description: "Recovery Time Objective (RTO) as a Go duration.",
								CustomType:  timetypes.GoDurationType{},
								Optional:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"rpo": schema.StringAttribute{
								Description: "Recovery Point Objective (RPO) as a Go duration.",
								CustomType:  timetypes.GoDurationType{},
								Optional:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
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

	// "Policy" cannot be handled by AutoFlex, since the parameter in the AWS API is a map
	planPolicyModel, diags := plan.Policy.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	var in resiliencehub.CreateResiliencyPolicyInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &in, flex.WithIgnoredFieldNamesAppend("Policy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, d := planPolicyModel.expandPolicy(ctx)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.Policy = p

	in.Tags = getTagsIn(ctx)

	out, err := conn.CreateResiliencyPolicy(ctx, &in)
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

	arn := aws.ToString(out.Policy.PolicyArn)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	created, err := waitResiliencyPolicyCreated(ctx, conn, arn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForCreation, ResNameResiliencyPolicy, plan.PolicyARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, created, &plan, flex.WithIgnoredFieldNamesAppend("Policy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.flattenPolicy(ctx, created.Policy)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResiliencyPolicy) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceResiliencyPolicyData

	conn := r.Meta().ResilienceHubClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResiliencyPolicyByARN(ctx, conn, state.PolicyARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionSetting, ResNameResiliencyPolicy, state.PolicyARN.ValueString(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state, flex.WithIgnoredFieldNamesAppend("Policy"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.flattenPolicy(ctx, out.Policy)

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

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		in := resiliencehub.UpdateResiliencyPolicyInput{
			PolicyArn: flex.StringFromFramework(ctx, plan.PolicyARN),
		}

		if !plan.PolicyDescription.Equal(state.PolicyDescription) {
			in.PolicyDescription = flex.StringFromFramework(ctx, plan.PolicyDescription)
		}

		if !plan.DataLocationConstraint.Equal(state.DataLocationConstraint) {
			in.DataLocationConstraint = plan.DataLocationConstraint.ValueEnum()
		}

		if !plan.Policy.Equal(state.Policy) {
			planPolicyModel, d := plan.Policy.ToPtr(ctx)
			resp.Diagnostics.Append(d...)
			if d.HasError() {
				return
			}
			p, d := planPolicyModel.expandPolicy(ctx)
			resp.Diagnostics.Append(d...)
			if d.HasError() {
				return
			}
			in.Policy = p
		}

		if !plan.PolicyName.Equal(state.PolicyName) {
			in.PolicyName = flex.StringFromFramework(ctx, plan.PolicyName)
		}

		if !plan.Tier.Equal(state.Tier) {
			in.Tier = plan.Tier.ValueEnum()
		}

		_, err := conn.UpdateResiliencyPolicy(ctx, &in)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("reading Resilience Hub policy ID (%s)", plan.PolicyARN.String()), err.Error())
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		updated, err := waitResiliencyPolicyUpdated(ctx, conn, plan.PolicyARN.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForUpdate, ResNameResiliencyPolicy, plan.PolicyARN.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, updated, &plan, flex.WithIgnoredFieldNamesAppend("Policy"))...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.flattenPolicy(ctx, updated.Policy)
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
		PolicyArn: flex.StringFromFramework(ctx, state.PolicyARN),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting Resilience Hub policy name (%s)", state.PolicyName.String()), err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitResiliencyPolicyDeleted(ctx, conn, state.PolicyARN.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ResilienceHub, create.ErrActionWaitingForDeletion, ResNameResiliencyPolicy, state.PolicyARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResiliencyPolicy) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
	statusCompleted     = "Completed"
)

func waitResiliencyPolicyCreated(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.ResiliencyPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusCompleted},
		Refresh:                   statusResiliencyPolicy(ctx, conn, arn),
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

func waitResiliencyPolicyUpdated(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.ResiliencyPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusCompleted},
		Refresh:                   statusResiliencyPolicy(ctx, conn, arn),
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

func waitResiliencyPolicyDeleted(ctx context.Context, conn *resiliencehub.Client, arn string, timeout time.Duration) (*awstypes.ResiliencyPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{},
		Refresh: statusResiliencyPolicy(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.ResiliencyPolicy); ok {
		return out, err
	}

	return nil, err
}

func statusResiliencyPolicy(ctx context.Context, conn *resiliencehub.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findResiliencyPolicyByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusCompleted, nil
	}
}

func findResiliencyPolicyByARN(ctx context.Context, conn *resiliencehub.Client, arn string) (*awstypes.ResiliencyPolicy, error) {
	in := &resiliencehub.DescribeResiliencyPolicyInput{
		PolicyArn: aws.String(arn),
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

func (m *resourceResiliencyPolicyModel) expandPolicy(ctx context.Context) (result map[string]awstypes.FailurePolicy, diags diag.Diagnostics) {
	failurePolicy := make(map[string]awstypes.FailurePolicy)

	// Policy key case must be modified to align with key case in CreateResiliencyPolicy API documentation.
	// See https://docs.aws.amazon.com/resilience-hub/latest/APIReference/API_CreateResiliencyPolicy.html
	failurePolicyKeyMap := map[awstypes.TestType]fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel]{
		awstypes.TestTypeAz:       m.AZ,
		awstypes.TestTypeHardware: m.Hardware,
		awstypes.TestTypeSoftware: m.Software,
		awstypes.TestTypeRegion:   m.Region,
	}
	for k, v := range failurePolicyKeyMap {
		if !v.IsNull() {
			resObjModel, d := v.ToPtr(ctx)
			diags.Append(d...)
			if d.HasError() {
				return result, diags
			}
			rpo, d := resObjModel.Rpo.ValueGoDuration()
			if d.HasError() {
				return result, diags
			}
			rto, d := resObjModel.Rto.ValueGoDuration()
			if d.HasError() {
				return result, diags
			}
			failurePolicy[string(k)] = awstypes.FailurePolicy{
				RpoInSecs: int32(rpo.Seconds()),
				RtoInSecs: int32(rto.Seconds()),
			}
		}
	}

	result = failurePolicy

	return result, diags
}

func (m *resourceResiliencyPolicyData) flattenPolicy(ctx context.Context, failurePolicy map[string]awstypes.FailurePolicy) {
	if len(failurePolicy) == 0 {
		m.Policy = fwtypes.NewObjectValueOfNull[resourceResiliencyPolicyModel](ctx)
	}

	newResObjModel := func(policyType awstypes.TestType, failurePolicy map[string]awstypes.FailurePolicy) fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] {
		if pv, exists := failurePolicy[string(policyType)]; exists {
			return fwtypes.NewObjectValueOfMust(ctx, &resourceResiliencyObjectiveModel{
				Rpo: timetypes.NewGoDurationValue(time.Duration(pv.RpoInSecs) * time.Second),
				Rto: timetypes.NewGoDurationValue(time.Duration(pv.RtoInSecs) * time.Second),
			})
		} else {
			return fwtypes.NewObjectValueOfNull[resourceResiliencyObjectiveModel](ctx)
		}
	}

	m.Policy = fwtypes.NewObjectValueOfMust(ctx, &resourceResiliencyPolicyModel{
		newResObjModel(awstypes.TestTypeAz, failurePolicy),
		newResObjModel(awstypes.TestTypeHardware, failurePolicy),
		newResObjModel(awstypes.TestTypeSoftware, failurePolicy),
		newResObjModel(awstypes.TestTypeRegion, failurePolicy),
	})
}

type resourceResiliencyPolicyData struct {
	DataLocationConstraint fwtypes.StringEnum[awstypes.DataLocationConstraint]  `tfsdk:"data_location_constraint"`
	EstimatedCostTier      fwtypes.StringEnum[awstypes.EstimatedCostTier]       `tfsdk:"estimated_cost_tier"`
	Policy                 fwtypes.ObjectValueOf[resourceResiliencyPolicyModel] `tfsdk:"policy"`
	PolicyARN              types.String                                         `tfsdk:"arn"`
	PolicyDescription      types.String                                         `tfsdk:"description"`
	PolicyName             types.String                                         `tfsdk:"name"`
	Tier                   fwtypes.StringEnum[awstypes.ResiliencyPolicyTier]    `tfsdk:"tier"`
	Tags                   tftags.Map                                           `tfsdk:"tags"`
	TagsAll                tftags.Map                                           `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                       `tfsdk:"timeouts"`
}

type resourceResiliencyPolicyModel struct {
	AZ       fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"az"`
	Hardware fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"hardware"`
	Software fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"software"`
	Region   fwtypes.ObjectValueOf[resourceResiliencyObjectiveModel] `tfsdk:"region"`
}

type resourceResiliencyObjectiveModel struct {
	Rpo timetypes.GoDuration `tfsdk:"rpo"`
	Rto timetypes.GoDuration `tfsdk:"rto"`
}
