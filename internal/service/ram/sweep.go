// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_ram_resource_share", sweepResourceShares)
}

func sweepResourceShares(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RAMClient(ctx)
	input := ram.GetResourceSharesInput{
		ResourceOwner: awstypes.ResourceOwnerSelf,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ram.NewGetResourceSharesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ResourceShares {
			if v.Status == awstypes.ResourceShareStatusDeleted {
				continue
			}

			r := resourceResourceShare()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ResourceShareArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
