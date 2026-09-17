// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package auditmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_auditmanager_assessment_delegation", name="Assessment Delegation")
func newAssessmentDelegationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &assessmentDelegationResource{}, nil
}

const (
	ResNameAssessmentDelegation = "AssessmentDelegation"
)

type assessmentDelegationResource struct {
	framework.ResourceWithModel[assessmentDelegationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *assessmentDelegationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			// and subsequent read operations. As such, this value cannot be used as the resource ID
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
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RoleType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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

const (
	assessmentDelegationResourceIDPartCount = 3
)

func (r *assessmentDelegationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data assessmentDelegationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	var createDelegationRequest awstypes.CreateDelegationRequest
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &createDelegationRequest)...)
	if response.Diagnostics.HasError() {
		return
	}

	assessmentID, roleARN, controlSetID := data.AssessmentID.ValueString(), data.RoleARN.ValueString(), data.ControlSetID.ValueString()
	id, _ := intflex.FlattenResourceId([]string{assessmentID, roleARN, controlSetID}, assessmentDelegationResourceIDPartCount, false)
	input := auditmanager.BatchCreateDelegationByAssessmentInput{
		AssessmentId:             aws.String(assessmentID),
		CreateDelegationRequests: []awstypes.CreateDelegationRequest{createDelegationRequest},
	}

	// Include retry handling to allow for IAM propagation
	//
	// Example:
	//   ResourceNotFoundException: The operation tried to access a nonexistent resource. The resource
	//   might not be specified correctly, or its status might not be active. Check and try again.
	outputRaw, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceNotFoundException](ctx, iamPropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.BatchCreateDelegationByAssessment(ctx, &input)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Audit Manager Assessment Delegation (%s)", id), err.Error())

		return
	}

	// This response object will return ALL delegations assigned to the assessment, not just those
	// added in this batch request. In order to write to state, the response should be filtered to
	// the item with a matching role_arn and control_set_id.
	//
	// Also, assessment_id is returned as null in the BatchCreateDelegationByAssessment response
	// object, and therefore is not included as one of the matching parameters.
	output, err := tfresource.AssertSingleValueResult(tfslices.Filter(outputRaw.(*auditmanager.BatchCreateDelegationByAssessmentOutput).Delegations, func(v awstypes.Delegation) bool {
		// IAM role names are case-insensitive.
		return strings.EqualFold(aws.ToString(v.RoleArn), roleARN) && aws.ToString(v.ControlSetId) == controlSetID
	}))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Audit Manager Assessment Delegation (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// The response from create operations always includes a nil AssessmentId. This is likely
	// a bug in the AWS API, so for now skip using the response output and copy the state
	// value directly from plan.
	data.AssessmentID = fwflex.StringValueToFramework(ctx, assessmentID)
	data.DelegationID = data.ID
	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *assessmentDelegationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data assessmentDelegationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	parts, err := intflex.ExpandResourceId(id, assessmentDelegationResourceIDPartCount, false)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	assessmentID, roleARN, controlSetID := parts[0], parts[1], parts[2]
	output, err := findAssessmentDelegationByThreePartKey(ctx, conn, assessmentID, roleARN, controlSetID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Assessment Delegation (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ControlSetID = fwflex.StringValueToFramework(ctx, controlSetID)
	data.DelegationID = fwflex.StringToFramework(ctx, output.Id)
	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *assessmentDelegationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data assessmentDelegationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	input := auditmanager.BatchDeleteDelegationByAssessmentInput{
		AssessmentId:  fwflex.StringFromFramework(ctx, data.AssessmentID),
		DelegationIds: []string{fwflex.StringValueFromFramework(ctx, data.DelegationID)},
	}
	_, err := conn.BatchDeleteDelegationByAssessment(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Audit Manager Assessment Delegation (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAssessmentDelegationByThreePartKey(ctx context.Context, conn *auditmanager.Client, assessmentID, roleARN, controlSetID string) (*awstypes.DelegationMetadata, error) {
	var input auditmanager.GetDelegationsInput

	pages := auditmanager.NewGetDelegationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Delegations {
			if aws.ToString(v.AssessmentId) == assessmentID &&
				strings.EqualFold(aws.ToString(v.RoleArn), roleARN) && // IAM role names are case-insensitive.
				aws.ToString(v.ControlSetName) == controlSetID {
				return &v, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

type assessmentDelegationResourceModel struct {
	framework.WithRegionModel
	AssessmentID types.String                          `tfsdk:"assessment_id"`
	Comment      types.String                          `tfsdk:"comment"`
	ControlSetID types.String                          `tfsdk:"control_set_id"`
	DelegationID types.String                          `tfsdk:"delegation_id"`
	ID           types.String                          `tfsdk:"id"`
	RoleARN      fwtypes.ARN                           `tfsdk:"role_arn"`
	RoleType     fwtypes.StringEnum[awstypes.RoleType] `tfsdk:"role_type"`
	Status       types.String                          `tfsdk:"status"`
}
