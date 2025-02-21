// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_neptunegraph_graph", sweepGraphs)
}

func sweepGraphs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input neptunegraph.ListGraphsInput
	conn := client.NeptuneGraphClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptunegraph.NewListGraphsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Graphs {
			id := aws.ToString(v.Id)

			if aws.ToBool(v.DeletionProtection) {
				input := neptunegraph.UpdateGraphInput{
					DeletionProtection: aws.Bool(false),
					GraphIdentifier:    aws.String(id),
				}

				if _, err := conn.UpdateGraph(ctx, &input); err != nil {
					return nil, fmt.Errorf("updating Graph (%s) DeletionProtection: %w", id, err)
				}

				const (
					timeout = 30 * time.Minute
				)
				if _, err := waitGraphUpdated(ctx, conn, id, timeout); err != nil {
					return nil, fmt.Errorf("waiting for Graph (%s) update: %w", id, err)
				}
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newGraphResource, client,
				framework.NewAttribute(names.AttrID, id)))
		}
	}

	return sweepResources, nil
}
