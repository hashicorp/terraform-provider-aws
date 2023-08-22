// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package signer

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_signer_signing_profile", &resource.Sweeper{
		Name: "aws_signer_signing_profile",
		F:    sweepSigningProfiles,
	})
}

func sweepSigningProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.SignerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	input := &signer.ListSigningProfilesInput{}

	pages := signer.NewListSigningProfilesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Signer Signing Profiles sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Signer Signing Profiles: %w", err)
		}

		for _, profile := range page.Profiles {
			name := aws.ToString(profile.ProfileName)

			r := ResourceSigningProfile()
			d := r.Data(nil)
			d.SetId(name)

			log.Printf("[INFO] Deleting Signer Signing Profile: %s", name)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Signer Signing Profiles for %s: %w", region, err)
	}

	return nil
}
