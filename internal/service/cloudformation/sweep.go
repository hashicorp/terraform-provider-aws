// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"golang.org/x/exp/slices"
)

func RegisterSweepers() {
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
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFormationConn(ctx)
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListStackSetsPagesWithContext(ctx, input, func(page *cloudformation.ListStackSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Summaries {
			input := &cloudformation.ListStackInstancesInput{
				StackSetName: v.StackSetName,
			}

			err := conn.ListStackInstancesPagesWithContext(ctx, input, func(page *cloudformation.ListStackInstancesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Summaries {
					stackSetID := aws.StringValue(v.StackSetId)

					if v.StackInstanceStatus != nil {
						if status := aws.StringValue(v.StackInstanceStatus.DetailedStatus); status == cloudformation.StackInstanceDetailedStatusSkippedSuspendedAccount {
							log.Printf("[INFO] Skipping CloudFormation StackSet Instance %s: DetailedStatus=%s", stackSetID, status)
							continue
						}
					}

					ouID := aws.StringValue(v.OrganizationalUnitId)
					accountOrOrgID := aws.StringValue(v.Account)
					if ouID != "" {
						accountOrOrgID = ouID
					}

					r := ResourceStackSetInstance()
					d := r.Data(nil)
					id := StackSetInstanceCreateResourceID(stackSetID, accountOrOrgID, aws.StringValue(v.Region))
					d.SetId(id)
					d.Set("call_as", cloudformation.CallAsSelf)
					if ouID != "" {
						d.Set("deployment_targets", []interface{}{map[string]interface{}{"organizational_unit_ids": schema.NewSet(schema.HashString, []interface{}{ouID})}})
					}

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFormation StackSet Instances (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation StackSet Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping CloudFormation StackSet Instances (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepStackSets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFormationConn(ctx)
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	// Attempt to determine whether or not Organizations access is enabled.
	orgAccessEnabled := false
	if servicePrincipalNames, err := tforganizations.FindEnabledServicePrincipalNames(ctx, client.OrganizationsConn(ctx)); err == nil {
		orgAccessEnabled = slices.Contains(servicePrincipalNames, "member.org.stacksets.cloudformation.amazonaws.com")
	}

	err = conn.ListStackSetsPagesWithContext(ctx, input, func(page *cloudformation.ListStackSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Summaries {
			name := aws.StringValue(v.StackSetName)

			if status := aws.StringValue(v.Status); status == cloudformation.StackSetStatusDeleted {
				log.Printf("[INFO] SkippingCloudFormation StackSet %s: Status=%s", name, status)
				continue
			}

			if permissionModel := aws.StringValue(v.PermissionModel); permissionModel == cloudformation.PermissionModelsServiceManaged && !orgAccessEnabled {
				log.Printf("[INFO] SkippingCloudFormation StackSet %s: PermissionModel=%s", name, permissionModel)
				continue
			}

			r := ResourceStackSet()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("call_as", cloudformation.CallAsSelf)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation StackSet sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFormation StackSets (%s): %w", region, err)
	}

	return nil
}

func sweepStacks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFormationConn(ctx)
	input := &cloudformation.ListStacksInput{
		StackStatusFilter: aws.StringSlice([]string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusImportComplete,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusUpdateComplete,
		}),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListStacksPagesWithContext(ctx, input, func(page *cloudformation.ListStacksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StackSummaries {
			name := aws.StringValue(v.StackName)
			inputU := &cloudformation.UpdateTerminationProtectionInput{
				EnableTerminationProtection: aws.Bool(false),
				StackName:                   aws.String(name),
			}

			log.Printf("[INFO] Disabling termination protection for CloudFormation Stack: %s", name)
			_, err := conn.UpdateTerminationProtectionWithContext(ctx, inputU)

			if err != nil {
				log.Printf("[ERROR] Disabling termination protection for CloudFormation Stack (%s): %s", name, err)
				continue
			}

			r := ResourceStack()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation Stack sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation Stacks: %s", err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFormation Stacks (%s): %w", region, err)
	}

	return nil
}
