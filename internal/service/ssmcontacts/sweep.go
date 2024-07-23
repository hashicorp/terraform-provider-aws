// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmcontacts

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmcontacts"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_ssmcontacts_rotation", &resource.Sweeper{
		Name: "aws_ssmcontacts_rotation",
		F:    sweepRotations,
	})
}

func sweepRotations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.SSMContactsClient(ctx)
	input := &ssmcontacts.ListRotationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ssmcontacts.NewListRotationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SSMContacts Rotations sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving SSMContacts Rotations: %w", err)
		}

		for _, v := range page.Rotations {
			id := aws.ToString(v.RotationArn)

			log.Printf("[INFO] Deleting SSMContacts Rotation: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceRotation, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SSMContacts Rotations for %s: %w", region, err)
	}

	return nil
}
