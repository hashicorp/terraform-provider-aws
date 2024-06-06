// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_elasticsearch_domain", &resource.Sweeper{
		Name: "aws_elasticsearch_domain",
		F:    sweepDomains,
	})
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ElasticsearchConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &elasticsearchservice.ListDomainNamesInput{}

	// ListDomainNames has no pagination support whatsoever
	output, err := conn.ListDomainNamesWithContext(ctx, input)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Elasticsearch Domain sweep for %s: %s", region, err)
		return errs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing Elasticsearch Domains: %w", err)
		log.Printf("[ERROR] %s", sweeperErr)
		errs = multierror.Append(errs, sweeperErr)
		return errs.ErrorOrNil()
	}

	if output == nil {
		log.Printf("[WARN] Skipping Elasticsearch Domain sweep for %s: empty response", region)
		return errs.ErrorOrNil()
	}

	for _, domainInfo := range output.DomainNames {
		if domainInfo == nil {
			continue
		}

		name := aws.StringValue(domainInfo.DomainName)

		if engineType := aws.StringValue(domainInfo.EngineType); engineType != elasticsearchservice.EngineTypeElasticsearch {
			log.Printf("[INFO] Skipping Elasticsearch Domain %s: EngineType = %s", name, engineType)
			continue
		}

		// Elasticsearch Domains have regularly gotten stuck in a "being deleted" state
		// e.g. Deleted and Processing are both true for days in the API
		// Filter out domains that are Deleted already.

		output, err := FindDomainByName(ctx, conn, name)
		if err != nil {
			sweeperErr := fmt.Errorf("error describing Elasticsearch Domain (%s): %w", name, err)
			log.Printf("[ERROR] %s", sweeperErr)
			errs = multierror.Append(errs, sweeperErr)
			continue
		}

		if output != nil && aws.BoolValue(output.Deleted) {
			log.Printf("[INFO] Skipping Elasticsearch Domain (%s) with deleted status", name)
			continue
		}

		r := ResourceDomain()
		d := r.Data(nil)
		d.SetId(name)
		d.Set(names.AttrDomainName, name)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Elasticsearch Domains for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Elasticsearch Domain sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
