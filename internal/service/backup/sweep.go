// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_backup_framework", &resource.Sweeper{
		Name: "aws_backup_framework",
		F:    sweepFramework,
	})

	resource.AddTestSweepers("aws_backup_report_plan", &resource.Sweeper{
		Name: "aws_backup_report_plan",
		F:    sweepReportPlan,
	})

	resource.AddTestSweepers("aws_backup_vault_lock_configuration", &resource.Sweeper{
		Name: "aws_backup_vault_lock_configuration",
		F:    sweepVaultLockConfiguration,
	})

	resource.AddTestSweepers("aws_backup_vault_notifications", &resource.Sweeper{
		Name: "aws_backup_vault_notifications",
		F:    sweepVaultNotifications,
	})

	resource.AddTestSweepers("aws_backup_vault_policy", &resource.Sweeper{
		Name: "aws_backup_vault_policy",
		F:    sweepVaultPolicies,
	})

	resource.AddTestSweepers("aws_backup_vault", &resource.Sweeper{
		Name: "aws_backup_vault",
		F:    sweepVaults,
		Dependencies: []string{
			"aws_backup_vault_lock_configuration",
			"aws_backup_vault_notifications",
			"aws_backup_vault_policy",
		},
	})
}

func sweepFramework(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListFrameworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListFrameworksPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Framework sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Frameworks for %s: %w", region, err))
		}

		for _, framework := range page.Frameworks {
			r := ResourceFramework()
			d := r.Data(nil)
			d.SetId(aws.ToString(framework.FrameworkName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Frameworks for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepReportPlan(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListReportPlansInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListReportPlansPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Report Plans sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Report Plans for %s: %w", region, err))
		}

		for _, reportPlan := range page.ReportPlans {
			r := ResourceReportPlan()
			d := r.Data(nil)
			d.SetId(aws.ToString(reportPlan.ReportPlanName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Report Plans for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVaultLockConfiguration(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}

	conn := client.BackupClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &backup.ListBackupVaultsInput{}

	pages := backup.NewListBackupVaultsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
		}

		for _, vault := range page.BackupVaultList {
			r := ResourceVaultLockConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(vault.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Backup Vault Lock Configuration for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Backup Vault Lock Configuration sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepVaultNotifications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}

	conn := client.BackupClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &backup.ListBackupVaultsInput{}

	pages := backup.NewListBackupVaultsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
		}

		for _, vault := range page.BackupVaultList {
			r := ResourceVaultNotifications()
			d := r.Data(nil)
			d.SetId(aws.ToString(vault.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Backup Vault Notifications for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Backup Vault Notifications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepVaultPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupVaultsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupVaultsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Vault Policies sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
		}

		for _, vault := range page.BackupVaultList {
			r := ResourceVaultPolicy()
			d := r.Data(nil)
			d.SetId(aws.ToString(vault.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Vault Policies for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVaults(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupVaultsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupVaultsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Vaults sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
		}

		for _, vault := range page.BackupVaultList {
			name := aws.ToString(vault.BackupVaultName)

			// Ignore Default and Automatic EFS Backup Vaults in region (cannot be deleted)
			if name == "Default" || name == "aws/efs/automatic-backup-vault" {
				log.Printf("[INFO] Skipping Backup Vault: %s", name)
				continue
			}

			r := ResourceVault()
			d := r.Data(nil)
			d.SetId(name)
			d.Set(names.AttrForceDestroy, true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Vaults for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
