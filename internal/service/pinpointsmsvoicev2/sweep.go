// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpointsmsvoicev2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_pinpointsmsvoicev2_phone_number", &resource.Sweeper{
		Name: "aws_pinpointsmsvoicev2_phone_number",
		F:    sweepPhoneNumbers,
	})
}

func sweepPhoneNumbers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &pinpointsmsvoicev2.DescribePhoneNumbersInput{}
	conn := client.PinpointSMSVoiceV2Client(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := pinpointsmsvoicev2.NewDescribePhoneNumbersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping End User Messaging SMS Phone Number sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing End User Messaging SMS Phone Numbers (%s): %w", region, err)
		}

		for _, v := range page.PhoneNumbers {
			id := aws.ToString(v.PhoneNumberId)

			if v := v.DeletionProtectionEnabled; v {
				log.Printf("[INFO] Skipping End User Messaging SMS Phone Number %s: DeletionProtectionEnabled=%t", id, v)
				continue
			}

			sweepResources = append(sweepResources, framework.NewSweepResource(newPhoneNumberResource, client,
				framework.NewAttribute(names.AttrID, id)))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping End User Messaging SMS Phone Numbers (%s): %w", region, err)
	}

	return nil
}
