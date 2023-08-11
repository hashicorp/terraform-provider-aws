// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package dlm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_dlm_lifecycle_policy", &resource.Sweeper{
		Name: "aws_dlm_lifecycle_policy",
		F:    sweepLifecyclePolicies,
	})

}

func sweepLifecyclePolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.DLMConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &dlm.GetLifecyclePoliciesInput{}
	policies, err := conn.GetLifecyclePoliciesWithContext(ctx, input)
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing DLM Lifecycle Policy for %s: %w", region, err))
	}

	for _, lifecyclePolicy := range policies.Policies {
		r := ResourceLifecyclePolicy()
		d := r.Data(nil)

		id := aws.StringValue(lifecyclePolicy.PolicyId)
		d.SetId(id)

		if err != nil {
			err := fmt.Errorf("error reading DLM Lifecycle Policy (%s): %w", id, err)
			log.Printf("[ERROR] %s", err)
			errs = multierror.Append(errs, err)
			continue
		}

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DLM Lifecycle Policy for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DLM Lifecycle Policy sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
