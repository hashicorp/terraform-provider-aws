// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudsearch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudsearch"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
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
	conn := client.CloudSearchClient(ctx)
	input := &cloudsearch.DescribeDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	domains, err := conn.DescribeDomains(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudSearch Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudSearch Domains (%s): %w", region, err)
	}

	for _, v := range domains.DomainStatusList {
		name := aws.ToString(v.DomainName)

		if deleted := aws.ToBool(v.Deleted); deleted {
			log.Printf("[INFO] Skipping CloudSearch Domain %s: Deleted=%t", name, deleted)
			continue
		}

		r := resourceDomain()
		d := r.Data(nil)
		d.SetId(name)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudSearch Domains (%s): %w", region, err)
	}

	return nil
}
