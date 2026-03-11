// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package auditmanager

import (
	"context"
	"fmt"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_auditmanager_assessment_report", name="Assessment Report")
func newAssessmentReportResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &assessmentReportResource{}, nil
}

type assessmentReportResource struct {
	framework.ResourceWithModel[assessmentReportResourceModel]
	framework.WithNoUpdate
	framework.WithImportByID
}

func (r *assessmentReportResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"assessment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"author": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *assessmentReportResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data assessmentReportResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input auditmanager.CreateAssessmentReportInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateAssessmentReport(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Audit Manager Assessment Report (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	assessmentReport := output.AssessmentReport
	data.Author = fwflex.StringToFramework(ctx, assessmentReport.Author)
	data.ID = fwflex.StringToFramework(ctx, assessmentReport.Id)
	data.Status = fwflex.StringValueToFramework(ctx, assessmentReport.Status)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *assessmentReportResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data assessmentReportResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	output, err := findAssessmentReportByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Audit Manager Assessment Report (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *assessmentReportResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data assessmentReportResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AuditManagerClient(ctx)

	// Retry on ValidationException (report status is not completed or failed)
	//
	// Example:
	//   ValidationException: The assessment report is currently being generated and can’t be
	//   deleted. You can only delete assessment reports that are completed or failed
	input := auditmanager.DeleteAssessmentReportInput{
		AssessmentId:       fwflex.StringFromFramework(ctx, data.AssessmentID),
		AssessmentReportId: fwflex.StringFromFramework(ctx, data.ID),
	}
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ValidationException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteAssessmentReport(ctx, &input)
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Audit Manager Assessment Report (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAssessmentReportByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.AssessmentReportMetadata, error) {
	var input auditmanager.ListAssessmentReportsInput

	pages := auditmanager.NewListAssessmentReportsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssessmentReports {
			if id == aws.ToString(v.Id) {
				return &v, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

type assessmentReportResourceModel struct {
	framework.WithRegionModel
	AssessmentID types.String `tfsdk:"assessment_id"`
	Author       types.String `tfsdk:"author"`
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Status       types.String `tfsdk:"status"`
}
