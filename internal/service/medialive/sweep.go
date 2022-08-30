//go:build sweep
// +build sweep

package medialive

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_medialive_input_security_group", &resource.Sweeper{
		Name: "aws_medialive_input_security_group",
		F:    sweepInputSecurityGroups,
	})
}

func sweepInputSecurityGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	conn := client.(*conns.AWSClient).MediaLiveConn
	sweepResources := make([]*sweep.SweepResource, 0)
	in := &medialive.ListInputSecurityGroupsInput{}
	var errs *multierror.Error

	pages := medialive.NewListInputSecurityGroupsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if sweep.SkipSweepError(err) {
			log.Println("[WARN] Skipping MediaLive Security Groups sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving MediaLive Security Groups: %w", err)
		}

		for _, group := range page.InputSecurityGroups {
			id := aws.ToString(group.Id)
			log.Printf("[INFO] Deleting MediaLive Security Group: %s", id)

			r := ResourceInputSecurityGroup()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping MediaLive Security Groups for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping MediaLive Security Groups sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
