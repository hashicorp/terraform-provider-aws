//go:build sweep
// +build sweep

package simpledb

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_simpledb_domain", &resource.Sweeper{
		Name: "aws_simpledb_domain",
		F:    sweepDomains,
	})
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SimpleDBConn()
	input := &simpledb.ListDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListDomainsPagesWithContext(ctx, input, func(page *simpledb.ListDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DomainNames {
			sweepResources = append(sweepResources, sweep.NewSweepFrameworkResource(newResourceDomain, aws.StringValue(v), client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SimpleDB Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing SimpleDB Domains (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SimpleDB Domains (%s): %w", region, err)
	}

	return nil
}
