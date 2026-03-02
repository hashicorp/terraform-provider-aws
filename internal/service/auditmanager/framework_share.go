// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package auditmanager

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_auditmanager_framework_share", name="Framework Share")
func newFrameworkShareResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &frameworkShareResource{}, nil
}

type frameworkShareResource struct {
	framework.ResourceWithModel[frameworkShareResourceModel]
	framework.WithImportByID
	framework.WithNoUpdate
}

func (r *frameworkShareResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrComment: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_account": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"destination_region": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"framework_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ShareRequestStatus](),
				Computed:   true,
			},
		},
	}
}

func (r *frameworkShareResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data frameworkShareResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	var input auditmanager.StartAssessmentFrameworkShareInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.StartAssessmentFrameworkShare(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Audit Manager Framework Share", err.Error())

		return
	}

	// Set values for unknowns.
	assessmentFrameworkShareRequest := output.AssessmentFrameworkShareRequest
	data.ID = fwflex.StringToFramework(ctx, assessmentFrameworkShareRequest.Id)
	fwtypes.StringEnumValue(assessmentFrameworkShareRequest.Status)
	data.Status = fwtypes.StringEnumValue(assessmentFrameworkShareRequest.Status)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *frameworkShareResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data frameworkShareResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	output, err := findFrameworkShareByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Framework Share (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *frameworkShareResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data frameworkShareResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	// Framework share requests in certain statuses must be revoked before deletion.
	id := fwflex.StringValueFromFramework(ctx, data.ID)
	output, err := findFrameworkShareByID(ctx, conn, id)

	if retry.NotFound(err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Framework Share (%s)", id), err.Error())

		return
	}

	nonRevokable := []awstypes.ShareRequestStatus{
		awstypes.ShareRequestStatusDeclined,
		awstypes.ShareRequestStatusExpired,
		awstypes.ShareRequestStatusFailed,
		awstypes.ShareRequestStatusRevoked,
	}
	if !slices.Contains(nonRevokable, output.Status) {
		input := auditmanager.UpdateAssessmentFrameworkShareInput{
			RequestId:   aws.String(id),
			RequestType: awstypes.ShareRequestTypeSent,
			Action:      awstypes.ShareRequestActionRevoke,
		}
		_, err := conn.UpdateAssessmentFrameworkShare(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("revoking Audit Manager Framework Share (%s)", id), err.Error())

			return
		}
	}

	input := auditmanager.DeleteAssessmentFrameworkShareInput{
		RequestId:   aws.String(id),
		RequestType: awstypes.ShareRequestTypeSent,
	}
	_, err = conn.DeleteAssessmentFrameworkShare(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Audit Manager Framework Share (%s)", id), err.Error())

		return
	}
}

func findFrameworkShareByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.AssessmentFrameworkShareRequest, error) {
	input := auditmanager.ListAssessmentFrameworkShareRequestsInput{
		RequestType: awstypes.ShareRequestTypeSent,
	}

	pages := auditmanager.NewListAssessmentFrameworkShareRequestsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssessmentFrameworkShareRequests {
			if id == aws.ToString(v.Id) {
				return &v, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

type frameworkShareResourceModel struct {
	framework.WithRegionModel
	Comment            types.String                                    `tfsdk:"comment"`
	DestinationAccount types.String                                    `tfsdk:"destination_account"`
	DestinationRegion  types.String                                    `tfsdk:"destination_region"`
	FrameworkID        types.String                                    `tfsdk:"framework_id"`
	ID                 types.String                                    `tfsdk:"id"`
	Status             fwtypes.StringEnum[awstypes.ShareRequestStatus] `tfsdk:"status"`
}
