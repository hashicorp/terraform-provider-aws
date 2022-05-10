//go:build sweep
// +build sweep

package batch

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).BatchConn
	iamconn := client.(*conns.AWSClient).IAMConn

	var sweeperErrs *multierror.Error

	input := &batch.DescribeComputeEnvironmentsInput{}
	r := ResourceComputeEnvironment()

	err = conn.DescribeComputeEnvironmentsPages(input, func(page *batch.DescribeComputeEnvironmentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, computeEnvironment := range page.ComputeEnvironments {
			name := aws.StringValue(computeEnvironment.ComputeEnvironmentName)

			d := r.Data(nil)
			d.SetId(name)
			d.Set("compute_environment_name", name)

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
			if aws.StringValue(computeEnvironment.Status) == batch.CEStatusInvalid {
				// Reusing the IAM Role name to prevent collisions and inventing a naming scheme
				serviceRoleARN, err := arn.Parse(aws.StringValue(computeEnvironment.ServiceRole))

				if err != nil {
					sweeperErr := fmt.Errorf("error parsing Batch Compute Environment (%s) Service Role ARN (%s): %w", name, aws.StringValue(computeEnvironment.ServiceRole), err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				servicePrincipal := fmt.Sprintf("%s.%s", batch.EndpointsID, sweep.PartitionDNSSuffix(region))
				serviceRoleName := strings.TrimPrefix(serviceRoleARN.Resource, "role/")
				serviceRolePolicyARN := arn.ARN{
					AccountID: "aws",
					Partition: sweep.Partition(region),
					Resource:  "policy/service-role/AWSBatchServiceRole",
					Service:   iam.ServiceName,
				}.String()

				iamCreateRoleInput := &iam.CreateRoleInput{
					AssumeRolePolicyDocument: aws.String(fmt.Sprintf("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\": \"%s\"},\"Action\":\"sts:AssumeRole\"}]}", servicePrincipal)),
					RoleName:                 aws.String(serviceRoleName),
				}

				_, err = iamconn.CreateRole(iamCreateRoleInput)

				if err != nil {
					sweeperErr := fmt.Errorf("error creating IAM Role (%s) for INVALID Batch Compute Environment (%s): %w", serviceRoleName, name, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				iamGetRoleInput := &iam.GetRoleInput{
					RoleName: aws.String(serviceRoleName),
				}

				err = iamconn.WaitUntilRoleExists(iamGetRoleInput)

				if err != nil {
					sweeperErr := fmt.Errorf("error waiting for IAM Role (%s) creation for INVALID Batch Compute Environment (%s): %w", serviceRoleName, name, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				iamAttachRolePolicyInput := &iam.AttachRolePolicyInput{
					PolicyArn: aws.String(serviceRolePolicyARN),
					RoleName:  aws.String(serviceRoleName),
				}

				_, err = iamconn.AttachRolePolicy(iamAttachRolePolicyInput)

				if err != nil {
					sweeperErr := fmt.Errorf("error attaching Batch IAM Policy (%s) to IAM Role (%s) for INVALID Batch Compute Environment (%s): %w", serviceRolePolicyARN, serviceRoleName, name, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Batch Compute Environment (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Compute Environment sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Batch Compute Environments: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepJobDefinitions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).BatchConn
	input := &batch.DescribeJobDefinitionsInput{
		Status: aws.String("ACTIVE"),
	}
	var sweeperErrs *multierror.Error

	err = conn.DescribeJobDefinitionsPages(input, func(page *batch.DescribeJobDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, jobDefinition := range page.JobDefinitions {
			arn := aws.StringValue(jobDefinition.JobDefinitionArn)

			log.Printf("[INFO] Deleting Batch Job Definition: %s", arn)
			_, err := conn.DeregisterJobDefinition(&batch.DeregisterJobDefinitionInput{
				JobDefinition: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Batch Job Definition (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Job Definitions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Batch Job Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepJobQueues(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).BatchConn

	out, err := conn.DescribeJobQueues(&batch.DescribeJobQueuesInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Batch Job Queue sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Batch Job Queues: %s", err)
	}
	for _, jobQueue := range out.JobQueues {
		name := jobQueue.JobQueueName

		log.Printf("[INFO] Disabling Batch Job Queue: %s", *name)
		err := DisableJobQueue(*name, conn)
		if err != nil {
			log.Printf("[ERROR] Failed to disable Batch Job Queue %s: %s", *name, err)
			continue
		}

		log.Printf("[INFO] Deleting Batch Job Queue: %s", *name)
		err = DeleteJobQueue(*name, conn)
		if err != nil {
			log.Printf("[ERROR] Failed to delete Batch Job Queue %s: %s", *name, err)
		}
	}

	return nil
}

func sweepSchedulingPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).BatchConn
	input := &batch.ListSchedulingPoliciesInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListSchedulingPoliciesPages(input, func(page *batch.ListSchedulingPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, schedulingPolicy := range page.SchedulingPolicies {
			arn := aws.StringValue(schedulingPolicy.Arn)

			log.Printf("[INFO] Deleting Batch Scheduling Policy: %s", arn)
			_, err := conn.DeleteSchedulingPolicy(&batch.DeleteSchedulingPolicyInput{
				Arn: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Batch Scheduling Policy (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Batch Scheduling Policies sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Batch Scheduling Policies: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
