// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
func newResourceAssessmentDelegation(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAssessmentDelegation{}, nil
}

const (
	ResNameAssessmentDelegation = "AssessmentDelegation"
)

type resourceAssessmentDelegation struct {
	framework.ResourceWithConfigure
}

func (r *resourceAssessmentDelegation) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_auditmanager_assessment_delegation"
}

func (r *resourceAssessmentDelegation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"assessment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrComment: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"control_set_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// The AWS-generated ID for delegations has been observed to change between creation
			// and subseqeunt read operations. As such, this value cannot be used as the resource ID
			// or the input to finder functions. However, it is still required as part of the delete
			// request input, so will be stored as a separate computed attribute.
			"delegation_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrRoleARN: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.RoleType](),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *resourceAssessmentDelegation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan resourceAssessmentDelegationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	delegationIn := awstypes.CreateDelegationRequest{
		RoleArn:      aws.String(plan.RoleARN.ValueString()),
		RoleType:     awstypes.RoleType(plan.RoleType.ValueString()),
		ControlSetId: aws.String(plan.ControlSetID.ValueString()),
	}
	if !plan.Comment.IsNull() {
		delegationIn.Comment = aws.String(plan.Comment.ValueString())
	}
	in := auditmanager.BatchCreateDelegationByAssessmentInput{
		AssessmentId:             aws.String(plan.AssessmentID.ValueString()),
		CreateDelegationRequests: []awstypes.CreateDelegationRequest{delegationIn},
	}

	// Include retry handling to allow for IAM propagation
	//
	// Example:
	//   ResourceNotFoundException: The operation tried to access a nonexistent resource. The resource
	//   might not be specified correctly, or its status might not be active. Check and try again.
	var out *auditmanager.BatchCreateDelegationByAssessmentOutput
	err := tfresource.Retry(ctx, iamPropagationTimeout, func() *retry.RetryError {
		var err error
		out, err = conn.BatchCreateDelegationByAssessment(ctx, &in)
		if err != nil {
			var nfe *awstypes.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAssessmentDelegation, plan.RoleARN.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil || len(out.Delegations) == 0 {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAssessmentDelegation, plan.RoleARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// This response object will return ALL delegations assigned to the assessment, not just those
	// added in this batch request. In order to write to state, the response should be filtered to
	// the item with a matching role_arn and control_set_id.
	//
	// Also, assessment_id is returned as null in the BatchCreateDelegationByAssessment response
	// object, and therefore is not included as one of the matching parameters.
	delegation, err := getMatchingDelegation(out.Delegations, plan.RoleARN.ValueString(), plan.ControlSetID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAssessmentDelegation, plan.RoleARN.String(), nil),
			err.Error(),
		)
		return
	}

	state := plan

	// The AWS-generated ID for delegations has been observed to change between creation
	// and subseqeunt read operations. As such, the ID attribute will use a combination of
	// attributes that are unique to a single delegation instead.
	id := toID(plan.AssessmentID.ValueString(), plan.RoleARN.ValueString(), plan.ControlSetID.ValueString())
	state.ID = types.StringValue(id)

	state.refreshFromOutput(ctx, delegation)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceAssessmentDelegation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceAssessmentDelegationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindAssessmentDelegationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameAssessmentDelegation, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.refreshFromOutputMetadata(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceAssessmentDelegation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceAssessmentDelegation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceAssessmentDelegationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.BatchDeleteDelegationByAssessment(ctx, &auditmanager.BatchDeleteDelegationByAssessmentInput{
		AssessmentId:  aws.String(state.AssessmentID.ValueString()),
		DelegationIds: []string{state.DelegationID.ValueString()},
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameAssessmentDelegation, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceAssessmentDelegation) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindAssessmentDelegationByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.DelegationMetadata, error) {
	assessmentID, roleARN, controlSetID := fromID(id)

	// The GetDelegations API behaves like a List* API, so the results are paged
	// through until an entry with a matching ID is found
	in := &auditmanager.GetDelegationsInput{}
	pages := auditmanager.NewGetDelegationsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, d := range page.Delegations {
			if aws.ToString(d.AssessmentId) == assessmentID &&
				strings.EqualFold(aws.ToString(d.RoleArn), roleARN) && // IAM role names are case-insensitive
				aws.ToString(d.ControlSetName) == controlSetID {
				return &d, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastRequest: in,
	}
}

// getMatchingDelegation will return the delegation matching the provided role ARN and
// control set ID. If no match is found, an error is returned.
func getMatchingDelegation(out []awstypes.Delegation, roleARN, controlSetID string) (*awstypes.Delegation, error) {
	for _, d := range out {
		if strings.EqualFold(aws.ToString(d.RoleArn), roleARN) && // IAM role names are case-insensitive
			aws.ToString(d.ControlSetId) == controlSetID {
			return &d, nil
		}
	}
	return nil, errors.New("no matching delegations in response")
}

func fromID(id string) (string, string, string) {
	parts := strings.Split(id, ",")
	if len(parts) != 3 {
		return "", "", ""
	}
	return parts[0], parts[1], parts[2]
}

func toID(assessmentID, roleARN, controlSetID string) string {
	return strings.Join([]string{assessmentID, roleARN, controlSetID}, ",")
}

type resourceAssessmentDelegationData struct {
	AssessmentID types.String `tfsdk:"assessment_id"`
	Comment      types.String `tfsdk:"comment"`
	ControlSetID types.String `tfsdk:"control_set_id"`
	DelegationID types.String `tfsdk:"delegation_id"`
	ID           types.String `tfsdk:"id"`
	RoleARN      types.String `tfsdk:"role_arn"`
	RoleType     types.String `tfsdk:"role_type"`
	Status       types.String `tfsdk:"status"`
}

// refreshFromOutput writes state data from an AWS response object
//
// This variant of the refresh method is for use with the create operation
// response type (Delegation).
func (rd *resourceAssessmentDelegationData) refreshFromOutput(ctx context.Context, out *awstypes.Delegation) {
	if out == nil {
		return
	}

	// The response from create operations always includes a nil AssessmentId. This is likely
	// a bug in the AWS API, so for now skip using the response output and copy the state
	// value directly from plan.
	// rd.AssessmentID = flex.StringToFramework(ctx, out.AssessmentId)

	rd.Comment = flex.StringToFramework(ctx, out.Comment)
	rd.ControlSetID = flex.StringToFramework(ctx, out.ControlSetId)
	rd.DelegationID = flex.StringToFramework(ctx, out.Id)
	rd.RoleARN = flex.StringToFramework(ctx, out.RoleArn)
	rd.RoleType = flex.StringValueToFramework(ctx, out.RoleType)
	rd.Status = flex.StringValueToFramework(ctx, out.Status)
}

// refreshFromOutputMetadata writes state data from an AWS response object
//
// This variant of the refresh method is for use with the get operation
// response type (DelegationMetadata). Notably, this response omits certain
// attributes such as comment, control_set_id, and role_type which means
// drift cannot be detected after the initial create action.
func (rd *resourceAssessmentDelegationData) refreshFromOutputMetadata(ctx context.Context, out *awstypes.DelegationMetadata) {
	if out == nil {
		return
	}

	rd.AssessmentID = flex.StringToFramework(ctx, out.AssessmentId)
	rd.DelegationID = flex.StringToFramework(ctx, out.Id)
	rd.RoleARN = flex.StringToFramework(ctx, out.RoleArn)
	rd.Status = flex.StringValueToFramework(ctx, out.Status)
}
