// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdacore

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdacore"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_lambdacore_network_connector", sweepNetworkConnectors)
}

func sweepNetworkConnectors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LambdaCoreClient(ctx)
	var input lambdacore.ListNetworkConnectorsInput
	var sweepResources []sweep.Sweepable

	pages := lambdacore.NewListNetworkConnectorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.NetworkConnectors {
			sweepResources = append(sweepResources, framework.NewSweepResource(newNetworkConnectorResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
