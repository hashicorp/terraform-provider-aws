// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package keyspaces

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	// No need to have separate sweeper for table as would be destroyed as part of keyspace
	resource.AddTestSweepers("aws_keyspaces_keyspace", &resource.Sweeper{
		Name: "aws_keyspaces_keyspace",
		F:    sweepKeyspaces,
	})
}

func sweepKeyspaces(region string) error { // nosemgrep:ci.keyspaces-in-func-name
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.KeyspacesClient(ctx)
	input := &keyspaces.ListKeyspacesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := keyspaces.NewListKeyspacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Keyspaces Keyspace sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Keyspaces Keyspaces (%s): %w", region, err)
		}

		for _, v := range page.Keyspaces {
			id := aws.ToString(v.KeyspaceName)

			switch id {
			case "system_schema", "system_schema_mcs", "system", "system_multiregion_info":
				// The default keyspaces cannot be deleted.
				continue
			}

			r := resourceKeyspace()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Keyspaces Keyspaces (%s): %w", region, err)
	}

	return nil
}
