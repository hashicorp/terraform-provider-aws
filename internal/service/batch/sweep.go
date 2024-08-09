// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/batch"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const propagationTimeout = 2 * time.Minute

func RegisterSweepers() {
	resource.AddTestSweepers("aws_batch_compute_environment", &resource.Sweeper{
		Name: "aws_batch_compute_environment",
		Dependencies: []string{
			"aws_batch_job_queue",
		},
		F: sweepComputeEnvironments,
	})

	resource.AddTestSweepers("aws_batch_job_definition", &resource.Sweeper{
		Name: "aws_batch_job_definition",
		F:    sweepJobDefinitions,
		Dependencies: []string{
			"aws_batch_job_queue",
		},
	})

	resource.AddTestSweepers("aws_batch_job_queue", &resource.Sweeper{
		Name: "aws_batch_job_queue",
		F:    sweepJobQueues,
	})

	resource.AddTestSweepers("aws_batch_scheduling_policy", &resource.Sweeper{
		Name: "aws_batch_scheduling_policy",
		F:    sweepSchedulingPolicies,
		Dependencies: []string{
			"aws_batch_job_queue",
		},
	})
}

func sweepComputeEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.BatchConn(ctx)
	iamconn := client.IAMClient(ctx)

	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &batch.DescribeComputeEnvironmentsInput{}
	err = conn.DescribeComputeEnvironmentsPagesWithContext(ctx, input, func(page *batch.DescribeComputeEnvironmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ComputeEnvironments {
			name := aws.StringValue(v.ComputeEnvironmentName)

			// Reference: https://aws.amazon.com/premiumsupport/knowledge-center/batch-invalid-compute-environment/
			//
			// When a Compute Environment becomes INVALID, it is typically because the associated
			// IAM Role has disappeared. There is no automatic resolution via the API, except to
			// associate a new IAM Role that is valid, then delete the Compute Environment.
			//
			// We avoid doing this in the resource because it would be very unexpected behavior
			// for the resource and this issue should be fixed in the API (e.g. Service Linked Role).
			//
			// To save writing much more logic around IAM Role deletion, we allow the
			// aws_iam_role sweeper to handle cleaning these up.
			if aws.StringValue(v.Status) == batch.CEStatusInvalid {
				// Reusing the IAM Role name to prevent collisions and inventing a naming scheme.
				serviceRoleARN, err := arn.Parse(aws.StringValue(v.ServiceRole))

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error parsing Batch Compute Environment (%s) Service Role ARN (%s): %w", name, aws.StringValue(v.ServiceRole), err))
					continue
				}

				servicePrincipal := fmt.Sprintf("%s.%s", batch.EndpointsID, sweep.PartitionDNSSuffix(region))
				serviceRoleName := strings.TrimPrefix(serviceRoleARN.Resource, "role/")
				serviceRolePolicyARN := arn.ARN{
					AccountID: "aws",
					Partition: sweep.Partition(region),
					Resource:  "policy/service-role/AWSBatchServiceRole",
					Service:   "iam",
				}.String()

				iamCreateRoleInput := &iam.CreateRoleInput{
					AssumeRolePolicyDocument: aws.String(fmt.Sprintf("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\": \"%s\"},\"Action\":\"sts:AssumeRole\"}]}", servicePrincipal)),
					RoleName:                 aws.String(serviceRoleName),
				}

				_, err = iamconn.CreateRole(ctx, iamCreateRoleInput)

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error creating IAM Role (%s) for INVALID Batch Compute Environment (%s): %w", serviceRoleName, name, err))
					continue
				}

				iamGetRoleInput := &iam.GetRoleInput{
					RoleName: aws.String(serviceRoleName),
				}

				waiter := iam.NewRoleExistsWaiter(iamconn)
				err = waiter.Wait(ctx, iamGetRoleInput, propagationTimeout)

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error waiting for IAM Role (%s) creation for INVALID Batch Compute Environment (%s): %w", serviceRoleName, name, err))
					continue
				}

				iamAttachRolePolicyInput := &iam.AttachRolePolicyInput{
					PolicyArn: aws.String(serviceRolePolicyARN),
					RoleName:  aws.String(serviceRoleName),
				}

				_, err = iamconn.AttachRolePolicy(ctx, iamAttachRolePolicyInput)

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error attaching Batch IAM Policy (%s) to IAM Role (%s) for INVALID Batch Compute Environment (%s): %w", serviceRolePolicyARN, serviceRoleName, name, err))
					continue
				}
			}

			r := ResourceComputeEnvironment()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Compute Environment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Batch Compute Environments (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Batch Compute Environments (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepJobDefinitions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &batch.DescribeJobDefinitionsInput{
		Status: aws.String("ACTIVE"),
	}
	conn := client.BatchConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeJobDefinitionsPagesWithContext(ctx, input, func(page *batch.DescribeJobDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.JobDefinitions {
			r := ResourceJobDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.JobDefinitionArn))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Job Definition sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Batch Job Definitions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Batch Job Definitions (%s): %w", region, err)
	}

	return nil
}

func sweepJobQueues(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &batch.DescribeJobQueuesInput{}
	conn := client.BatchConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeJobQueuesPagesWithContext(ctx, input, func(page *batch.DescribeJobQueuesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.JobQueues {
			id := aws.StringValue(v.JobQueueArn)

			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceJobQueue, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Job Queue sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Batch Job Queues (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Batch Job Queues (%s): %w", region, err)
	}

	return nil
}

func sweepSchedulingPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &batch.ListSchedulingPoliciesInput{}
	conn := client.BatchConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListSchedulingPoliciesPagesWithContext(ctx, input, func(page *batch.ListSchedulingPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SchedulingPolicies {
			r := ResourceSchedulingPolicy()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Arn))

			sweepResources = append(sweepResources, sdk.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Scheduling Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Batch Scheduling Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Batch Scheduling Policies (%s): %w", region, err)
	}

	return nil
}
