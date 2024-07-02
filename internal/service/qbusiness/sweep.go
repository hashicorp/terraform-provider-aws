// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_qbusiness_app", &resource.Sweeper{
		Name: "aws_qbusiness_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.QBusinessClient(ctx)

	input := &qbusiness.ListApplicationsInput{}

	for {
		output, err := conn.ListApplications(ctx, input)
		if err != nil {
			if awsv1.SkipSweepError(err) {
				log.Printf("[WARN] Skipping QBusiness app sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("error retrieving Qbusiness apps: %s", err)
		}

		if len(output.Applications) == 0 {
			log.Print("[DEBUG] No QBusiness apps to sweep")
			return nil
		}

		for _, item := range output.Applications {
			name := item.DisplayName

			log.Printf("[INFO] Deleting QBusiness app %s", aws.ToString(name))
			_, err := conn.DeleteApplication(ctx, &qbusiness.DeleteApplicationInput{
				ApplicationId: item.ApplicationId,
			})
			if err != nil {
				return fmt.Errorf("error deleting QBusiness app %s: %s", aws.ToString(name), err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}
