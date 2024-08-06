// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_config_aggregate_authorization", &resource.Sweeper{
		Name: "aws_config_aggregate_authorization",
		F:    sweepAggregateAuthorizations,
	})

	resource.AddTestSweepers("aws_config_config_rule", &resource.Sweeper{
		Name: "aws_config_config_rule",
		F:    sweepConfigRules,
		Dependencies: []string{
			"aws_config_remediation_configuration",
		},
	})

	resource.AddTestSweepers("aws_config_configuration_aggregator", &resource.Sweeper{
		Name: "aws_config_configuration_aggregator",
		F:    sweepConfigurationAggregators,
	})

	resource.AddTestSweepers("aws_config_configuration_recorder", &resource.Sweeper{
		Name: "aws_config_configuration_recorder",
		F:    sweepConfigurationRecorder,
	})

	resource.AddTestSweepers("aws_config_conformance_pack", &resource.Sweeper{
		Name: "aws_config_conformance_pack",
		F:    sweepConformancePacks,
	})

	resource.AddTestSweepers("aws_config_delivery_channel", &resource.Sweeper{
		Name: "aws_config_delivery_channel",
		F:    sweepDeliveryChannels,
		Dependencies: []string{
			"aws_config_configuration_recorder",
		},
	})

	resource.AddTestSweepers("aws_config_remediation_configuration", &resource.Sweeper{
		Name: "aws_config_remediation_configuration",
		F:    sweepRemediationConfigurations,
	})
}

func sweepAggregateAuthorizations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ConfigServiceClient(ctx)
	input := &configservice.DescribeAggregationAuthorizationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := configservice.NewDescribeAggregationAuthorizationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ConfigService Aggregate Authorization sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ConfigService Aggregate Authorizations (%s): %w", region, err)
		}

		for _, v := range page.AggregationAuthorizations {
			r := resourceAggregateAuthorization()
			d := r.Data(nil)
			d.SetId(aggregateAuthorizationCreateResourceID(aws.ToString(v.AuthorizedAccountId), aws.ToString(v.AuthorizedAwsRegion)))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ConfigService Aggregate Authorizations (%s): %w", region, err)
	}

	return nil
}

func sweepConfigRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ConfigServiceClient(ctx)
	input := &configservice.DescribeConfigRulesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := configservice.NewDescribeConfigRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ConfigService Config Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ConfigService Config Rules (%s): %w", region, err)
		}

		for _, v := range page.ConfigRules {
			name := aws.ToString(v.ConfigRuleName)

			if createdBy := aws.ToString(v.CreatedBy); createdBy != "" {
				log.Printf("[INFO] Skipping ConfigService Config Rule %s: CreatedBy=%s", name, createdBy)
				continue
			}

			r := resourceConfigRule()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ConfigService Config Rules (%s): %w", region, err)
	}

	return nil
}

func sweepConfigurationAggregators(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ConfigServiceClient(ctx)
	input := &configservice.DescribeConfigurationAggregatorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := configservice.NewDescribeConfigurationAggregatorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ConfigService Configuration Aggregator sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ConfigService Configuration Aggregators (%s): %w", region, err)
		}

		for _, v := range page.ConfigurationAggregators {
			r := resourceConfigurationAggregator()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ConfigurationAggregatorName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ConfigService Configuration Aggregators (%s): %w", region, err)
	}

	return nil
}

type configurationRecorderSweeper struct {
	client *conns.AWSClient
	name   string
}

func (s *configurationRecorderSweeper) Delete(ctx context.Context, timeout time.Duration, optFns ...tfresource.OptionsFunc) error {
	r := resourceConfigurationRecorderStatus()
	d := r.Data(nil)
	d.SetId(s.name)

	if err := sdk.NewSweepResource(r, d, s.client).Delete(ctx, timeout, optFns...); err != nil {
		return err
	}

	r = resourceConfigurationRecorder()
	d = r.Data(nil)
	d.SetId(s.name)

	return sdk.NewSweepResource(r, d, s.client).Delete(ctx, timeout, optFns...)
}

func sweepConfigurationRecorder(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ConfigServiceClient(ctx)
	input := &configservice.DescribeConfigurationRecordersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeConfigurationRecorders(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ConfigService Configuration Recorder sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ConfigService Configuration Recorders (%s): %w", region, err)
	}

	for _, v := range output.ConfigurationRecorders {
		sweepResources = append(sweepResources, &configurationRecorderSweeper{client: client, name: aws.ToString(v.Name)})
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ConfigService Configuration Recorders (%s): %w", region, err)
	}

	return nil
}

func sweepConformancePacks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ConfigServiceClient(ctx)
	input := &configservice.DescribeConformancePacksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := configservice.NewDescribeConformancePacksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ConfigService Conformance Pack sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ConfigService Conformance Packs (%s): %w", region, err)
		}

		for _, v := range page.ConformancePackDetails {
			r := resourceConformancePack()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ConformancePackName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ConfigService Conformance Packs (%s): %w", region, err)
	}

	return nil
}

func sweepDeliveryChannels(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ConfigServiceClient(ctx)
	input := &configservice.DescribeDeliveryChannelsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.DescribeDeliveryChannels(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ConfigService Delivery Channel sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing ConfigService Delivery Channels (%s): %w", region, err)
	}

	for _, v := range output.DeliveryChannels {
		r := resourceDeliveryChannel()
		d := r.Data(nil)
		d.SetId(aws.ToString(v.Name))

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ConfigService Delivery Channels (%s): %w", region, err)
	}

	return nil
}

func sweepRemediationConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ConfigServiceClient(ctx)
	input := &configservice.DescribeConfigRulesInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := configservice.NewDescribeConfigRulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ConfigService Remediation Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing ConfigService Config Rules (%s): %w", region, err))
			break
		}

		for _, v := range page.ConfigRules {
			name := aws.ToString(v.ConfigRuleName)

			if createdBy := aws.ToString(v.CreatedBy); createdBy != "" {
				log.Printf("[INFO] Skipping ConfigService Config Rule %s: CreatedBy=%s", name, createdBy)
				continue
			}

			input := &configservice.DescribeRemediationConfigurationsInput{
				ConfigRuleNames: []string{name},
			}
			output, err := conn.DescribeRemediationConfigurations(ctx, input)

			if awsv2.SkipSweepError(err) {
				break
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing ConfigService Remediation Configurations (%s): %w", region, err))
				break
			}

			for _, v := range output.RemediationConfigurations {
				r := resourceRemediationConfiguration()
				d := r.Data(nil)
				d.SetId(aws.ToString(v.ConfigRuleName))

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping ConfigService Remediation Configurations (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
