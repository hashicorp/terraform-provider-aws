//go:build sweep
// +build sweep

package auditmanager

import (
	"context"
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
	resource.AddTestSweepers("aws_auditmanager_control", &resource.Sweeper{
		Name: "aws_auditmanager_control",
		F:    sweepControls,
	})
	resource.AddTestSweepers("aws_auditmanager_framework", &resource.Sweeper{
		Name: "aws_auditmanager_framework",
		F:    sweepFrameworks,
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

func sweepControls(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	conn := client.(*conns.AWSClient).AuditManagerClient
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

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Controls for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Controls sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepFrameworks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	conn := client.(*conns.AWSClient).AuditManagerClient
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

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AuditManager Frameworks for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AuditManager Frameworks sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
