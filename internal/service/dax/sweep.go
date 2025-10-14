// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dax"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_dax_cluster", sweepClusters)
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DAXClient(ctx)
	var input dax.DescribeClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := describeClustersPages(ctx, conn, &input, func(page *dax.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(cluster.ClusterName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	// GovCloud (with no DAX support) has an endpoint that responds with:
	// InvalidParameterValueException: Access Denied to API Version: DAX_V3
	if errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "Access Denied to API Version: DAX_V3") {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
