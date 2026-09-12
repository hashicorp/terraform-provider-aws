// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_interconnect_connection", sweepConnections)
}

func sweepConnections(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.InterconnectClient(ctx)

	var sweepResources []sweep.Sweepable

	input := interconnect.ListConnectionsInput{}
	pages := interconnect.NewListConnectionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Connections {
			sweepResources = append(sweepResources, framework.NewSweepResource(newConnectionResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
