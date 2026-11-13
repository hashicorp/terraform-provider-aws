// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_continuum_code_scans", name="Code Scans")
func newCodeScansDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &codeScansDataSource{}, nil
}

type codeScansDataSource struct {
	framework.DataSourceWithModel[codeScansDataSourceModel]
}

func (d *codeScansDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{Required: true},
			names.AttrID:     framework.IDAttribute(),
			"scans":          framework.DataSourceComputedListOfObjectAttribute[codeScanModel](ctx),
		},
	}
}

func (d *codeScansDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data codeScansDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ContinuumClient(ctx)

	environmentID := data.EnvironmentID.ValueString()
	var scans []awstypes.CodeScanSummary
	pages := continuum.NewListCodeScansPaginator(conn, &continuum.ListCodeScansInput{EnvironmentId: aws.String(environmentID)})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, environmentID)
			return
		}
		scans = append(scans, page.Scans...)
	}

	data.ID = types.StringValue(environmentID)
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, scans, &data.Scans))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type codeScansDataSourceModel struct {
	framework.WithRegionModel
	EnvironmentID types.String                                   `tfsdk:"environment_id"`
	ID            types.String                                   `tfsdk:"id"`
	Scans         fwtypes.ListNestedObjectValueOf[codeScanModel] `tfsdk:"scans"`
}

type codeScanModel struct {
	CompletedAt    timetypes.RFC3339                             `tfsdk:"completed_at"`
	CreatedAt      timetypes.RFC3339                             `tfsdk:"created_at"`
	FailureMessage types.String                                  `tfsdk:"failure_message"`
	FailureStage   fwtypes.StringEnum[awstypes.ScanFailureStage] `tfsdk:"failure_stage"`
	Origin         fwtypes.StringEnum[awstypes.ScanOrigin]       `tfsdk:"origin"`
	ScanRunID      types.String                                  `tfsdk:"scan_run_id"`
	Status         fwtypes.StringEnum[awstypes.ScanStatus]       `tfsdk:"status"`
}

// @FrameworkDataSource("aws_continuum_code_issues", name="Code Issues")
func newCodeIssuesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &codeIssuesDataSource{}, nil
}

type codeIssuesDataSource struct {
	framework.DataSourceWithModel[codeIssuesDataSourceModel]
}

func (d *codeIssuesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{Required: true},
			names.AttrID:     framework.IDAttribute(),
			"issues":         framework.DataSourceComputedListOfObjectAttribute[codeIssueModel](ctx),
		},
	}
}

func (d *codeIssuesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data codeIssuesDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ContinuumClient(ctx)

	environmentID := data.EnvironmentID.ValueString()
	var issues []awstypes.IssueSummary
	pages := continuum.NewListCodeIssuesPaginator(conn, &continuum.ListCodeIssuesInput{EnvironmentId: aws.String(environmentID)})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, environmentID)
			return
		}
		issues = append(issues, page.Items...)
	}

	data.ID = types.StringValue(environmentID)
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, issues, &data.Issues))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type codeIssuesDataSourceModel struct {
	framework.WithRegionModel
	EnvironmentID types.String                                    `tfsdk:"environment_id"`
	ID            types.String                                    `tfsdk:"id"`
	Issues        fwtypes.ListNestedObjectValueOf[codeIssueModel] `tfsdk:"issues"`
}

type codeIssueModel struct {
	Confidence      types.Int32                                    `tfsdk:"confidence"`
	CreatedAt       timetypes.RFC3339                              `tfsdk:"created_at"`
	IssueARN        types.String                                   `tfsdk:"issue_arn"`
	IssueID         types.String                                   `tfsdk:"issue_id"`
	Name            types.String                                   `tfsdk:"name"`
	Phase           fwtypes.StringEnum[awstypes.IssuePhase]        `tfsdk:"phase"`
	Severity        fwtypes.StringEnum[awstypes.IssueSeverity]     `tfsdk:"severity"`
	Status          fwtypes.StringEnum[awstypes.IssueStatus]       `tfsdk:"status"`
	StatusMessage   types.String                                   `tfsdk:"status_message"`
	StatusReason    fwtypes.StringEnum[awstypes.IssueStatusReason] `tfsdk:"status_reason"`
	StatusUpdatedAt timetypes.RFC3339                              `tfsdk:"status_updated_at"`
	UpdatedAt       timetypes.RFC3339                              `tfsdk:"updated_at"`
}

// @FrameworkDataSource("aws_continuum_code_findings", name="Code Findings")
func newCodeFindingsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &codeFindingsDataSource{}, nil
}

type codeFindingsDataSource struct {
	framework.DataSourceWithModel[codeFindingsDataSourceModel]
}

func (d *codeFindingsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{Required: true},
			"findings":       framework.DataSourceComputedListOfObjectAttribute[codeFindingModel](ctx),
			names.AttrID:     framework.IDAttribute(),
		},
	}
}

func (d *codeFindingsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data codeFindingsDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ContinuumClient(ctx)

	environmentID := data.EnvironmentID.ValueString()
	var findings []awstypes.FindingSummary
	pages := continuum.NewListCodeFindingsPaginator(conn, &continuum.ListCodeFindingsInput{EnvironmentId: aws.String(environmentID)})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, environmentID)
			return
		}
		findings = append(findings, page.Items...)
	}

	data.ID = types.StringValue(environmentID)
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, findings, &data.Findings))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type codeFindingsDataSourceModel struct {
	framework.WithRegionModel
	EnvironmentID types.String                                      `tfsdk:"environment_id"`
	Findings      fwtypes.ListNestedObjectValueOf[codeFindingModel] `tfsdk:"findings"`
	ID            types.String                                      `tfsdk:"id"`
}

type codeFindingModel struct {
	Category         types.String                                      `tfsdk:"category"`
	CompletedAt      timetypes.RFC3339                                 `tfsdk:"completed_at"`
	CreatedAt        timetypes.RFC3339                                 `tfsdk:"created_at"`
	CreatedIssueID   types.String                                      `tfsdk:"created_issue_id"`
	DiscardReason    fwtypes.StringEnum[awstypes.FindingDiscardReason] `tfsdk:"discard_reason"`
	DiscardReasoning types.String                                      `tfsdk:"discard_reasoning"`
	FilePath         types.String                                      `tfsdk:"file_path"`
	FindingARN       types.String                                      `tfsdk:"finding_arn"`
	FindingID        types.String                                      `tfsdk:"finding_id"`
	ScanRunID        types.String                                      `tfsdk:"scan_run_id"`
	Severity         fwtypes.StringEnum[awstypes.FindingSeverityId]    `tfsdk:"severity"`
	Source           types.String                                      `tfsdk:"source"`
	Status           fwtypes.StringEnum[awstypes.FindingStatus]        `tfsdk:"status"`
	StatusMessage    types.String                                      `tfsdk:"status_message"`
	Title            types.String                                      `tfsdk:"title"`
	Type             fwtypes.StringEnum[awstypes.FindingType]          `tfsdk:"type"`
}
