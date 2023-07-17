// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package kinesis

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.KinesisConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	input := &kinesis.ListStreamsInput{}
	err = conn.ListStreamsPagesWithContext(ctx, input, func(page *kinesis.ListStreamsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StreamSummaries {
			r := ResourceStream()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.StreamARN))
			d.Set("enforce_consumer_deletion", true)
			d.Set("name", v.StreamName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Stream sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing Kinesis Streams (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Kinesis Streams (%s): %w", region, err)
	}

	return nil
}
