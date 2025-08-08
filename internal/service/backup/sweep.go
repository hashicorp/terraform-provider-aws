// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_backup_framework", &resource.Sweeper{
		Name: "aws_backup_framework",
		F:    sweepFrameworks,
	})

	resource.AddTestSweepers("aws_backup_plan", &resource.Sweeper{
		Name: "aws_backup_plan",
		F:    sweepPlans,
		Dependencies: []string{
			"aws_backup_selection",
		},
	})

	resource.AddTestSweepers("aws_backup_selection", &resource.Sweeper{
		Name: "aws_backup_selection",
		F:    sweepSelections,
	})

	resource.AddTestSweepers("aws_backup_report_plan", &resource.Sweeper{
		Name: "aws_backup_report_plan",
		F:    sweepReportPlans,
	})

	resource.AddTestSweepers("aws_backup_restore_testing_plan", &resource.Sweeper{
		Name: "aws_backup_restore_testing_plan",
		F:    sweepRestoreTestingPlans,
		Dependencies: []string{
			"aws_backup_restore_testing_selection",
		},
	})

	resource.AddTestSweepers("aws_backup_restore_testing_selection", &resource.Sweeper{
		Name: "aws_backup_restore_testing_selection",
		F:    sweepRestoreTestingSelections,
	})

	resource.AddTestSweepers("aws_backup_vault_lock_configuration", &resource.Sweeper{
		Name: "aws_backup_vault_lock_configuration",
		F:    sweepVaultLockConfigurations,
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

func sweepFrameworks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListFrameworksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListFrameworksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Framework sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Frameworks (%s): %w", region, err)
		}

		for _, v := range page.Frameworks {
			r := resourceFramework()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FrameworkName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Frameworks (%s): %w", region, err)
	}

	return nil
}

func sweepPlans(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupPlansInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupPlansPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Plan sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Plans (%s): %w", region, err)
		}

		for _, v := range page.BackupPlansList {
			planID := aws.ToString(v.BackupPlanId)

			if strings.HasPrefix(planID, "aws/") {
				log.Printf("[INFO] Skipping Backup Plan: %s", planID)
				continue
			}

			r := resourcePlan()
			d := r.Data(nil)
			d.SetId(planID)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Plans (%s): %w", region, err)
	}

	return nil
}

func sweepSelections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupPlansInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupPlansPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Selection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Plans (%s): %w", region, err)
		}

		for _, v := range page.BackupPlansList {
			planID := aws.ToString(v.BackupPlanId)

			if strings.HasPrefix(planID, "aws/") {
				log.Printf("[INFO] Skipping Backup Plan: %s", planID)
				continue
			}

			input := &backup.ListBackupSelectionsInput{
				BackupPlanId: aws.String(planID),
			}

			pages := backup.NewListBackupSelectionsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.BackupSelectionsList {
					r := resourceSelection()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.SelectionId))
					d.Set("plan_id", planID)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Selections (%s): %w", region, err)
	}

	return nil
}

func sweepReportPlans(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListReportPlansInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListReportPlansPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Report Plan sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Report Plans for %s: %w", region, err)
		}

		for _, v := range page.ReportPlans {
			r := resourceReportPlan()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ReportPlanName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Report Plans for %s: %w", region, err)
	}

	return nil
}

func sweepRestoreTestingPlans(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListRestoreTestingPlansInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListRestoreTestingPlansPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Restore Testing Plan sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Restore Testing Plans for %s: %w", region, err)
		}

		for _, v := range page.RestoreTestingPlans {
			sweepResources = append(sweepResources, framework.NewSweepResource(newRestoreTestingPlanResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.RestoreTestingPlanName))))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Restore Testing Plans for %s: %w", region, err)
	}

	return nil
}

func sweepRestoreTestingSelections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListRestoreTestingPlansInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListRestoreTestingPlansPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Restore Testing Plan sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Restore Testing Plans for %s: %w", region, err)
		}

		for _, v := range page.RestoreTestingPlans {
			restoreTestingPlanName := aws.ToString(v.RestoreTestingPlanName)
			input := &backup.ListRestoreTestingSelectionsInput{
				RestoreTestingPlanName: aws.String(restoreTestingPlanName),
			}

			pages := backup.NewListRestoreTestingSelectionsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.RestoreTestingSelections {
					sweepResources = append(sweepResources, framework.NewSweepResource(newRestoreTestingSelectionResource, client,
						framework.NewAttribute(names.AttrName, aws.ToString(v.RestoreTestingSelectionName)),
						framework.NewAttribute("restore_testing_plan_name", restoreTestingPlanName)))
				}
			}
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Restore Testing Selections for %s: %w", region, err)
	}

	return nil
}

func sweepVaultLockConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupVaultsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupVaultsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Vault Lock Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Vaults for %s: %w", region, err)
		}

		for _, v := range page.BackupVaultList {
			r := resourceVaultLockConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Vault Lock Configurations for %s: %w", region, err)
	}

	return nil
}

func sweepVaultNotifications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupVaultsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupVaultsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Vault Notifications sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Vaults for %s: %w", region, err)
		}

		for _, v := range page.BackupVaultList {
			r := resourceVaultNotifications()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Vault Notifications for %s: %w", region, err)
	}

	return nil
}

func sweepVaultPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupVaultsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupVaultsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Vault Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Vaults for %s: %w", region, err)
		}

		for _, v := range page.BackupVaultList {
			r := resourceVaultPolicy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Vault Policies for %s: %w", region, err)
	}

	return nil
}

func sweepVaults(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.BackupClient(ctx)
	input := &backup.ListBackupVaultsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := backup.NewListBackupVaultsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Backup Vault sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Backup Vaults for %s: %w", region, err)
		}

		for _, v := range page.BackupVaultList {
			name := aws.ToString(v.BackupVaultName)

			// Ignore Default and Automatic EFS Backup Vaults in region (cannot be deleted)
			if name == "Default" || name == "aws/efs/automatic-backup-vault" {
				log.Printf("[INFO] Skipping Backup Vault: %s", name)
				continue
			}

			r := resourceVault()
			d := r.Data(nil)
			d.SetId(name)
			d.Set(names.AttrForceDestroy, true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Backup Vaults for %s: %w", region, err)
	}

	return nil
}
