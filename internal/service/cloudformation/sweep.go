// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"fmt"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	conn := client.CloudFormationClient(ctx)
	input := &cloudformation.ListStackSetsInput{
		Status: awstypes.StackSetStatusActive,
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudformation.NewListStackSetsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFormation StackSet Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
		}

		for _, v := range page.Summaries {
			input := &cloudformation.ListStackInstancesInput{
				StackSetName: v.StackSetName,
			}

			pages := cloudformation.NewListStackInstancesPaginator(conn, input)

			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFormation StackSet Instances (%s): %w", region, err))
				}

				for _, v := range page.Summaries {
					stackSetID := aws.ToString(v.StackSetId)

					if v.StackInstanceStatus != nil {
						if v.StackInstanceStatus.DetailedStatus == awstypes.StackInstanceDetailedStatusSkippedSuspendedAccount {
							log.Printf("[INFO] Skipping CloudFormation StackSet Instance %s: DetailedStatus=%s", stackSetID, v.StackInstanceStatus.DetailedStatus)
							continue
						}
					}

					ouID := aws.ToString(v.OrganizationalUnitId)
					accountOrOrgID := aws.ToString(v.Account)
					if ouID != "" {
						accountOrOrgID = ouID
					}

					r := resourceStackSetInstance()
					d := r.Data(nil)
					id := errs.Must(flex.FlattenResourceId([]string{stackSetID, accountOrOrgID, aws.ToString(v.Region)}, stackSetInstanceResourceIDPartCount, false))
					d.SetId(id)
					d.Set("call_as", awstypes.CallAsSelf)
					if ouID != "" {
						d.Set("deployment_targets", []interface{}{map[string]interface{}{"organizational_unit_ids": schema.NewSet(schema.HashString, []interface{}{ouID})}})
					}

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
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
	conn := client.CloudFormationClient(ctx)
	input := &cloudformation.ListStackSetsInput{
		Status: awstypes.StackSetStatusActive,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	// Attempt to determine whether or not Organizations access is enabled.
	orgAccessEnabled := false
	if servicePrincipalNames, err := tforganizations.FindEnabledServicePrincipalNames(ctx, client.OrganizationsClient(ctx)); err == nil {
		orgAccessEnabled = slices.Contains(servicePrincipalNames, "member.org.stacksets.cloudformation.amazonaws.com")
	}

	pages := cloudformation.NewListStackSetsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFormation StackSet sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
		}

		for _, v := range page.Summaries {
			name := aws.ToString(v.StackSetName)

			if v.Status == awstypes.StackSetStatusDeleted {
				log.Printf("[INFO] SkippingCloudFormation StackSet %s: Status=%s", name, v.Status)
				continue
			}

			if v.PermissionModel == awstypes.PermissionModelsServiceManaged && !orgAccessEnabled {
				log.Printf("[INFO] SkippingCloudFormation StackSet %s: PermissionModel=%s", name, v.PermissionModel)
				continue
			}

			r := resourceStackSet()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("call_as", awstypes.CallAsSelf)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
	conn := client.CloudFormationClient(ctx)
	input := &cloudformation.ListStacksInput{
		StackStatusFilter: []awstypes.StackStatus{
			awstypes.StackStatusCreateComplete,
			awstypes.StackStatusImportComplete,
			awstypes.StackStatusRollbackComplete,
			awstypes.StackStatusUpdateComplete,
		},
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudformation.NewListStacksPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFormation Stack sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudFormation Stacks: %s", err)
		}

		for _, v := range page.StackSummaries {
			name := aws.ToString(v.StackName)
			inputU := &cloudformation.UpdateTerminationProtectionInput{
				EnableTerminationProtection: aws.Bool(false),
				StackName:                   aws.String(name),
			}

			log.Printf("[INFO] Disabling termination protection for CloudFormation Stack: %s", name)
			_, err := conn.UpdateTerminationProtection(ctx, inputU)

			if err != nil {
				log.Printf("[ERROR] Disabling termination protection for CloudFormation Stack (%s): %s", name, err)
				continue
			}

			r := resourceStack()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFormation Stacks (%s): %w", region, err)
	}

	return nil
}
