//go:build sweep
// +build sweep

package ssm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_ssm_default_patch_baseline", &resource.Sweeper{
		Name: "aws_ssm_default_patch_baseline",
		F:    sweepResourceDefaultPatchBaselines,
	})

	resource.AddTestSweepers("aws_ssm_maintenance_window", &resource.Sweeper{
		Name: "aws_ssm_maintenance_window",
		F:    sweepMaintenanceWindows,
	})

	resource.AddTestSweepers("aws_ssm_patch_baseline", &resource.Sweeper{
		Name: "aws_ssm_patch_baseline",
		F:    sweepResourcePatchBaselines,
		Dependencies: []string{
			"aws_ssm_default_patch_baseline",
		},
	})

	resource.AddTestSweepers("aws_ssm_resource_data_sync", &resource.Sweeper{
		Name: "aws_ssm_resource_data_sync",
		F:    sweepResourceDataSyncs,
	})
}

func sweepResourceDefaultPatchBaselines(region string) error {
	ctx := sweep.Context(region)
	c, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	client := c.(ssmClient)

	conn := client.SSMClient()

	var sweepables []sweep.Sweepable
	var errs *multierror.Error

	paginator := patchBaselinesPaginator(conn, ownerIsSelfFilter())
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("listing Default Patch Baselines for %s: %w", region, err))
			break
		}

		for _, identity := range tfslices.Filter(page.BaselineIdentities, func(b types.PatchBaselineIdentity) bool {
			return b.DefaultBaseline
		}) {
			baselineID := aws.ToString(identity.BaselineId)
			pb, err := findPatchBaselineByID(ctx, conn, baselineID)
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("reading Patch Baseline (%s): %w", baselineID, err))
				continue
			}
			sweepables = append(sweepables, defaultPatchBaselineSweeper{
				client: client,
				os:     pb.OperatingSystem,
			})
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepables); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Default Patch Baselines for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Default Patch Baselines sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

type defaultPatchBaselineSweeper struct {
	client ssmClient
	os     types.OperatingSystem
}

func (s defaultPatchBaselineSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) (err error) {
	diags := defaultPatchBaselineRestoreOSDefault(ctx, s.client, s.os)

	for _, d := range sdkdiag.Warnings(diags) {
		log.Printf("[WARN] %s", sdkdiag.DiagnosticString(d))
	}

	for _, d := range sdkdiag.Errors(diags) {
		err = multierror.Append(err, errors.New(sdkdiag.DiagnosticString(d)))
	}
	return
}

func sweepMaintenanceWindows(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).SSMConn()
	input := &ssm.DescribeMaintenanceWindowsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeMaintenanceWindowsWithContext(ctx, input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSM Maintenance Window sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving SSM Maintenance Windows: %s", err)
		}

		for _, window := range output.WindowIdentities {
			id := aws.ToString(window.WindowId)
			input := &ssm.DeleteMaintenanceWindowInput{
				WindowId: window.WindowId,
			}

			log.Printf("[INFO] Deleting SSM Maintenance Window: %s", id)

			_, err := conn.DeleteMaintenanceWindowWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, ssm.ErrCodeDoesNotExistException) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("deleting SSM Maintenance Window (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.ToString(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func sweepResourcePatchBaselines(region string) error {
	ctx := sweep.Context(region)
	c, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	client := c.(ssmClient)

	conn := client.SSMClient()

	var sweepables []sweep.Sweepable
	var errs *multierror.Error

	paginator := patchBaselinesPaginator(conn, ownerIsSelfFilter())
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("listing Patch Baselines for %s: %w", region, err))
			break
		}

		for _, identity := range page.BaselineIdentities {
			baselineID := aws.ToString(identity.BaselineId)
			r := ResourcePatchBaseline()
			d := r.Data(nil)

			d.SetId(baselineID)

			sweepables = append(sweepables, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepables); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Patch Baselines for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Patch Baselines sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepResourceDataSyncs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).SSMConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &ssm.ListResourceDataSyncInput{}

	err = conn.ListResourceDataSyncPagesWithContext(ctx, input, func(page *ssm.ListResourceDataSyncOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, resourceDataSync := range page.ResourceDataSyncItems {
			r := ResourceResourceDataSync()
			d := r.Data(nil)

			d.SetId(aws.ToString(resourceDataSync.SyncName))
			d.Set("name", resourceDataSync.SyncName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing SSM Resource Data Sync for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping SSM Resource Data Sync for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping SSM Resource Data Sync sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
