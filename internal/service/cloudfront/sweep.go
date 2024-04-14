// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cloudfront_cache_policy", &resource.Sweeper{
		Name: "aws_cloudfront_cache_policy",
		F:    sweepCachePolicies,
		Dependencies: []string{
			"aws_cloudfront_distribution",
		},
	})

	// DO NOT add a continuous deployment policy sweeper as these are swept as part of the distribution sweeper
	// resource.AddTestSweepers("aws_cloudfront_continuous_deployment_policy", &resource.Sweeper{
	//	Name: "aws_cloudfront_continuous_deployment_policy",
	//	F:    sweepContinuousDeploymentPolicies,
	//})

	resource.AddTestSweepers("aws_cloudfront_distribution", &resource.Sweeper{
		Name: "aws_cloudfront_distribution",
		F:    sweepDistributions,
	})

	resource.AddTestSweepers("aws_cloudfront_field_level_encryption_config", &resource.Sweeper{
		Name: "aws_cloudfront_field_level_encryption_config",
		F:    sweepFieldLevelEncryptionConfigs,
	})

	resource.AddTestSweepers("aws_cloudfront_field_level_encryption_profile", &resource.Sweeper{
		Name: "aws_cloudfront_field_level_encryption_profile",
		F:    sweepFieldLevelEncryptionProfiles,
		Dependencies: []string{
			"aws_cloudfront_field_level_encryption_config",
		},
	})

	resource.AddTestSweepers("aws_cloudfront_function", &resource.Sweeper{
		Name: "aws_cloudfront_function",
		F:    sweepFunctions,
	})

	resource.AddTestSweepers("aws_cloudfront_key_group", &resource.Sweeper{
		Name: "aws_cloudfront_key_group",
		F:    sweepKeyGroup,
	})

	resource.AddTestSweepers("aws_cloudfront_monitoring_subscription", &resource.Sweeper{
		Name: "aws_cloudfront_monitoring_subscription",
		F:    sweepMonitoringSubscriptions,
		Dependencies: []string{
			"aws_cloudfront_distribution",
		},
	})

	resource.AddTestSweepers("aws_cloudfront_origin_access_control", &resource.Sweeper{
		Name: "aws_cloudfront_origin_access_control",
		F:    sweepOriginAccessControls,
		Dependencies: []string{
			"aws_cloudfront_distribution",
		},
	})

	resource.AddTestSweepers("aws_cloudfront_origin_request_policy", &resource.Sweeper{
		Name: "aws_cloudfront_origin_request_policy",
		F:    sweepOriginRequestPolicies,
		Dependencies: []string{
			"aws_cloudfront_distribution",
		},
	})

	resource.AddTestSweepers("aws_cloudfront_realtime_log_config", &resource.Sweeper{
		Name: "aws_cloudfront_realtime_log_config",
		F:    sweepRealtimeLogsConfig,
	})

	resource.AddTestSweepers("aws_cloudfront_response_headers_policy", &resource.Sweeper{
		Name: "aws_cloudfront_response_headers_policy",
		F:    sweepResponseHeadersPolicies,
		Dependencies: []string{
			"aws_cloudfront_distribution",
		},
	})
}

func sweepCachePolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListCachePoliciesInput{
		Type: awstypes.CachePolicyType(awstypes.ResponseHeadersPolicyTypeCustom),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = ListCachePoliciesPages(ctx, input, func(page *cloudfront.ListCachePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CachePolicyList.Items {
			id := aws.ToString(v.CachePolicy.Id)

			output, err := FindCachePolicyByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}

			r := ResourceCachePolicy()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	}, conn)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Cache Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Cache Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Cache Policies (%s): %w", region, err)
	}

	return nil
}

func sweepDistributions(region string) error {
	var result *multierror.Error

	// 1. Production Distributions
	if err := sweepDistributionsByProductionStaging(region, false); err != nil {
		result = multierror.Append(result, err)
	}

	// 2. Continuous Deployment Policies
	if err := sweepContinuousDeploymentPolicies(region); err != nil {
		result = multierror.Append(result, err)
	}

	// 3. Staging Distributions
	if err := sweepDistributionsByProductionStaging(region, true); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func sweepDistributionsByProductionStaging(region string, staging bool) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListDistributionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	if staging {
		log.Print("[INFO] Sweeping staging distributions")
	} else {
		log.Print("[INFO] Sweeping production distributions")
	}

	pages := cloudfront.NewListDistributionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("error listing distributions (%s): %w", region, err)
		}

		for _, v := range page.DistributionList.Items {
			id := aws.ToString(v.Id)

			output, err := FindDistributionByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}

			r := ResourceDistribution()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Distribution sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Distributions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)
	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Distributions (%s): %w", region, err)
	}

	return nil
}

func sweepContinuousDeploymentPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListContinuousDeploymentPoliciesInput{}

	log.Printf("[INFO] Sweeping continuous deployment policies")
	var result *multierror.Error

	// ListContinuousDeploymentPolicies does not have a paginator
	for {
		output, err := conn.ListContinuousDeploymentPolicies(ctx, input)
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Continuous Deployment Policy sweep for %s: %s", region, err)
			return result.ErrorOrNil()
		}
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("listing CloudFront Continuous Deployment Policies: %w", err))
			break
		}

		if output == nil || output.ContinuousDeploymentPolicyList == nil {
			log.Printf("[WARN] CloudFront Continuous Deployment Policies: empty response")
			break
		}

		for _, cdp := range output.ContinuousDeploymentPolicyList.Items {
			if err := DeleteCDP(ctx, conn, aws.ToString(cdp.ContinuousDeploymentPolicy.Id)); err != nil {
				result = multierror.Append(result, err)
			}
		}

		if output.ContinuousDeploymentPolicyList.NextMarker == nil {
			break
		}

		input.Marker = output.ContinuousDeploymentPolicyList.NextMarker
	}

	return result.ErrorOrNil()
}

func sweepFunctions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &cloudfront.ListFunctionsInput{}
	err = ListFunctionsPages(ctx, input, func(page *cloudfront.ListFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.FunctionList.Items {
			name := aws.ToString(item.Name)

			output, err := findFunctionByTwoPartKey(ctx, conn, name, string(awstypes.FunctionStageDevelopment))

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error reading CloudFront Function (%s): %w", name, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			r := resourceFunction()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	}, conn)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Function sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFront Functions: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping CloudFront Functions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepKeyGroup(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	var sweeperErrs *multierror.Error

	input := &cloudfront.ListKeyGroupsInput{}

	for {
		output, err := conn.ListKeyGroups(ctx, input)
		if err != nil {
			if awsv1.SkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudFront key group sweep for %s: %s", region, err)
				return nil
			}
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudFront key group: %w", err))
			return sweeperErrs.ErrorOrNil()
		}

		if output == nil || output.KeyGroupList == nil || len(output.KeyGroupList.Items) == 0 {
			log.Print("[DEBUG] No CloudFront key group to sweep")
			return nil
		}

		for _, item := range output.KeyGroupList.Items {
			id := item.KeyGroup.Id
			out, err := conn.GetKeyGroup(ctx, &cloudfront.GetKeyGroupInput{
				Id: id,
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error reading CloudFront key group %s: %w", aws.ToString(id), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = conn.DeleteKeyGroup(ctx, &cloudfront.DeleteKeyGroupInput{
				Id:      id,
				IfMatch: out.ETag,
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error sweeping CloudFront key group %s: %w", aws.ToString(id), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if output.KeyGroupList.NextMarker == nil {
			break
		}
		input.Marker = output.KeyGroupList.NextMarker
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepMonitoringSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	var sweeperErrs *multierror.Error

	distributionSummaries := make([]*awstypes.DistributionSummary, 0)

	input := &cloudfront.ListDistributionsInput{}

	pages := cloudfront.NewListDistributionsPaginator(conn, input)
	for pages.HasMorePages() {
		if err != nil {
			return fmt.Errorf("error sweeping CloudFront distributions (%s): %w", region, err)
		}
	}

	if err != nil {
		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Monitoring Subscriptions sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error listing CloudFront Distributions: %s", err)
	}

	if len(distributionSummaries) == 0 {
		log.Print("[DEBUG] No CloudFront Distributions found")
		return nil
	}

	for _, distributionSummary := range distributionSummaries {
		_, err := conn.GetMonitoringSubscription(ctx, &cloudfront.GetMonitoringSubscriptionInput{
			DistributionId: distributionSummary.Id,
		})
		if err != nil {
			return fmt.Errorf("error reading CloudFront Monitoring Subscription %s: %s", aws.ToString(distributionSummary.Id), err)
		}

		_, err = conn.DeleteMonitoringSubscription(ctx, &cloudfront.DeleteMonitoringSubscriptionInput{
			DistributionId: distributionSummary.Id,
		})
		if err != nil {
			return fmt.Errorf("error deleting CloudFront Monitoring Subscription %s: %s", aws.ToString(distributionSummary.Id), err)
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRealtimeLogsConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &cloudfront.ListRealtimeLogConfigsInput{}
	for {
		output, err := conn.ListRealtimeLogConfigs(ctx, input)

		if awsv1.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Real-time Log Configs sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudFront Real-time Log Configs: %w", err))
			return sweeperErrs
		}

		for _, config := range output.RealtimeLogConfigs.Items {
			id := aws.ToString(config.ARN)

			log.Printf("[INFO] Deleting CloudFront Real-time Log Config: %s", id)
			r := ResourceRealtimeLogConfig()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.ToString(output.RealtimeLogConfigs.NextMarker) == "" {
			break
		}
		input.Marker = output.RealtimeLogConfigs.NextMarker
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping CloudFront Real-time Log Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFieldLevelEncryptionConfigs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListFieldLevelEncryptionConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = ListFieldLevelEncryptionConfigsPages(ctx, input, func(page *cloudfront.ListFieldLevelEncryptionConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FieldLevelEncryptionList.Items {
			id := aws.ToString(v.Id)

			output, err := FindFieldLevelEncryptionConfigByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}

			r := ResourceFieldLevelEncryptionConfig()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	}, conn)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Field-level Encryption Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Field-level Encryption Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Field-level Encryption Configs (%s): %w", region, err)
	}

	return nil
}

func sweepFieldLevelEncryptionProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListFieldLevelEncryptionProfilesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = ListFieldLevelEncryptionProfilesPages(ctx, input, func(page *cloudfront.ListFieldLevelEncryptionProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FieldLevelEncryptionProfileList.Items {
			id := aws.ToString(v.Id)

			output, err := FindFieldLevelEncryptionProfileByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}

			r := ResourceFieldLevelEncryptionProfile()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	}, conn)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Field-level Encryption Profile sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Field-level Encryption Profiles (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Field-level Encryption Profiles (%s): %w", region, err)
	}

	return nil
}

func sweepOriginRequestPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListOriginRequestPoliciesInput{
		Type: awstypes.OriginRequestPolicyType(awstypes.ResponseHeadersPolicyTypeCustom),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = ListOriginRequestPoliciesPages(ctx, input, func(page *cloudfront.ListOriginRequestPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OriginRequestPolicyList.Items {
			id := aws.ToString(v.OriginRequestPolicy.Id)

			output, err := FindOriginRequestPolicyByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}

			r := ResourceOriginRequestPolicy()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	}, conn)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Origin Request Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Origin Request Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Origin Request Policies (%s): %w", region, err)
	}

	return nil
}

func sweepResponseHeadersPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListResponseHeadersPoliciesInput{
		Type: awstypes.ResponseHeadersPolicyTypeCustom,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = ListResponseHeadersPoliciesPages(ctx, input, func(page *cloudfront.ListResponseHeadersPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResponseHeadersPolicyList.Items {
			id := aws.ToString(v.ResponseHeadersPolicy.Id)

			output, err := FindResponseHeadersPolicyByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}

			r := ResourceResponseHeadersPolicy()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	}, conn)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Response Headers Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Response Headers Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Response Headers Policies (%s): %w", region, err)
	}

	return nil
}

func sweepOriginAccessControls(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListOriginAccessControlsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = ListOriginAccessControlsPages(ctx, input, func(page *cloudfront.ListOriginAccessControlsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OriginAccessControlList.Items {
			id := aws.ToString(v.Id)

			output, err := findOriginAccessControlByID(ctx, conn, id)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				log.Printf("[WARN] %s", err)
				continue
			}

			r := ResourceOriginAccessControl()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	}, conn)

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Origin Access Control sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Origin Access Controls (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Origin Access Controls (%s): %w", region, err)
	}

	return nil
}
