//go:build sweep
// +build sweep

package auditmanager

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager"
	"github.com/aws/aws-sdk-go-v2/service/auditmanager/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	if errors.As(err, &ade) {
		return true
	}
	return false
}

func sweepAssessments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AuditManagerClient()
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentsInput{}
	var errs *multierror.Error

	pages := auditmanager.NewListAssessmentsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Assessments sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Assessments: %w", err)
		}

		for _, assessment := range page.AssessmentMetadata {
			id := aws.ToString(assessment.Id)

			log.Printf("[INFO] Deleting AuditManager Assessment: %s", id)
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceAssessment, id, client))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Assessments for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Assessments sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepAssessmentDelegations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AuditManagerClient()
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.GetDelegationsInput{}
	var errs *multierror.Error

	pages := auditmanager.NewGetDelegationsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Assesment Delegations sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Assessment Delegations: %w", err)
		}

		for _, d := range page.Delegations {
			id := "" // ID is a combination of attributes for this resource, but not used for deletion

			// assessment ID is required for delete operations
			assessmentIDAttr := sweep.FrameworkSupplementalAttribute{
				Path:  "assessment_id",
				Value: aws.ToString(d.AssessmentId),
			}
			// delegation ID is required for delete operations
			delegationIDAttr := sweep.FrameworkSupplementalAttribute{
				Path:  "delegation_id",
				Value: aws.ToString(d.Id),
			}

			log.Printf("[INFO] Deleting AuditManager Assessment Delegation: %s", delegationIDAttr.Value)
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceAssessmentDelegation, id, client, assessmentIDAttr, delegationIDAttr))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Assessment Delegations for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Assessment Delegations sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepAssessmentReports(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AuditManagerClient()
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentReportsInput{}
	var errs *multierror.Error

	pages := auditmanager.NewListAssessmentReportsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Assesment Reports sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Assessment Reports: %w", err)
		}

		for _, report := range page.AssessmentReports {
			id := aws.ToString(report.Id)
			// assessment ID is required for delete operations
			assessmentIDAttr := sweep.FrameworkSupplementalAttribute{
				Path:  "assessment_id",
				Value: aws.ToString(report.AssessmentId),
			}

			log.Printf("[INFO] Deleting AuditManager Assessment Report: %s", id)
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceAssessmentReport, id, client, assessmentIDAttr))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Assessment Reports for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Assessment Reports sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepControls(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AuditManagerClient()
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListControlsInput{ControlType: types.ControlTypeCustom}
	var errs *multierror.Error

	pages := auditmanager.NewListControlsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Controls sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Controls: %w", err)
		}

		for _, control := range page.ControlMetadataList {
			id := aws.ToString(control.Id)

			log.Printf("[INFO] Deleting AuditManager Control: %s", id)
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceControl, id, client))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Controls for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Controls sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepFrameworks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AuditManagerClient()
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentFrameworksInput{FrameworkType: types.FrameworkTypeCustom}
	var errs *multierror.Error

	pages := auditmanager.NewListAssessmentFrameworksPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Frameworks sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Frameworks: %w", err)
		}

		for _, framework := range page.FrameworkMetadataList {
			id := aws.ToString(framework.Id)

			log.Printf("[INFO] Deleting AuditManager Framework: %s", id)
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceFramework, id, client))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Frameworks for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Frameworks sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepFrameworkShares(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AuditManagerClient()
	sweepResources := make([]sweep.Sweepable, 0)
	in := &auditmanager.ListAssessmentFrameworkShareRequestsInput{RequestType: types.ShareRequestTypeSent}
	var errs *multierror.Error

	pages := auditmanager.NewListAssessmentFrameworkShareRequestsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) || isCompleteSetupError(err) {
			log.Printf("[WARN] Skipping AuditManager Framework Shares sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving AuditManager Framework Shares: %w", err)
		}

		for _, share := range page.AssessmentFrameworkShareRequests {
			id := aws.ToString(share.Id)

			log.Printf("[INFO] Deleting AuditManager Framework Share: %s", id)
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceFrameworkShare, id, client))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Framework Shares for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Framework Shares sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
