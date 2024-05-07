// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	sweep.Register("aws_lakeformation_resource", sweepResource)
}

func sweepResource(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LakeFormationClient(ctx)

	var sweepResources []sweep.Sweepable
	r := ResourceResource()

	pages := lakeformation.NewListResourcesPaginator(conn, &lakeformation.ListResourcesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		for _, v := range page.ResourceInfoList {
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ResourceArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
