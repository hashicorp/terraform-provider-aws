// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_ssm_default_patch_baseline", &resource.Sweeper{
		Name: "aws_ssm_default_patch_baseline",
		F:    sweepDefaultPatchBaselines,
	})

	resource.AddTestSweepers("aws_ssm_maintenance_window", &resource.Sweeper{
		Name: "aws_ssm_maintenance_window",
		F:    sweepMaintenanceWindows,
	})

	resource.AddTestSweepers("aws_ssm_patch_baseline", &resource.Sweeper{
		Name: "aws_ssm_patch_baseline",
		F:    sweepPatchBaselines,
		Dependencies: []string{
			"aws_ssm_default_patch_baseline",
			"aws_ssm_patch_group",
		},
	})

	resource.AddTestSweepers("aws_ssm_patch_group", &resource.Sweeper{
		Name: "aws_ssm_patch_group",
		F:    sweepPatchGroups,
	})

	resource.AddTestSweepers("aws_ssm_resource_data_sync", &resource.Sweeper{
		Name: "aws_ssm_resource_data_sync",
		F:    sweepResourceDataSyncs,
	})
}

func sweepDefaultPatchBaselines(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SSMClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := patchBaselinesPaginator(conn, ownerIsSelfFilter())
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSM Default Patch Baseline sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SSM Default Patch Baselines (%s): %w", region, err)
		}

		for _, identity := range tfslices.Filter(page.BaselineIdentities, func(v awstypes.PatchBaselineIdentity) bool {
			return v.DefaultBaseline
		}) {
			baselineID := aws.ToString(identity.BaselineId)
			pb, err := findPatchBaselineByID(ctx, conn, baselineID)

			if err != nil {
				continue
			}
			sweepResources = append(sweepResources, defaultPatchBaselineSweeper{
				conn: conn,
				os:   pb.OperatingSystem,
			})
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SSM Default Patch Baselines (%s): %w", region, err)
	}

	return nil
}

type defaultPatchBaselineSweeper struct {
	conn *ssm.Client
	os   awstypes.OperatingSystem
}

func (s defaultPatchBaselineSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	diags := defaultPatchBaselineRestoreOSDefault(ctx, s.conn, s.os)

	for _, d := range sdkdiag.Warnings(diags) {
		log.Printf("[WARN] %s", sdkdiag.DiagnosticString(d))
	}

	return sdkdiag.DiagnosticsError(diags)
}

func sweepMaintenanceWindows(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SSMClient(ctx)
	input := &ssm.DescribeMaintenanceWindowsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ssm.NewDescribeMaintenanceWindowsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSM Maintenance Window sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SSM Maintenance Windows (%s): %w", region, err)
		}

		for _, v := range page.WindowIdentities {
			r := resourceMaintenanceWindow()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WindowId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SSM Maintenance Windows (%s): %w", region, err)
	}

	return nil
}

func sweepPatchBaselines(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SSMClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := patchBaselinesPaginator(conn, ownerIsSelfFilter())
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSM Patch Baseline sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SSM Patch Baselines (%s): %w", region, err)
		}

		for _, v := range page.BaselineIdentities {
			r := resourcePatchBaseline()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BaselineId))
			d.Set("operating_system", v.OperatingSystem)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SSM Patch Baselines (%s): %w", region, err)
	}

	return nil
}

func sweepPatchGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SSMClient(ctx)
	input := &ssm.DescribePatchGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ssm.NewDescribePatchGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSM Patch Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SSM Patch Groups (%s): %w", region, err)
		}

		for _, v := range page.Mappings {
			r := resourcePatchGroup()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", aws.ToString(v.PatchGroup), aws.ToString(v.BaselineIdentity.BaselineId)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SSM Patch Groups (%s): %w", region, err)
	}

	return nil
}

func sweepResourceDataSyncs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SSMClient(ctx)
	input := &ssm.ListResourceDataSyncInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ssm.NewListResourceDataSyncPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSM Resource Data Sync sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SSM Resource Data Syncs (%s): %w", region, err)
		}

		for _, v := range page.ResourceDataSyncItems {
			r := resourceResourceDataSync()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SyncName))
			d.Set(names.AttrName, v.SyncName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SSM Resource Data Syncs (%s): %w", region, err)
	}

	return nil
}
