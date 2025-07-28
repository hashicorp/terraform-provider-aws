// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_chime_voice_connector", &resource.Sweeper{
		Name: "aws_chime_voice_connector",
		F:    sweepVoiceConnectors,
	})
}

func sweepVoiceConnectors(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.ChimeSDKVoiceClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	in := &chimesdkvoice.ListVoiceConnectorsInput{}

	pages := chimesdkvoice.NewListVoiceConnectorsPaginator(conn, in)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Chime Voice Connector sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Chime Voice Connectors: %w", err)
		}

		for _, vc := range page.VoiceConnectors {
			id := aws.ToString(vc.VoiceConnectorId)

			r := ResourceVoiceConnector()
			d := r.Data(nil)
			d.SetId(id)

			log.Printf("[INFO] Deleting Chime Voice Connector: %s", id)
			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Chime Voice Connectors for %s: %w", region, err)
	}

	return nil
}
