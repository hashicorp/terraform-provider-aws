// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_qbusiness_application", sweepApplications)
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &qbusiness.ListApplicationsInput{}
	conn := client.QBusinessClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := qbusiness.NewListApplicationsPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Applications {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceApplication, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.ApplicationId))),
			)
		}
	}

	return sweepResources, nil
}
