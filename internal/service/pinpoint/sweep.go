// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_pinpoint_app", &resource.Sweeper{
		Name: "aws_pinpoint_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.PinpointClient(ctx)

	input := &pinpoint.GetAppsInput{}

	for {
		output, err := conn.GetApps(ctx, input)
		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Pinpoint app sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Pinpoint apps: %s", err)
		}

		if len(output.ApplicationsResponse.Item) == 0 {
			log.Print("[DEBUG] No Pinpoint apps to sweep")
			return nil
		}

		for _, item := range output.ApplicationsResponse.Item {
			name := aws.ToString(item.Name)

			log.Printf("[INFO] Deleting Pinpoint app %s", name)
			_, err := conn.DeleteApp(ctx, &pinpoint.DeleteAppInput{
				ApplicationId: item.Id,
			})
			if err != nil {
				return fmt.Errorf("Error deleting Pinpoint app %s: %s", name, err)
			}
		}

		if output.ApplicationsResponse.NextToken == nil {
			break
		}
		input.Token = output.ApplicationsResponse.NextToken
	}

	return nil
}
