// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_cloudformation_stack_set_instance", sweepStackSetInstances)
	awsv2.Register("aws_cloudformation_stack_set", sweepStackSets, "aws_cloudformation_stack_set_instance")
	awsv2.Register("aws_cloudformation_stack", sweepStacks, "aws_cloudformation_stack_set_instance")
}

func sweepStackSetInstances(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CloudFormationClient(ctx)
	input := cloudformation.ListStackSetsInput{
		Status: awstypes.StackSetStatusActive,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudformation.NewListStackSetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Summaries {
			input := cloudformation.ListStackInstancesInput{
				StackSetName: v.StackSetName,
			}

			pages := cloudformation.NewListStackInstancesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.Summaries {
					stackSetID := aws.ToString(v.StackSetId)

					if v.StackInstanceStatus != nil && v.StackInstanceStatus.DetailedStatus == awstypes.StackInstanceDetailedStatusSkippedSuspendedAccount {
						log.Printf("[INFO] Skipping CloudFormation StackSet Instance %s: DetailedStatus=%s", stackSetID, v.StackInstanceStatus.DetailedStatus)
						continue
					}

					ouID := aws.ToString(v.OrganizationalUnitId)
					accountOrOrgID := aws.ToString(v.Account)
					if ouID != "" {
						accountOrOrgID = ouID
					}

					r := resourceStackSetInstance()
					d := r.Data(nil)
					id, _ := flex.FlattenResourceId([]string{stackSetID, accountOrOrgID, aws.ToString(v.Region)}, stackSetInstanceResourceIDPartCount, false)
					d.SetId(id)
					d.Set("call_as", awstypes.CallAsSelf)
					if ouID != "" {
						d.Set("deployment_targets", []any{map[string]any{"organizational_unit_ids": schema.NewSet(schema.HashString, []any{ouID})}})
					}

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepStackSets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CloudFormationClient(ctx)
	input := cloudformation.ListStackSetsInput{
		Status: awstypes.StackSetStatusActive,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	// Attempt to determine whether or not Organizations access is enabled.
	orgAccessEnabled := false
	if servicePrincipalNames, err := tforganizations.FindEnabledServicePrincipalNames(ctx, client.OrganizationsClient(ctx)); err == nil {
		orgAccessEnabled = slices.Contains(servicePrincipalNames, "member.org.stacksets.cloudformation.amazonaws.com")
	}

	pages := cloudformation.NewListStackSetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Summaries {
			name := aws.ToString(v.StackSetName)

			if status := v.Status; status == awstypes.StackSetStatusDeleted {
				log.Printf("[INFO] SkippingCloudFormation StackSet %s: Status=%s", name, status)
				continue
			}

			if permissionModel := v.PermissionModel; permissionModel == awstypes.PermissionModelsServiceManaged && !orgAccessEnabled {
				log.Printf("[INFO] SkippingCloudFormation StackSet %s: PermissionModel=%s", name, permissionModel)
				continue
			}

			r := resourceStackSet()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("call_as", awstypes.CallAsSelf)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepStacks(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CloudFormationClient(ctx)
	input := cloudformation.ListStacksInput{
		StackStatusFilter: []awstypes.StackStatus{
			awstypes.StackStatusCreateComplete,
			awstypes.StackStatusImportComplete,
			awstypes.StackStatusRollbackComplete,
			awstypes.StackStatusUpdateComplete,
		},
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudformation.NewListStacksPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.StackSummaries {
			name := aws.ToString(v.StackName)
			input := cloudformation.UpdateTerminationProtectionInput{
				EnableTerminationProtection: aws.Bool(false),
				StackName:                   aws.String(name),
			}

			log.Printf("[INFO] Disabling termination protection for CloudFormation Stack: %s", name)
			_, err := conn.UpdateTerminationProtection(ctx, &input)

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

	return sweepResources, nil
}
