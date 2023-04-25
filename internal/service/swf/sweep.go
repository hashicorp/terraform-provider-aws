//go:build sweep
// +build sweep

package swf

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/swf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_swf_domain", &resource.Sweeper{
		Name: "aws_swf_domain",
		F:    sweepDomains,
	})
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SWFConn()
	input := &swf.ListDomainsInput{
		RegistrationStatus: aws.String(swf.RegistrationStatusRegistered),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListDomainsPagesWithContext(ctx, input, func(page *swf.ListDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DomainInfos {
			r := ResourceDomain()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SWF Domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing SWF Domains (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SWF Domains (%s): %w", region, err)
	}

	return nil
}
