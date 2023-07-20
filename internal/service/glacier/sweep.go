// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package glacier

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glacier"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &glacier.ListVaultsInput{}
	conn := client.GlacierClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := glacier.NewListVaultsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Glacier Vault sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Glacier Vaults (%s): %w", region, err)
		}

		for _, v := range page.VaultList {
			r := resourceVault()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VaultName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Glacier Vaults (%s): %w", region, err)
	}

	return nil
}
