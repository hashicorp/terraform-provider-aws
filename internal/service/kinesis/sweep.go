// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_kinesis_stream", &resource.Sweeper{
		Name: "aws_kinesis_stream",
		F:    sweepStreams,
	})
}

func sweepStreams(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.KinesisClient(ctx)
	input := &kinesis.ListStreamsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kinesis.NewListStreamsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Kinesis Stream sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Kinesis Streams (%s): %w", region, err)
		}

		for _, v := range page.StreamSummaries {
			r := resourceStream()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.StreamARN))
			d.Set("enforce_consumer_deletion", true)
			d.Set(names.AttrName, v.StreamName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Kinesis Streams (%s): %w", region, err)
	}

	return nil
}
