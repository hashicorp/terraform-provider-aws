//go:build sweep
// +build sweep

package sesv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_sesv2_configuration_set", &resource.Sweeper{
		Name: "aws_sesv2_configuration_set",
		F:    sweepConfigurationSets,
	})

	resource.AddTestSweepers("aws_sesv2_contact_list", &resource.Sweeper{
		Name: "aws_sesv2_contact_list",
		F:    sweepContactLists,
	})
}

func sweepConfigurationSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).SESV2Client()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &sesv2.ListConfigurationSetsInput{}

	err = ListConfigurationSetsPages(ctx, conn, input, func(page *sesv2.ListConfigurationSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, configurationSet := range page.ConfigurationSets {
			r := ResourceConfigurationSet()
			d := r.Data(nil)

			d.SetId(configurationSet)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing Configuration Sets for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Configuration Sets for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Configuration Sets sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepContactLists(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).SESV2Client()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &sesv2.ListContactListsInput{}

	err = ListContactListsPages(ctx, conn, input, func(page *sesv2.ListContactListsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, contactList := range page.ContactLists {
			r := ResourceContactList()
			d := r.Data(nil)

			d.SetId(aws.ToString(contactList.ContactListName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing Contact Lists for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Contact Lists for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Contact Lists sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
