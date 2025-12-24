// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sesv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_sesv2_configuration_set", sweepConfigurationSets)
	awsv2.Register("aws_sesv2_contact_list", sweepContactLists)
}

func sweepConfigurationSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SESV2Client(ctx)
	var input sesv2.ListConfigurationSetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sesv2.NewListConfigurationSetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ConfigurationSets {
			r := resourceConfigurationSet()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepContactLists(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SESV2Client(ctx)
	var input sesv2.ListContactListsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sesv2.NewListContactListsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ContactLists {
			r := resourceContactList()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ContactListName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
