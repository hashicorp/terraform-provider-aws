// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package sesv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.SESV2Client(ctx)
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

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Configuration Sets for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Configuration Sets sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepContactLists(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.SESV2Client(ctx)
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

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Contact Lists for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Contact Lists sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
