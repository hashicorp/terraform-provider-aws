// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	input := &elasticsearchservice.ListDomainNamesInput{
		EngineType: awstypes.EngineTypeElasticsearch,
	}
	conn := client.ElasticsearchClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.ListDomainNames(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Elasticsearch Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing MSK Clusters (%s): %w", region, err)
	}

	for _, v := range output.DomainNames {
		name := aws.ToString(v.DomainName)

		if engineType := v.EngineType; engineType != awstypes.EngineTypeElasticsearch {
			log.Printf("[INFO] Skipping Elasticsearch Domain %s: EngineType=%s", name, engineType)
			continue
		}

		output, err := findDomainByName(ctx, conn, name)

		if err != nil {
			continue
		}

		if aws.ToBool(output.Deleted) {
			log.Printf("[INFO] Skipping Elasticsearch Domain %s: Deleted", name)
			continue
		}

		r := resourceDomain()
		d := r.Data(nil)
		d.SetId(name)
		d.Set(names.AttrDomainName, name)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Elasticsearch Domains (%s): %w", region, err)
	}

	return nil
}
