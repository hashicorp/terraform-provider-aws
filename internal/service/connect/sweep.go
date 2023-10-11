// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package connect

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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

	conn := client.ConnectConn(ctx)

	var errs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	// MaxResults:  Maximum value of 10. https://docs.aws.amazon.com/connect/latest/APIReference/API_ListInstances.html
	input := &connect.ListInstancesInput{MaxResults: aws.Int64(ListInstancesMaxResults)}

	err = conn.ListInstancesPagesWithContext(ctx, input, func(page *connect.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, instanceSummary := range page.InstanceSummaryList {
			if instanceSummary == nil {
				continue
			}

			id := aws.StringValue(instanceSummary.Id)

			log.Printf("[INFO] Deleting Connect Instance (%s)", id)
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Connect Instances: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Connect Instances for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Connect Instances sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
