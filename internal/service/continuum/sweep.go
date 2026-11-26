// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_continuum_integration", sweepIntegrations)
}

func sweepIntegrations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ContinuumClient(ctx)
	var input continuum.ListIntegrationsInput
	var sweepResources []sweep.Sweepable

	pages := continuum.NewListIntegrationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newIntegrationResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.IntegrationId)),
			))
		}
	}

	return sweepResources, nil
}
