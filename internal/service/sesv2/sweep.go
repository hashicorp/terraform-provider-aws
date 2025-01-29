// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
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
	input := &sesv2.ListConfigurationSetsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sesv2.NewListConfigurationSetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SESv2 Configuration Set sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SESv2 Configuration Sets (%s): %w", region, err)
		}

		for _, v := range page.ConfigurationSets {
			r := resourceConfigurationSet()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SESv2 Configuration Sets (%s): %w", region, err)
	}

	return nil
}

func sweepContactLists(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SESV2Client(ctx)
	input := &sesv2.ListContactListsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sesv2.NewListContactListsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SESv2 Contact List sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SESv2 Contact Lists (%s): %w", region, err)
		}

		for _, v := range page.ContactLists {
			r := resourceContactList()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ContactListName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SESv2 Contact Lists (%s): %w", region, err)
	}

	return nil
}
