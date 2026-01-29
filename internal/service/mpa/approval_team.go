// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mpa

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mpa"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mpa/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

// @FrameworkResource("aws_mpa_approval_team", name="Approval Team")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/mpa;mpa.GetApprovalTeamOutput")
func newApprovalTeamResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &approvalTeamResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type approvalTeamResource struct {
	framework.ResourceWithModel[approvalTeamResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *approvalTeamResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
			"last_update_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"number_of_approvers": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatusCode: schema.StringAttribute{
				Computed: true,
			},
			"status_message": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"version_id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"approval_strategy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[approvalStrategyModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"m_of_n": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[mofNApprovalStrategyModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"min_approvals_required": schema.Int64Attribute{
										Required: true,
										Validators: []validator.Int64{
											int64validator.AtLeast(1),
										},
									},
								},
							},
						},
					},
				},
			},
			"approver": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[approverModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"primary_identity_id": schema.StringAttribute{
							Required: true,
						},
						"primary_identity_source_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			"policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[policyReferenceModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"policy_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
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

func (r *approvalTeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan approvalTeamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MPAClient(ctx)

	name := plan.Name.ValueString()
	input := mpa.CreateApprovalTeamInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Description: fwflex.StringFromFramework(ctx, plan.Description),
		Name:        fwflex.StringFromFramework(ctx, plan.Name),
		Tags:        getTagsIn(ctx),
	}

	resp.Diagnostics.Append(expandApprovalStrategy(ctx, plan.ApprovalStrategy, &input.ApprovalStrategy)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(expandApprovers(ctx, plan.Approvers, &input.Approvers)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(expandPolicies(ctx, plan.Policies, &input.Policies)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateApprovalTeam(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("creating MPA Approval Team", err.Error())
		return
	}

	plan.ARN = fwflex.StringToFramework(ctx, output.Arn)
	plan.ID = fwflex.StringToFramework(ctx, output.Arn)
	plan.VersionID = fwflex.StringToFramework(ctx, output.VersionId)
	plan.CreationTime = fwflex.TimeToFramework(ctx, output.CreationTime)

	team, err := findApprovalTeamByARN(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		resp.State.SetAttribute(ctx, path.Root(names.AttrID), plan.ID)
		resp.Diagnostics.AddError("reading MPA Approval Team ("+name+") after create", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenApprovalTeamResponse(ctx, team, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *approvalTeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state approvalTeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MPAClient(ctx)

	team, err := findApprovalTeamByARN(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading MPA Approval Team ("+state.ID.ValueString()+")", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenApprovalTeamResponse(ctx, team, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *approvalTeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state approvalTeamResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MPAClient(ctx)

	if !plan.ApprovalStrategy.Equal(state.ApprovalStrategy) ||
		!plan.Approvers.Equal(state.Approvers) ||
		!plan.Description.Equal(state.Description) {
		input := mpa.UpdateApprovalTeamInput{
			Arn:         fwflex.StringFromFramework(ctx, plan.ARN),
			Description: fwflex.StringFromFramework(ctx, plan.Description),
		}

		resp.Diagnostics.Append(expandApprovalStrategy(ctx, plan.ApprovalStrategy, &input.ApprovalStrategy)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(expandApprovers(ctx, plan.Approvers, &input.Approvers)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateApprovalTeam(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError("updating MPA Approval Team ("+plan.ID.ValueString()+")", err.Error())
			return
		}
	}

	team, err := findApprovalTeamByARN(ctx, conn, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("reading MPA Approval Team ("+plan.ID.ValueString()+") after update", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenApprovalTeamResponse(ctx, team, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *approvalTeamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state approvalTeamResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().MPAClient(ctx)

	team, err := findApprovalTeamByARN(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading MPA Approval Team ("+state.ID.ValueString()+") for delete", err.Error())
		return
	}

	input := mpa.DeleteInactiveApprovalTeamVersionInput{
		Arn:       fwflex.StringFromFramework(ctx, state.ARN),
		VersionId: team.VersionId,
	}

	_, err = conn.DeleteInactiveApprovalTeamVersion(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("deleting MPA Approval Team ("+state.ID.ValueString()+")", err.Error())
		return
	}
}

func (r *approvalTeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

// Resource model types

type approvalTeamResourceModel struct {
	ApprovalStrategy  fwtypes.ListNestedObjectValueOf[approvalStrategyModel] `tfsdk:"approval_strategy"`
	Approvers         fwtypes.ListNestedObjectValueOf[approverModel]         `tfsdk:"approver"`
	ARN               types.String                                           `tfsdk:"arn"`
	CreationTime      timetypes.RFC3339                                      `tfsdk:"creation_time"`
	Description       types.String                                           `tfsdk:"description"`
	ID                types.String                                           `tfsdk:"id"`
	LastUpdateTime    timetypes.RFC3339                                      `tfsdk:"last_update_time"`
	Name              types.String                                           `tfsdk:"name"`
	NumberOfApprovers types.Int64                                            `tfsdk:"number_of_approvers"`
	Policies          fwtypes.ListNestedObjectValueOf[policyReferenceModel]  `tfsdk:"policy"`
	Status            types.String                                           `tfsdk:"status"`
	StatusCode        types.String                                           `tfsdk:"status_code"`
	StatusMessage     types.String                                           `tfsdk:"status_message"`
	Tags              tftags.Map                                             `tfsdk:"tags"`
	TagsAll           tftags.Map                                             `tfsdk:"tags_all"`
	Timeouts          timeouts.Value                                         `tfsdk:"timeouts"`
	VersionID         types.String                                           `tfsdk:"version_id"`
}

type approvalStrategyModel struct {
	MofN fwtypes.ListNestedObjectValueOf[mofNApprovalStrategyModel] `tfsdk:"m_of_n"`
}

type mofNApprovalStrategyModel struct {
	MinApprovalsRequired types.Int64 `tfsdk:"min_approvals_required"`
}

type approverModel struct {
	PrimaryIdentityID        types.String `tfsdk:"primary_identity_id"`
	PrimaryIdentitySourceARN fwtypes.ARN  `tfsdk:"primary_identity_source_arn"`
}

type policyReferenceModel struct {
	PolicyARN fwtypes.ARN `tfsdk:"policy_arn"`
}

// Helper functions

func findApprovalTeamByARN(ctx context.Context, conn *mpa.Client, arn string) (*mpa.GetApprovalTeamOutput, error) {
	input := mpa.GetApprovalTeamInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetApprovalTeam(ctx, &input)
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

func expandApprovalStrategy(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[approvalStrategyModel], apiObject *awstypes.ApprovalStrategy) diag.Diagnostics {
	var diags diag.Diagnostics

	data, d := tfList.ToPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if data == nil {
		return diags
	}

	if !data.MofN.IsNull() && !data.MofN.IsUnknown() {
		mofnData, d := data.MofN.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		if mofnData != nil {
			*apiObject = &awstypes.ApprovalStrategyMemberMofN{
				Value: awstypes.MofNApprovalStrategy{
					MinApprovalsRequired: aws.Int32(int32(mofnData.MinApprovalsRequired.ValueInt64())),
				},
			}
		}
	}

	return diags
}

func expandApprovers(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[approverModel], apiObject *[]awstypes.ApprovalTeamRequestApprover) diag.Diagnostics {
	var diags diag.Diagnostics

	data, d := tfList.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	result := make([]awstypes.ApprovalTeamRequestApprover, 0, len(data))
	for _, item := range data {
		result = append(result, awstypes.ApprovalTeamRequestApprover{
			PrimaryIdentityId:        fwflex.StringFromFramework(ctx, item.PrimaryIdentityID),
			PrimaryIdentitySourceArn: item.PrimaryIdentitySourceARN.ValueStringPointer(),
		})
	}

	*apiObject = result
	return diags
}

func expandPolicies(ctx context.Context, tfList fwtypes.ListNestedObjectValueOf[policyReferenceModel], apiObject *[]awstypes.PolicyReference) diag.Diagnostics {
	var diags diag.Diagnostics

	data, d := tfList.ToSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	result := make([]awstypes.PolicyReference, 0, len(data))
	for _, item := range data {
		result = append(result, awstypes.PolicyReference{
			PolicyArn: item.PolicyARN.ValueStringPointer(),
		})
	}

	*apiObject = result
	return diags
}

func flattenApprovalTeamResponse(ctx context.Context, apiObject *mpa.GetApprovalTeamOutput, data *approvalTeamResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ARN = fwflex.StringToFramework(ctx, apiObject.Arn)
	data.ID = fwflex.StringToFramework(ctx, apiObject.Arn)
	data.Name = fwflex.StringToFramework(ctx, apiObject.Name)
	data.Description = fwflex.StringToFramework(ctx, apiObject.Description)
	data.CreationTime = fwflex.TimeToFramework(ctx, apiObject.CreationTime)
	data.LastUpdateTime = fwflex.TimeToFramework(ctx, apiObject.LastUpdateTime)
	data.NumberOfApprovers = fwflex.Int32ToFrameworkInt64(ctx, apiObject.NumberOfApprovers)
	data.Status = fwflex.StringValueToFramework(ctx, string(apiObject.Status))
	data.StatusCode = fwflex.StringValueToFramework(ctx, string(apiObject.StatusCode))
	data.StatusMessage = fwflex.StringToFramework(ctx, apiObject.StatusMessage)
	data.VersionID = fwflex.StringToFramework(ctx, apiObject.VersionId)

	diags.Append(flattenApprovalStrategyResponse(ctx, apiObject.ApprovalStrategy, &data.ApprovalStrategy)...)
	if diags.HasError() {
		return diags
	}

	diags.Append(flattenApproversResponse(ctx, apiObject.Approvers, &data.Approvers)...)
	if diags.HasError() {
		return diags
	}

	diags.Append(flattenPoliciesResponse(ctx, apiObject.Policies, &data.Policies)...)
	if diags.HasError() {
		return diags
	}

	return diags
}

func flattenApprovalStrategyResponse(ctx context.Context, apiObject awstypes.ApprovalStrategyResponse, tfObject *fwtypes.ListNestedObjectValueOf[approvalStrategyModel]) diag.Diagnostics {
	var diags diag.Diagnostics

	switch v := apiObject.(type) {
	case *awstypes.ApprovalStrategyResponseMemberMofN:
		mofnModel := mofNApprovalStrategyModel{
			MinApprovalsRequired: fwflex.Int32ToFrameworkInt64(ctx, v.Value.MinApprovalsRequired),
		}
		strategyModel := approvalStrategyModel{
			MofN: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &mofnModel),
		}
		*tfObject = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &strategyModel)
	}

	return diags
}

func flattenApproversResponse(ctx context.Context, apiObject []awstypes.GetApprovalTeamResponseApprover, tfObject *fwtypes.ListNestedObjectValueOf[approverModel]) diag.Diagnostics {
	var diags diag.Diagnostics

	result := make([]*approverModel, 0, len(apiObject))
	for _, item := range apiObject {
		model := &approverModel{
			PrimaryIdentityID:        fwflex.StringToFramework(ctx, item.PrimaryIdentityId),
			PrimaryIdentitySourceARN: fwtypes.ARNValue(aws.ToString(item.PrimaryIdentitySourceArn)),
		}
		result = append(result, model)
	}

	*tfObject = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, result)
	return diags
}

func flattenPoliciesResponse(ctx context.Context, apiObject []awstypes.PolicyReference, tfObject *fwtypes.ListNestedObjectValueOf[policyReferenceModel]) diag.Diagnostics {
	var diags diag.Diagnostics

	result := make([]*policyReferenceModel, 0, len(apiObject))
	for _, item := range apiObject {
		model := &policyReferenceModel{
			PolicyARN: fwtypes.ARNValue(aws.ToString(item.PolicyArn)),
		}
		result = append(result, model)
	}

	*tfObject = fwtypes.NewListNestedObjectValueOfSliceMust(ctx, result)
	return diags
}
