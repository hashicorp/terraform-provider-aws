// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package cloudsearch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudsearch_domain", &resource.Sweeper{
		Name: "aws_cloudsearch_domain",
		F:    sweepDomains,
	})
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudSearchConn(ctx)
	input := &cloudsearch.DescribeDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	domains, err := conn.DescribeDomainsWithContext(ctx, input)

	for _, domain := range domains.DomainStatusList {
		if aws.BoolValue(domain.Deleted) {
			continue
		}

		r := ResourceDomain()
		d := r.Data(nil)
		d.SetId(aws.StringValue(domain.DomainName))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudSearch Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudSearch Domains (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudSearch Domains (%s): %w", region, err)
	}

	return nil
}
