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
			input := &backup.ListRecoveryPointsByBackupVaultInput{
				BackupVaultName: aws.String(name),
			}

			err := conn.ListRecoveryPointsByBackupVaultPages(input, func(page *backup.ListRecoveryPointsByBackupVaultOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, recoveryPoint := range page.RecoveryPoints {
					arn := aws.StringValue(recoveryPoint.RecoveryPointArn)

					log.Printf("[INFO] Deleting Recovery Point (%s) in Backup Vault (%s)", arn, name)
					_, err := conn.DeleteRecoveryPoint(&backup.DeleteRecoveryPointInput{
						BackupVaultName:  aws.String(name),
						RecoveryPointArn: aws.String(arn),
					})

					if err != nil {
						log.Printf("[WARN] Failed to delete Recovery Point (%s) in Backup Vault (%s): %s", arn, name, err)
						failedToDeleteRecoveryPoint = true
					}
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Reovery Points in Backup Vault (%s) for %s: %w", name, region, err))
			}

			// Ignore Default and Automatic EFS Backup Vaults in region (cannot be deleted)
			if name == "Default" || name == "aws/efs/automatic-backup-vault" {
				log.Printf("[INFO] Skipping Backup Vault: %s", name)
				continue
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
