//go:build sweep
// +build sweep

package transfer

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_transfer_server", &resource.Sweeper{
		Name: "aws_transfer_server",
		F:    sweepServers,
	})
}

func sweepServers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).TransferConn
	input := &transfer.ListServersInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListServersPages(input, func(page *transfer.ListServersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, server := range page.Servers {
			r := ResourceServer()
			d := r.Data(nil)
			d.SetId(aws.StringValue(server.ServerId))
			d.Set("force_destroy", true) // In lieu of an aws_transfer_user sweeper.
			d.Set("identity_provider_type", server.IdentityProviderType)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Transfer Server sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Transfer Servers (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Transfer Servers (%s): %w", region, err)
	}

	return nil
}
