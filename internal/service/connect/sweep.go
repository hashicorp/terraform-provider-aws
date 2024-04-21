// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_connect_instance", &resource.Sweeper{
		Name: "aws_connect_instance",
		F:    sweepInstance,
	})
}

func sweepInstance(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.ConnectClient(ctx)

	var errs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	// MaxResults:  Maximum value of 10. https://docs.aws.amazon.com/connect/latest/APIReference/API_ListInstances.html
	input := &connect.ListInstancesInput{MaxResults: aws.Int32(ListInstancesMaxResults)}

	pages := connect.NewListInstancesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing Connect Instances: %w", err))
		}

		for _, instanceSummary := range page.InstanceSummaryList {
			id := aws.ToString(instanceSummary.Id)

			log.Printf("[INFO] Deleting Connect Instance (%s)", id)
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Connect Instances for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Connect Instances sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
