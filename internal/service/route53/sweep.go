// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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
	conn := client.Route53Client(ctx)
	input := &route53.ListHealthChecksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53.NewListHealthChecksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route 53 Health Check sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Route 53 Health Checks (%s): %w", region, err)
		}

		for _, v := range page.HealthChecks {
			id := aws.ToString(v.Id)

			if v.LinkedService != nil {
				log.Printf("[INFO] Skipping Route 53 Health Check %s: %s", id, aws.ToString(v.LinkedService.Description))
				continue
			}

			r := resourceHealthCheck()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Route 53 Health Checks (%s): %w", region, err)
	}

	return nil
}

func sweepKeySigningKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.Route53Client(ctx)
	input := &route53.ListHostedZonesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53.NewListHostedZonesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Key-Signing Key sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Route 53 Hosted Zones (%s): %w", region, err)
		}

	MAIN:
		for _, v := range page.HostedZones {
			zoneID := aws.ToString(v.Id)

			for _, domain := range hostedZonesToPreserve() {
				if strings.Contains(aws.ToString(v.Name), domain) {
					log.Printf("[DEBUG] Skipping Route53 Hosted Zone (%s): %s", domain, zoneID)
					continue MAIN
				}
			}

			output, err := findHostedZoneDNSSECByZoneID(ctx, conn, zoneID)

			if err != nil {
				continue
			}

			for _, v := range output.KeySigningKeys {
				r := resourceKeySigningKey()
				d := r.Data(nil)
				d.SetId(zoneID)
				d.Set(names.AttrHostedZoneID, zoneID)
				d.Set(names.AttrName, v.Name)
				d.Set(names.AttrStatus, v.Status)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Route 53 Key-Signing Keys (%s): %w", region, err)
	}

	return nil
}

func sweepQueryLogs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.Route53Client(ctx)
	input := &route53.ListQueryLoggingConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53.NewListQueryLoggingConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		// In unsupported AWS partitions, the API may return an error even the SDK cannot handle.
		// Reference: https://github.com/aws/aws-sdk-go/issues/3313
		if awsv2.SkipSweepError(err) || tfawserr.ErrMessageContains(err, errCodeSerializationError, "failed to unmarshal error message") || tfawserr.ErrMessageContains(err, errCodeAccessDenied, "Unable to determine service/operation name to be authorized") {
			log.Printf("[WARN] Skipping Route53 Query Logging Config sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Route 53 Query Logging Configs (%s): %w", region, err)
		}

		for _, v := range page.QueryLoggingConfigs {
			id := aws.ToString(v.Id)

			r := resourceQueryLog()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Route 53  Query Logging Configs (%s): %w", region, err)
	}

	return nil
}

func sweepTrafficPolicies(region string) error {
	ctx := sweep.Context(region)
	if region == names.USGovEast1RegionID || region == names.USGovWest1RegionID {
		log.Printf("[WARN] Skipping Route 53 Traffic Policy sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.Route53Client(ctx)
	input := &route53.ListTrafficPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listTrafficPoliciesPages(ctx, conn, input, func(page *route53.ListTrafficPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficPolicySummaries {
			r := resourceTrafficPolicy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
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
	if region == names.USGovEast1RegionID || region == names.USGovWest1RegionID {
		log.Printf("[WARN] Skipping Route 53 Traffic Policy Instance sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.Route53Client(ctx)
	input := &route53.ListTrafficPolicyInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listTrafficPolicyInstancesPages(ctx, conn, input, func(page *route53.ListTrafficPolicyInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficPolicyInstances {
			r := resourceTrafficPolicyInstance()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
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
	conn := client.Route53Client(ctx)
	input := &route53.ListHostedZonesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := route53.NewListHostedZonesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route 53 Hosted Zone sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Route 53 Hosted Zones (%s): %w", region, err)
		}

	MAIN:
		for _, v := range page.HostedZones {
			zoneID := aws.ToString(v.Id)

			for _, domain := range hostedZonesToPreserve() {
				if strings.Contains(aws.ToString(v.Name), domain) {
					log.Printf("[DEBUG] Skipping Route53 Hosted Zone %s: %s", zoneID, domain)
					continue MAIN
				}
			}

			if v.LinkedService != nil {
				log.Printf("[INFO] Skipping Route 53 Hosted Zone %s: %s", zoneID, aws.ToString(v.LinkedService.Description))
				continue MAIN
			}

			r := resourceZone()
			d := r.Data(nil)
			d.SetId(zoneID)
			d.Set(names.AttrForceDestroy, true)
			d.Set(names.AttrName, v.Name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Route 53 Hosted Zones (%s): %w", region, err)
	}

	return nil
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
