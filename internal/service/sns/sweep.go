//go:build sweep
// +build sweep

package sns

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_sns_platform_application", &resource.Sweeper{
		Name: "aws_sns_platform_application",
		F:    sweepPlatformApplications,
	})

	resource.AddTestSweepers("aws_sns_topic", &resource.Sweeper{
		Name: "aws_sns_topic",
		F:    sweepTopics,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_backup_vault_notifications",
			"aws_budgets_budget",
			"aws_config_delivery_channel",
			"aws_dax_cluster",
			"aws_db_event_subscription",
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
			"aws_glacier_vault",
			"aws_iot_topic_rule",
			"aws_neptune_event_subscription",
			"aws_redshift_event_subscription",
			"aws_s3_bucket",
			"aws_ses_configuration_set",
			"aws_ses_domain_identity",
			"aws_ses_email_identity",
			"aws_ses_receipt_rule_set",
			"aws_sns_platform_application",
		},
	})
}

func sweepPlatformApplications(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SNSConn
	var sweeperErrs *multierror.Error

	err = conn.ListPlatformApplicationsPages(&sns.ListPlatformApplicationsInput{}, func(page *sns.ListPlatformApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, platformApplication := range page.PlatformApplications {
			arn := aws.StringValue(platformApplication.PlatformApplicationArn)

			log.Printf("[INFO] Deleting SNS Platform Application: %s", arn)
			_, err := conn.DeletePlatformApplication(&sns.DeletePlatformApplicationInput{
				PlatformApplicationArn: aws.String(arn),
			})
			if tfawserr.ErrMessageContains(err, sns.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SNS Platform Application (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SNS Platform Applications sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SNS Platform Applications: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTopics(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SNSConn
	var sweeperErrs *multierror.Error

	err = conn.ListTopicsPages(&sns.ListTopicsInput{}, func(page *sns.ListTopicsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, topic := range page.Topics {
			arn := aws.StringValue(topic.TopicArn)

			log.Printf("[INFO] Deleting SNS Topic: %s", arn)
			_, err := conn.DeleteTopic(&sns.DeleteTopicInput{
				TopicArn: aws.String(arn),
			})
			if tfawserr.ErrMessageContains(err, sns.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting SNS Topic (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SNS Topics sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SNS Topics: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
