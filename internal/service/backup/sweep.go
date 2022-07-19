//go:build sweep
// +build sweep

package backup

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).BackupConn
	input := &backup.ListFrameworksInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListFrameworksPages(input, func(page *backup.ListFrameworksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, framework := range page.Frameworks {
			r := ResourceFramework()
			d := r.Data(nil)
			d.SetId(aws.StringValue(framework.FrameworkName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Backup Framework sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Frameworks for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Frameworks for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepReportPlan(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).BackupConn
	input := &backup.ListReportPlansInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListReportPlansPages(input, func(page *backup.ListReportPlansOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, reportPlan := range page.ReportPlans {
			r := ResourceReportPlan()
			d := r.Data(nil)
			d.SetId(aws.StringValue(reportPlan.ReportPlanName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Backup Report Plans sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Report Plans for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Report Plans for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVaultLockConfiguration(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).BackupConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &backup.ListBackupVaultsInput{}

	err = conn.ListBackupVaultsPages(input, func(page *backup.ListBackupVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.BackupVaultList {
			if vault == nil {
				continue
			}

			r := ResourceVaultLockConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(vault.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Backup Vault Lock Configuration for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Backup Vault Lock Configuration sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepVaultNotifications(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).BackupConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &backup.ListBackupVaultsInput{}

	err = conn.ListBackupVaultsPages(input, func(page *backup.ListBackupVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.BackupVaultList {
			if vault == nil {
				continue
			}

			r := ResourceVaultNotifications()
			d := r.Data(nil)
			d.SetId(aws.StringValue(vault.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Backup Vault Notifications for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Backup Vault Notifications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepVaultPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).BackupConn
	input := &backup.ListBackupVaultsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListBackupVaultsPages(input, func(page *backup.ListBackupVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.BackupVaultList {
			r := ResourceVaultPolicy()
			d := r.Data(nil)
			d.SetId(aws.StringValue(vault.BackupVaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Backup Vault Policies sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Vault Policies for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVaults(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).BackupConn
	input := &backup.ListBackupVaultsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListBackupVaultsPages(input, func(page *backup.ListBackupVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.BackupVaultList {
			failedToDeleteRecoveryPoint := false
			name := aws.StringValue(vault.BackupVaultName)

			// Ignore Default and Automatic EFS Backup Vaults in region (cannot be deleted)
			if name == "Default" || name == "aws/efs/automatic-backup-vault" {
				log.Printf("[INFO] Skipping Backup Vault: %s", name)
				continue
			}
			input := &backup.ListRecoveryPointsByBackupVaultInput{
				BackupVaultName: vault.BackupVaultName,
			}

			err := conn.ListRecoveryPointsByBackupVaultPages(input, func(page *backup.ListRecoveryPointsByBackupVaultOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, recoveryPoint := range page.RecoveryPoints {
					_, err := conn.DeleteRecoveryPoint(&backup.DeleteRecoveryPointInput{
						BackupVaultName:  vault.BackupVaultName,
						RecoveryPointArn: recoveryPoint.RecoveryPointArn,
					})

					if err != nil {
						log.Printf("[WARN] Failed to delete Recovery Point (%s) in Backup Vault (%s): %s", aws.StringValue(recoveryPoint.RecoveryPointArn), name, err)
						failedToDeleteRecoveryPoint = true
						return true
					}
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Reovery Points in Backup Vault (%s) for %s: %w", name, region, err))
			}

			// Backup Vault deletion only supported when empty
			// Reference: https://docs.aws.amazon.com/aws-backup/latest/devguide/API_DeleteBackupVault.html
			if failedToDeleteRecoveryPoint {
				log.Printf("[INFO] Skipping Backup Vault (%s): not empty", name)
				continue
			}

			r := ResourceVault()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Backup Vaults sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Vaults for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
