// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_mediaconnect_flow", sweepFlows)
}

func sweepFlows(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.MediaConnectClient(ctx)
	input := mediaconnect.ListFlowsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := mediaconnect.NewListFlowsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Flows {
			sweepResources = append(sweepResources, framework.NewSweepResource(newFlowResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.FlowArn))))
		}
	}

	return sweepResources, nil
}
