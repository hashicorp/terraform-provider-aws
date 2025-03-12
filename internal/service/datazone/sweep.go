// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_datazone_domain", sweepDomains)
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DataZoneClient(ctx)
	input := &datazone.ListDomainsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := datazone.NewListDomainsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, da := range page.Items {
			id := aws.ToString(da.Id)

			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceDomain, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	return sweepResources, nil
}
