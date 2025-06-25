// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_auditmanager_assessment", sweepAssessments, "aws_auditmanager_assessment_delegation", "aws_auditmanager_assessment_report")
	awsv2.Register("aws_auditmanager_assessment_delegation", sweepAssessmentDelegations)
	awsv2.Register("aws_auditmanager_assessment_report", sweepAssessmentReports)
	awsv2.Register("aws_auditmanager_control", sweepControls, "aws_auditmanager_framework")
	awsv2.Register("aws_auditmanager_framework", sweepFrameworks, "aws_auditmanager_assessment", "aws_auditmanager_framework_share")
	awsv2.Register("aws_auditmanager_framework_share", sweepFrameworkShares)
}

func sweepAssessments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AuditManagerClient(ctx)
	var input auditmanager.ListAssessmentsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := auditmanager.NewListAssessmentsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.AccessDeniedException](err) {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssessmentMetadata {
			id := aws.ToString(v.Id)

			log.Printf("[INFO] Deleting AuditManager Assessment: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newAssessmentResource, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	return sweepResources, nil
}

func sweepAssessmentDelegations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AuditManagerClient(ctx)
	var input auditmanager.GetDelegationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := auditmanager.NewGetDelegationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.AccessDeniedException](err) {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Delegations {
			id := aws.ToString(v.Id)

			log.Printf("[INFO] Deleting AuditManager Assessment Delegation: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newAssessmentDelegationResource, client,
				framework.NewAttribute("assessment_id", aws.ToString(v.AssessmentId)),
				framework.NewAttribute("delegation_id", id),
			))
		}
	}

	return sweepResources, nil
}

func sweepAssessmentReports(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AuditManagerClient(ctx)
	var input auditmanager.ListAssessmentReportsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := auditmanager.NewListAssessmentReportsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.AccessDeniedException](err) {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssessmentReports {
			id := aws.ToString(v.Id)

			log.Printf("[INFO] Deleting AuditManager Assessment Report: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newAssessmentReportResource, client,
				framework.NewAttribute(names.AttrID, id),
				framework.NewAttribute("assessment_id", aws.ToString(v.AssessmentId)),
			))
		}
	}

	return sweepResources, nil
}

func sweepControls(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AuditManagerClient(ctx)
	var input auditmanager.ListControlsInput
	input.ControlType = types.ControlTypeCustom
	sweepResources := make([]sweep.Sweepable, 0)

	pages := auditmanager.NewListControlsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.AccessDeniedException](err) {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ControlMetadataList {
			id := aws.ToString(v.Id)

			log.Printf("[INFO] Deleting AuditManager Control: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newControlResource, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	return sweepResources, nil
}

func sweepFrameworks(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AuditManagerClient(ctx)
	var input auditmanager.ListAssessmentFrameworksInput
	input.FrameworkType = types.FrameworkTypeCustom
	sweepResources := make([]sweep.Sweepable, 0)

	pages := auditmanager.NewListAssessmentFrameworksPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.AccessDeniedException](err) {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.FrameworkMetadataList {
			id := aws.ToString(v.Id)

			log.Printf("[INFO] Deleting AuditManager Framework: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newFrameworkResource, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	return sweepResources, nil
}

func sweepFrameworkShares(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AuditManagerClient(ctx)
	var input auditmanager.ListAssessmentFrameworkShareRequestsInput
	input.RequestType = types.ShareRequestTypeSent
	sweepResources := make([]sweep.Sweepable, 0)

	pages := auditmanager.NewListAssessmentFrameworkShareRequestsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.AccessDeniedException](err) {
			break
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AssessmentFrameworkShareRequests {
			id := aws.ToString(v.Id)

			log.Printf("[INFO] Deleting AuditManager Framework Share: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newFrameworkShareResource, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	return sweepResources, nil
}
