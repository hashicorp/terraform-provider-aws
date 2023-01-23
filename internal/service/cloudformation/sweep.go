//go:build sweep
// +build sweep

package cloudformation

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudformation_stack_set_instance", &resource.Sweeper{
		Name: "aws_cloudformation_stack_set_instance",
		F:    sweepStackSetInstances,
	})

	resource.AddTestSweepers("aws_cloudformation_stack_set", &resource.Sweeper{
		Name: "aws_cloudformation_stack_set",
		Dependencies: []string{
			"aws_cloudformation_stack_set_instance",
		},
		F: sweepStackSets,
	})

	resource.AddTestSweepers("aws_cloudformation_stack", &resource.Sweeper{
		Name: "aws_cloudformation_stack",
		Dependencies: []string{
			"aws_cloudformation_stack_set_instance",
		},
		F: sweepStacks,
	})
}

func sweepStackSetInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFormationConn()
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListStackSetsPagesWithContext(ctx, input, func(page *cloudformation.ListStackSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, summary := range page.Summaries {
			input := &cloudformation.ListStackInstancesInput{
				StackSetName: summary.StackSetName,
			}

			err = conn.ListStackInstancesPagesWithContext(ctx, input, func(page *cloudformation.ListStackInstancesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, summary := range page.Summaries {
					r := ResourceStackSetInstance()
					d := r.Data(nil)
					id := StackSetInstanceCreateResourceID(
						aws.StringValue(summary.StackSetId),
						aws.StringValue(summary.Account),
						aws.StringValue(summary.Region),
					)
					d.SetId(id)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFormation StackSet Instances (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation StackSet Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping CloudFormation StackSet Instances (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepStackSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFormationConn()
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListStackSetsPagesWithContext(ctx, input, func(page *cloudformation.ListStackSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, summary := range page.Summaries {
			r := ResourceStackSet()
			d := r.Data(nil)
			d.SetId(aws.StringValue(summary.StackSetName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation StackSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFormation StackSets (%s): %w", region, err)
	}

	return nil
}

func sweepStacks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).CloudFormationConn()
	input := &cloudformation.ListStacksInput{
		StackStatusFilter: aws.StringSlice([]string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusImportComplete,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusUpdateComplete,
		}),
	}
	var sweeperErrs *multierror.Error

	err = conn.ListStacksPagesWithContext(ctx, input, func(page *cloudformation.ListStacksOutput, lastPage bool) bool {
		for _, stack := range page.StackSummaries {
			name := aws.StringValue(stack.StackName)

			updateTerminationProtectionInput := &cloudformation.UpdateTerminationProtectionInput{
				EnableTerminationProtection: aws.Bool(false),
				StackName:                   stack.StackName,
			}

			log.Printf("[INFO] Disabling termination protection for CloudFormation Stack: %s", name)
			_, err := conn.UpdateTerminationProtectionWithContext(ctx, updateTerminationProtectionInput)

			if err != nil {
				sweeperErr := fmt.Errorf("error disabling termination protection for CloudFormation Stack (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			input := &cloudformation.DeleteStackInput{
				StackName: stack.StackName,
			}

			log.Printf("[INFO] Deleting CloudFormation Stack: %s", name)
			_, err = conn.DeleteStackWithContext(ctx, input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudFormation Stack (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation Stack sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation Stacks: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}
