// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_auditmanager_assessment", &resource.Sweeper{
		Name: "aws_auditmanager_assessment",
		F:    sweepAssessments,
		Dependencies: []string{
			"aws_auditmanager_control",
			"aws_auditmanager_framework",
			"aws_iam_role",
			"aws_s3_bucket",
		},
	})
	resource.AddTestSweepers("aws_auditmanager_assessment_delegation", &resource.Sweeper{
		Name: "aws_auditmanager_assessment_delegation",
		F:    sweepAssessmentDelegations,
	})
	resource.AddTestSweepers("aws_auditmanager_assessment_report", &resource.Sweeper{
		Name: "aws_auditmanager_assessment_report",
		F:    sweepAssessmentReports,
	})
	resource.AddTestSweepers("aws_auditmanager_control", &resource.Sweeper{
		Name: "aws_auditmanager_control",
		F:    sweepControls,
	})
	resource.AddTestSweepers("aws_auditmanager_framework", &resource.Sweeper{
		Name: "aws_auditmanager_framework",
		F:    sweepFrameworks,
	})
	resource.AddTestSweepers("aws_auditmanager_framework_share", &resource.Sweeper{
		Name: "aws_auditmanager_framework_share",
		F:    sweepFrameworkShares,
	})
}

// isCompleteSetupError checks whether the returned error message indicates
// AuditManager isn't yet enabled in the current region.
//
// For example:
// AccessDeniedException: Please complete AWS Audit Manager setup from home page to enable this action in this account.
func isCompleteSetupError(err error) bool {
	var ade *types.AccessDeniedException
	return errors.As(err, &ade)
}

func sweepAssessments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AuditManagerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentsInput{}

	pages := auditmanager.NewListAssessmentsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Assessments sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Assessments: %w", err)
		}

		for _, assessment := range page.AssessmentMetadata {
			id := aws.ToString(assessment.Id)

			log.Printf("[INFO] Deleting AuditManager Assessment: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceAssessment, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping AuditManager Assessments for %s: %w", region, err)
	}

	return nil
}

func sweepAssessmentDelegations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AuditManagerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.GetDelegationsInput{}

	pages := auditmanager.NewGetDelegationsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Assesment Delegations sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Assessment Delegations: %w", err)
		}

		for _, d := range page.Delegations {
			log.Printf("[INFO] Deleting AuditManager Assessment Delegation: %s", aws.ToString(d.Id))
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceAssessmentDelegation, client,
				framework.NewAttribute("assessment_id", aws.ToString(d.AssessmentId)),
				framework.NewAttribute("delegation_id", aws.ToString(d.Id)),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping AuditManager Assessment Delegations for %s: %w", region, err)
	}

	return nil
}

func sweepAssessmentReports(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AuditManagerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentReportsInput{}

	pages := auditmanager.NewListAssessmentReportsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Assesment Reports sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Assessment Reports: %w", err)
		}

		for _, report := range page.AssessmentReports {
			id := aws.ToString(report.Id)

			log.Printf("[INFO] Deleting AuditManager Assessment Report: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceAssessmentReport, client,
				framework.NewAttribute(names.AttrID, id),
				framework.NewAttribute("assessment_id", aws.ToString(report.AssessmentId)),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping AuditManager Assessment Reports for %s: %w", region, err)
	}

	return nil
}

func sweepControls(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AuditManagerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListControlsInput{ControlType: types.ControlTypeCustom}

	pages := auditmanager.NewListControlsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Controls sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Controls: %w", err)
		}

		for _, control := range page.ControlMetadataList {
			id := aws.ToString(control.Id)

			log.Printf("[INFO] Deleting AuditManager Control: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceControl, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping AuditManager Controls for %s: %w", region, err)
	}

	return nil
}

func sweepFrameworks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AuditManagerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentFrameworksInput{FrameworkType: types.FrameworkTypeCustom}

	pages := auditmanager.NewListAssessmentFrameworksPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Frameworks sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Frameworks: %w", err)
		}

		for _, f := range page.FrameworkMetadataList {
			id := aws.ToString(f.Id)

			log.Printf("[INFO] Deleting AuditManager Framework: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceFramework, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping AuditManager Frameworks for %s: %w", region, err)
	}

	return nil
}

func sweepFrameworkShares(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AuditManagerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentFrameworkShareRequestsInput{RequestType: types.ShareRequestTypeSent}

	pages := auditmanager.NewListAssessmentFrameworkShareRequestsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Framework Shares sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Framework Shares: %w", err)
		}

		for _, share := range page.AssessmentFrameworkShareRequests {
			id := aws.ToString(share.Id)

			log.Printf("[INFO] Deleting AuditManager Framework Share: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceFrameworkShare, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping AuditManager Framework Shares for %s: %w", region, err)
	}

	return nil
}
