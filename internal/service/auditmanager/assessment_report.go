// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const reportCompletionTimeout = 5 * time.Minute

// @FrameworkResource
func newResourceAssessmentReport(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceAssessmentReport{}, nil
}

const (
	ResNameAssessmentReport = "AssessmentReport"
)

type resourceAssessmentReport struct {
	framework.ResourceWithConfigure
}

func (r *resourceAssessmentReport) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_auditmanager_assessment_report"
}

func (r *resourceAssessmentReport) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
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

func (r *resourceAssessmentReport) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var plan resourceAssessmentReportData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := auditmanager.CreateAssessmentReportInput{
		AssessmentId: aws.String(plan.AssessmentID.ValueString()),
		Name:         aws.String(plan.Name.ValueString()),
	}
	if !plan.Description.IsNull() {
		in.Description = aws.String(plan.Description.ValueString())
	}

	out, err := conn.CreateAssessmentReport(ctx, &in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAssessmentReport, plan.Name.String(), nil),
			err.Error(),
		)
		return
	}
	if out == nil || out.AssessmentReport == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionCreating, ResNameAssessmentReport, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	state := plan
	state.refreshFromOutput(ctx, out.AssessmentReport)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourceAssessmentReport) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceAssessmentReportData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := FindAssessmentReportByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.AddWarning(
			"AWS Resource Not Found During Refresh",
			fmt.Sprintf("Automatically removing from Terraform State instead of returning the error, which may trigger resource recreation. Original Error: %s", err.Error()),
		)
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionReading, ResNameAssessmentReport, state.ID.String(), nil),
			err.Error(),
		)
		return
	}

	state.refreshFromOutputMetadata(ctx, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// There is no update API, so this method is a no-op
func (r *resourceAssessmentReport) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceAssessmentReport) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().AuditManagerClient(ctx)

	var state resourceAssessmentReportData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retry on ValidationException (report status is not completed or failed)
	//
	// Example:
	//   ValidationException: The assessment report is currently being generated and can’t be
	//   deleted. You can only delete assessment reports that are completed or failed
	err := tfresource.Retry(ctx, reportCompletionTimeout, func() *retry.RetryError {
		_, err := conn.DeleteAssessmentReport(ctx, &auditmanager.DeleteAssessmentReportInput{
			AssessmentId:       aws.String(state.AssessmentID.ValueString()),
			AssessmentReportId: aws.String(state.ID.ValueString()),
		})
		if err != nil {
			var ve *awstypes.ValidationException
			if errors.As(err, &ve) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AuditManager, create.ErrActionDeleting, ResNameAssessmentReport, state.ID.String(), nil),
			err.Error(),
		)
	}
}

func (r *resourceAssessmentReport) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func FindAssessmentReportByID(ctx context.Context, conn *auditmanager.Client, id string) (*awstypes.AssessmentReportMetadata, error) {
	// There is no GetAssessmentReport API, so make use of the ListAssessmentReports API
	// and return when an ID match is found
	in := &auditmanager.ListAssessmentReportsInput{}
	pages := auditmanager.NewListAssessmentReportsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, report := range page.AssessmentReports {
			if id == aws.ToString(report.Id) {
				return &report, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastRequest: in,
	}
}

type resourceAssessmentReportData struct {
	AssessmentID types.String `tfsdk:"assessment_id"`
	Author       types.String `tfsdk:"author"`
	Description  types.String `tfsdk:"description"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Status       types.String `tfsdk:"status"`
}

// refreshFromOutput writes state data from an AWS response object
//
// This variant of the refresh method is for use with the create operation
// response type (AssesmentReport).
func (rd *resourceAssessmentReportData) refreshFromOutput(ctx context.Context, out *awstypes.AssessmentReport) {
	if out == nil {
		return
	}

	rd.AssessmentID = flex.StringToFramework(ctx, out.AssessmentId)
	rd.Author = flex.StringToFramework(ctx, out.Author)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.ID = flex.StringToFramework(ctx, out.Id)
	rd.Name = flex.StringToFramework(ctx, out.Name)
	rd.Status = flex.StringValueToFramework(ctx, out.Status)
}

// refreshFromOutputMetadata writes state data from an AWS response object
//
// This variant of the refresh method is for use with the list operation
// response type (AssesmentReportMetadata).
func (rd *resourceAssessmentReportData) refreshFromOutputMetadata(ctx context.Context, out *awstypes.AssessmentReportMetadata) {
	if out == nil {
		return
	}

	rd.AssessmentID = flex.StringToFramework(ctx, out.AssessmentId)
	rd.Author = flex.StringToFramework(ctx, out.Author)
	rd.Description = flex.StringToFramework(ctx, out.Description)
	rd.ID = flex.StringToFramework(ctx, out.Id)
	rd.Name = flex.StringToFramework(ctx, out.Name)
	rd.Status = flex.StringValueToFramework(ctx, out.Status)
}
