// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package route53

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_route53_health_check", &resource.Sweeper{
		Name: "aws_route53_health_check",
		F:    sweepHealthChecks,
	})

	resource.AddTestSweepers("aws_route53_key_signing_key", &resource.Sweeper{
		Name: "aws_route53_key_signing_key",
		F:    sweepKeySigningKeys,
	})

	resource.AddTestSweepers("aws_route53_query_log", &resource.Sweeper{
		Name: "aws_route53_query_log",
		F:    sweepQueryLogs,
	})

	resource.AddTestSweepers("aws_route53_traffic_policy", &resource.Sweeper{
		Name: "aws_route53_traffic_policy",
		F:    sweepTrafficPolicies,
		Dependencies: []string{
			"aws_route53_traffic_policy_instance",
		},
	})

	resource.AddTestSweepers("aws_route53_traffic_policy_instance", &resource.Sweeper{
		Name: "aws_route53_traffic_policy_instance",
		F:    sweepTrafficPolicyInstances,
	})

	resource.AddTestSweepers("aws_route53_zone", &resource.Sweeper{
		Name: "aws_route53_zone",
		Dependencies: []string{
			"aws_service_discovery_http_namespace",
			"aws_service_discovery_public_dns_namespace",
			"aws_service_discovery_private_dns_namespace",
			"aws_elb",
			"aws_route53_key_signing_key",
			"aws_route53_traffic_policy",
		},
		F: sweepZones,
	})
}

func sweepHealthChecks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.Route53Conn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &route53.ListHealthChecksInput{}

	err = conn.ListHealthChecksPagesWithContext(ctx, input, func(page *route53.ListHealthChecksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, detail := range page.HealthChecks {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			r := ResourceHealthCheck()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Route53 Health Checks for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources, tfresource.WithDelayRand(1*time.Minute), tfresource.WithMinPollInterval(10*time.Second), tfresource.WithPollInterval(18*time.Second)); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Route53 Health Checks for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Health Checks sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepKeySigningKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.Route53Conn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &route53.ListHostedZonesInput{}

	err = conn.ListHostedZonesPagesWithContext(ctx, input, func(page *route53.ListHostedZonesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

	MAIN:
		for _, detail := range page.HostedZones {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			for _, domain := range hostedZonesToPreserve() {
				if strings.Contains(aws.StringValue(detail.Name), domain) {
					log.Printf("[DEBUG] Skipping Route53 Hosted Zone (%s): %s", domain, id)
					continue MAIN
				}
			}

			dnsInput := &route53.GetDNSSECInput{
				HostedZoneId: detail.Id,
			}

			output, err := conn.GetDNSSECWithContext(ctx, dnsInput)

			if tfawserr.ErrMessageContains(err, route53.ErrCodeInvalidArgument, "private hosted zones") {
				continue
			}

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("getting Route53 DNS SEC for %s: %w", region, err))
			}

			for _, dns := range output.KeySigningKeys {
				r := ResourceKeySigningKey()
				d := r.Data(nil)
				d.SetId(id)
				d.Set("hosted_zone_id", id)
				d.Set("name", dns.Name)
				d.Set("status", dns.Status)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}

		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("getting Route53 Key-Signing Keys for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources, tfresource.WithDelayRand(1*time.Minute), tfresource.WithMinPollInterval(30*time.Second), tfresource.WithPollInterval(30*time.Second)); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Route53 Key-Signing Keys for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Key-Signing Keys sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepQueryLogs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.Route53Conn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	err = conn.ListQueryLoggingConfigsPagesWithContext(ctx, &route53.ListQueryLoggingConfigsInput{}, func(page *route53.ListQueryLoggingConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLoggingConfig := range page.QueryLoggingConfigs {
			id := aws.StringValue(queryLoggingConfig.Id)

			r := ResourceQueryLog()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})
	// In unsupported AWS partitions, the API may return an error even the SDK cannot handle.
	// Reference: https://github.com/aws/aws-sdk-go/issues/3313
	if sweep.SkipSweepError(err) || tfawserr.ErrMessageContains(err, "SerializationError", "failed to unmarshal error message") || tfawserr.ErrMessageContains(err, "AccessDeniedException", "Unable to determine service/operation name to be authorized") {
		log.Printf("[WARN] Skipping Route53 query logging configurations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving Route53 query logging configurations: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Route53 query logging configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepTrafficPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.Route53Conn(ctx)
	input := &route53.ListTrafficPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listTrafficPoliciesPages(ctx, conn, input, func(page *route53.ListTrafficPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficPolicySummaries {
			r := ResourceTrafficPolicy()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route 53 Traffic Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Route 53 Traffic Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Route 53 Traffic Policies (%s): %w", region, err)
	}

	return nil
}

func sweepTrafficPolicyInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.Route53Conn(ctx)
	input := &route53.ListTrafficPolicyInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listTrafficPolicyInstancesPages(ctx, conn, input, func(page *route53.ListTrafficPolicyInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficPolicyInstances {
			r := ResourceTrafficPolicyInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route 53 Traffic Policy Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Route 53 Traffic Policy Instances (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Route 53 Traffic Policy Instances (%s): %w", region, err)
	}

	return nil
}

func sweepZones(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	conn := client.Route53Conn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &route53.ListHostedZonesInput{}

	err = conn.ListHostedZonesPagesWithContext(ctx, input, func(page *route53.ListHostedZonesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

	MAIN:
		for _, detail := range page.HostedZones {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			for _, domain := range hostedZonesToPreserve() {
				if strings.Contains(aws.StringValue(detail.Name), domain) {
					log.Printf("[DEBUG] Skipping Route53 Hosted Zone (%s): %s", domain, id)
					continue MAIN
				}
			}

			r := ResourceZone()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("force_destroy", true)
			d.Set("name", detail.Name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing Route53 Hosted Zones for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources, tfresource.WithDelayRand(1*time.Minute), tfresource.WithMinPollInterval(10*time.Second), tfresource.WithPollInterval(18*time.Second)); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Route53 Hosted Zones for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Hosted Zones sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func hostedZonesToPreserve() []string {
	return []string{
		"acmetest.hashicorp.engineering",
		"tfacc.hashicorptest.com",
		"aws.tfacc.hashicorptest.com",
		"hashicorp.com",
		"terraform-provider-aws-acctest-acm.com",
	}
}
