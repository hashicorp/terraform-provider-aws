//go:build sweep
// +build sweep

package glacier

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_glacier_vault", &resource.Sweeper{
		Name: "aws_glacier_vault",
		F:    sweepVaults,
	})
}

func sweepVaults(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).GlacierConn
	var sweeperErrs *multierror.Error

	err = conn.ListVaultsPages(&glacier.ListVaultsInput{}, func(page *glacier.ListVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.VaultList {
			name := aws.StringValue(vault.VaultName)

			// First attempt to delete the vault's notification configuration in case the vault deletion fails.
			log.Printf("[INFO] Deleting Glacier Vault (%s) Notifications", name)
			_, err := conn.DeleteVaultNotifications(&glacier.DeleteVaultNotificationsInput{
				VaultName: aws.String(name),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Glacier Vault (%s) Notifications: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			log.Printf("[INFO] Deleting Glacier Vault: %s", name)
			_, err = conn.DeleteVault(&glacier.DeleteVaultInput{
				VaultName: aws.String(name),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Glacier Vault (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Glacier Vaults sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Glacier Vaults: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
