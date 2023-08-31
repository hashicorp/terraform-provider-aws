// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package mediapackage

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_media_package_channel", &resource.Sweeper{
		Name: "aws_media_package_channel",
		F:    sweepChannels,
	})
}

func sweepChannels(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.MediaPackageClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &mediapackage.ListChannelsInput{}

	pages := mediapackage.NewListChannelsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Println("[WARN] Skipping MediaPackage Channels sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("retrieving MediaPackage Channels: %w", err)
		}

		for _, channel := range page.Channels {
			id := aws.ToString(channel.Id)
			log.Printf("[INFO] Deleting MediaChannel Channel: %s", id)

			r := ResourceChannel()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("sweeping Mediapackage Channels for %s: %w", region, err)
	}

	return nil
}
