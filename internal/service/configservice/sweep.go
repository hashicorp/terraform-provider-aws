//go:build sweep
// +build sweep

package configservice

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_config_aggregate_authorization", &resource.Sweeper{
		Name: "aws_config_aggregate_authorization",
		F:    sweepAggregateAuthorizations,
	})

	resource.AddTestSweepers("aws_config_configuration_aggregator", &resource.Sweeper{
		Name: "aws_config_configuration_aggregator",
		F:    sweepConfigurationAggregators,
	})

	resource.AddTestSweepers("aws_config_configuration_recorder", &resource.Sweeper{
		Name: "aws_config_configuration_recorder",
		F:    sweepConfigurationRecorder,
	})

	resource.AddTestSweepers("aws_config_delivery_channel", &resource.Sweeper{
		Name: "aws_config_delivery_channel",
		Dependencies: []string{
			"aws_config_configuration_recorder",
		},
		F: sweepDeliveryChannels,
	})
}

func sweepAggregateAuthorizations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ConfigServiceConn()

	aggregateAuthorizations, err := DescribeAggregateAuthorizations(ctx, conn)
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Aggregate Authorizations sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving config aggregate authorizations: %s", err)
	}

	if len(aggregateAuthorizations) == 0 {
		log.Print("[DEBUG] No config aggregate authorizations to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d config aggregate authorizations", len(aggregateAuthorizations))

	for _, auth := range aggregateAuthorizations {
		log.Printf("[INFO] Deleting config authorization %s", *auth.AggregationAuthorizationArn)
		_, err := conn.DeleteAggregationAuthorizationWithContext(ctx, &configservice.DeleteAggregationAuthorizationInput{
			AuthorizedAccountId: auth.AuthorizedAccountId,
			AuthorizedAwsRegion: auth.AuthorizedAwsRegion,
		})
		if err != nil {
			return fmt.Errorf("Error deleting config aggregate authorization %s: %s", *auth.AggregationAuthorizationArn, err)
		}
	}

	return nil
}

func sweepConfigurationAggregators(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ConfigServiceConn()

	resp, err := conn.DescribeConfigurationAggregatorsWithContext(ctx, &configservice.DescribeConfigurationAggregatorsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Configuration Aggregators sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving config configuration aggregators: %s", err)
	}

	if len(resp.ConfigurationAggregators) == 0 {
		log.Print("[DEBUG] No config configuration aggregators to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d config configuration aggregators", len(resp.ConfigurationAggregators))

	for _, agg := range resp.ConfigurationAggregators {
		log.Printf("[INFO] Deleting config configuration aggregator %s", *agg.ConfigurationAggregatorName)
		_, err := conn.DeleteConfigurationAggregatorWithContext(ctx, &configservice.DeleteConfigurationAggregatorInput{
			ConfigurationAggregatorName: agg.ConfigurationAggregatorName,
		})

		if err != nil {
			return fmt.Errorf("error deleting config configuration aggregator %s: %w",
				aws.StringValue(agg.ConfigurationAggregatorName), err)
		}
	}

	return nil
}

func sweepConfigurationRecorder(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ConfigServiceConn()

	req := &configservice.DescribeConfigurationRecordersInput{}
	resp, err := conn.DescribeConfigurationRecordersWithContext(ctx, req)
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Configuration Recorders sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Configuration Recorders: %s", err)
	}

	if len(resp.ConfigurationRecorders) == 0 {
		log.Print("[DEBUG] No AWS Config Configuration Recorder to sweep")
		return nil
	}

	for _, cr := range resp.ConfigurationRecorders {
		_, err := conn.StopConfigurationRecorderWithContext(ctx, &configservice.StopConfigurationRecorderInput{
			ConfigurationRecorderName: cr.Name,
		})
		if err != nil {
			return err
		}

		_, err = conn.DeleteConfigurationRecorderWithContext(ctx, &configservice.DeleteConfigurationRecorderInput{
			ConfigurationRecorderName: cr.Name,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting Configuration Recorder (%s): %s",
				*cr.Name, err)
		}
	}

	return nil
}

func sweepDeliveryChannels(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ConfigServiceConn()

	req := &configservice.DescribeDeliveryChannelsInput{}
	var resp *configservice.DescribeDeliveryChannelsOutput
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.DescribeDeliveryChannelsWithContext(ctx, req)
		if err != nil {
			// ThrottlingException: Rate exceeded
			if tfawserr.ErrMessageContains(err, "ThrottlingException", "Rate exceeded") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Delivery Channels sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Delivery Channels: %s", err)
	}

	if len(resp.DeliveryChannels) == 0 {
		log.Print("[DEBUG] No AWS Config Delivery Channel to sweep")
		return nil
	}

	for _, dc := range resp.DeliveryChannels {
		_, err := conn.DeleteDeliveryChannelWithContext(ctx, &configservice.DeleteDeliveryChannelInput{
			DeliveryChannelName: dc.Name,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting Delivery Channel (%s): %s",
				*dc.Name, err)
		}
	}

	return nil
}
