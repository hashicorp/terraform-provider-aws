//go:build sweep
// +build sweep

package keyspaces

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/keyspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	// No need to have separate sweeper for table as would be destroyed as part of keyspace
	resource.AddTestSweepers("aws_keyspaces_keyspace", &resource.Sweeper{
		Name: "aws_keyspaces_keyspace",
		F:    sweepKeyspaces,
	})
}

func sweepKeyspaces(region string) error { // nosemgrep:keyspaces-in-func-name
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).KeyspacesConn
	input := &keyspaces.ListKeyspacesInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListKeyspacesPages(input, func(page *keyspaces.ListKeyspacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Keyspaces {
			id := aws.StringValue(v.KeyspaceName)

			switch id {
			case "system_schema", "system_schema_mcs", "system":
				// The default keyspaces cannot be deleted.
				continue
			}

			r := ResourceKeyspace()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Keyspaces Keyspace sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Keyspaces Keyspaces (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Keyspaces Keyspaces (%s): %w", region, err)
	}

	return nil
}
