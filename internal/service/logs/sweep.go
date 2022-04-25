//go:build sweep
// +build sweep

package logs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_log_group", &resource.Sweeper{
		Name: "aws_cloudwatch_log_group",
		F:    sweepGroups,
		Dependencies: []string{
			"aws_api_gateway_rest_api",
			"aws_cloudhsm_v2_cluster",
			"aws_cloudtrail",
			"aws_datasync_task",
			"aws_db_instance",
			"aws_directory_service_directory",
			"aws_ec2_client_vpn_endpoint",
			"aws_eks_cluster",
			"aws_elasticsearch_domain",
			"aws_flow_log",
			"aws_glue_job",
			"aws_kinesis_analytics_application",
			"aws_kinesis_firehose_delivery_stream",
			"aws_lambda_function",
			"aws_mq_broker",
			"aws_msk_cluster",
			"aws_rds_cluster",
			"aws_route53_query_log",
			"aws_sagemaker_endpoint",
			"aws_storagegateway_gateway",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_query_definition", &resource.Sweeper{
		Name: "aws_cloudwatch_query_definition",
		F:    sweeplogQueryDefinitions,
	})

	resource.AddTestSweepers("aws_cloudwatch_log_resource_policy", &resource.Sweeper{
		Name: "aws_cloudwatch_log_resource_policy",
		F:    sweepResourcePolicies,
	})
}

func sweepGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LogsConn
	var sweeperErrs *multierror.Error

	input := &cloudwatchlogs.DescribeLogGroupsInput{}

	err = conn.DescribeLogGroupsPages(input, func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, logGroup := range page.LogGroups {
			if logGroup == nil {
				continue
			}

			input := &cloudwatchlogs.DeleteLogGroupInput{
				LogGroupName: logGroup.LogGroupName,
			}
			name := aws.StringValue(logGroup.LogGroupName)

			log.Printf("[INFO] Deleting CloudWatch Log Group: %s", name)
			_, err := conn.DeleteLogGroup(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudWatch Log Group (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Log Groups sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudWatch Log Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweeplogQueryDefinitions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).LogsConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{}

	// AWS SDK Go does not currently provide paginator
	for {
		output, err := conn.DescribeQueryDefinitions(input)

		if err != nil {
			err := fmt.Errorf("error reading CloudWatch Log Query Definition: %w", err)
			log.Printf("[ERROR] %s", err)
			errs = multierror.Append(errs, err)
			break
		}

		for _, queryDefinition := range output.QueryDefinitions {
			r := ResourceQueryDefinition()
			d := r.Data(nil)

			d.SetId(aws.StringValue(queryDefinition.QueryDefinitionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping CloudWatch Log Query Definition for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping CloudWatch Log Query Definition sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepResourcePolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LogsConn

	input := &cloudwatchlogs.DescribeResourcePoliciesInput{}

	for {
		output, err := conn.DescribeResourcePolicies(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudWatchLog Resource Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error describing CloudWatchLog Resource Policy: %s", err)
		}

		for _, resourcePolicy := range output.ResourcePolicies {
			policyName := aws.StringValue(resourcePolicy.PolicyName)
			deleteInput := &cloudwatchlogs.DeleteResourcePolicyInput{
				PolicyName: resourcePolicy.PolicyName,
			}

			log.Printf("[INFO] Deleting CloudWatch Log Resource Policy: %s", policyName)

			if _, err := conn.DeleteResourcePolicy(deleteInput); err != nil {
				return fmt.Errorf("error deleting CloudWatch log resource policy (%s): %s", policyName, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}
