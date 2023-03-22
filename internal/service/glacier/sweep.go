//go:build sweep
// +build sweep

package glacier

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &glacier.ListVaultsInput{}
	conn := client.(*conns.AWSClient).GlacierConn()
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListVaultsPagesWithContext(ctx, input, func(page *glacier.ListVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VaultList {
			r := ResourceVault()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.VaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Glacier Vault sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Glacier Vaults (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Glacier Vaults (%s): %w", region, err)
	}

	return nil
}
