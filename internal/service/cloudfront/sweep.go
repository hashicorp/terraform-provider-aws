// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
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

	resource.AddTestSweepers("aws_cloudfront_realtime_log_config", &resource.Sweeper{
		Name: "aws_cloudfront_realtime_log_config",
		F:    sweepRealtimeLogsConfig,
	})

	awsv2.Register("aws_cloudfront_connection_function", sweepConnectionFunctions, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_field_level_encryption_profile", sweepFieldLevelEncryptionProfiles, "aws_cloudfront_field_level_encryption_config")
	awsv2.Register("aws_cloudfront_origin_access_control", sweepOriginAccessControls, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_origin_request_policy", sweepOriginRequestPolicies, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_response_headers_policy", sweepResponseHeadersPolicies, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_trust_store", sweepTrustStores, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_vpc_origin", sweepVPCOrigins)
}

func sweepCachePolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListCachePoliciesInput{
		Type: awstypes.CachePolicyTypeCustom,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listCachePoliciesPages(ctx, conn, input, func(page *cloudfront.ListCachePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CachePolicyList.Items {
			id := aws.ToString(v.CachePolicy.Id)
			output, err := findCachePolicyByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceCachePolicy()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
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
	var errs []error

	// 1. Production Distributions
	if err := sweepDistributionsByProductionOrStaging(region, false); err != nil {
		errs = append(errs, err)
	}

	// 2. Continuous Deployment Policies
	if err := sweepContinuousDeploymentPolicies(region); err != nil {
		errs = append(errs, err)
	}

	// 3. Staging Distributions
	if err := sweepDistributionsByProductionOrStaging(region, true); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func sweepDistributionsByProductionOrStaging(region string, staging bool) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListDistributionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	if staging {
		log.Print("[INFO] Sweeping staging CloudFront Distributions")
	} else {
		log.Print("[INFO] Sweeping production CloudFront Distributions")
	}

	pages := cloudfront.NewListDistributionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Distribution sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudFront Distributions (%s): %w", region, err)
		}

		for _, v := range page.DistributionList.Items {
			id := aws.ToString(v.Id)
			output, err := findDistributionByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			if staging != aws.ToBool(output.Distribution.DistributionConfig.Staging) {
				continue
			}

			r := resourceDistribution()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListContinuousDeploymentPoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listContinuousDeploymentPoliciesPages(ctx, conn, input, func(page *cloudfront.ListContinuousDeploymentPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ContinuousDeploymentPolicyList.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newContinuousDeploymentPolicyResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.ContinuousDeploymentPolicy.Id)),
			))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Continuous Deployment Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Continuous Deployment Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Continuous Deployment Policies (%s): %w", region, err)
	}

	return nil
}

func sweepFunctions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListFunctionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listFunctionsPages(ctx, conn, input, func(page *cloudfront.ListFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FunctionList.Items {
			name := aws.ToString(v.Name)
			output, err := findFunctionByTwoPartKey(ctx, conn, name, awstypes.FunctionStageDevelopment)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceFunction()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Function sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Functions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Functions (%s): %w", region, err)
	}

	return nil
}

func sweepKeyGroup(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListKeyGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listKeyGroupsPages(ctx, conn, input, func(page *cloudfront.ListKeyGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.KeyGroupList.Items {
			id := aws.ToString(v.KeyGroup.Id)
			output, err := findKeyGroupByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceKeyGroup()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Key Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Key Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Key Groups (%s): %w", region, err)
	}

	return nil
}

func sweepMonitoringSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListDistributionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudfront.NewListDistributionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Monitoring Subscription sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudFront Distributions (%s): %w", region, err)
		}

		for _, v := range page.DistributionList.Items {
			r := resourceMonitoringSubscription()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Monitoring Subscriptions (%s): %w", region, err)
	}

	return nil
}

func sweepRealtimeLogsConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListRealtimeLogConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRealtimeLogConfigsPages(ctx, conn, input, func(page *cloudfront.ListRealtimeLogConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RealtimeLogConfigs.Items {
			r := resourceRealtimeLogConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ARN))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Real-time Log Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Real-time Log Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Real-time Log Configs (%s): %w", region, err)
	}

	return nil
}

func sweepFieldLevelEncryptionConfigs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.CloudFrontClient(ctx)
	input := &cloudfront.ListFieldLevelEncryptionConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listFieldLevelEncryptionConfigsPages(ctx, conn, input, func(page *cloudfront.ListFieldLevelEncryptionConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FieldLevelEncryptionList.Items {
			id := aws.ToString(v.Id)
			output, err := findFieldLevelEncryptionConfigByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceFieldLevelEncryptionConfig()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
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

func sweepFieldLevelEncryptionProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListFieldLevelEncryptionProfilesInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listFieldLevelEncryptionProfilesPages(ctx, conn, &input, func(page *cloudfront.ListFieldLevelEncryptionProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FieldLevelEncryptionProfileList.Items {
			id := aws.ToString(v.Id)
			output, err := findFieldLevelEncryptionProfileByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceFieldLevelEncryptionProfile()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepOriginRequestPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := cloudfront.ListOriginRequestPoliciesInput{
		Type: awstypes.OriginRequestPolicyTypeCustom,
	}
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listOriginRequestPoliciesPages(ctx, conn, &input, func(page *cloudfront.ListOriginRequestPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OriginRequestPolicyList.Items {
			id := aws.ToString(v.OriginRequestPolicy.Id)
			output, err := findOriginRequestPolicyByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceOriginRequestPolicy()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepResponseHeadersPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := cloudfront.ListResponseHeadersPoliciesInput{
		Type: awstypes.ResponseHeadersPolicyTypeCustom,
	}
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listResponseHeadersPoliciesPages(ctx, conn, &input, func(page *cloudfront.ListResponseHeadersPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResponseHeadersPolicyList.Items {
			id := aws.ToString(v.ResponseHeadersPolicy.Id)
			output, err := findResponseHeadersPolicyByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceResponseHeadersPolicy()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepOriginAccessControls(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListOriginAccessControlsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listOriginAccessControlsPages(ctx, conn, &input, func(page *cloudfront.ListOriginAccessControlsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OriginAccessControlList.Items {
			id := aws.ToString(v.Id)
			output, err := findOriginAccessControlByID(ctx, conn, id)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				continue
			}

			r := resourceOriginAccessControl()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("etag", output.ETag)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepConnectionFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListConnectionFunctionsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudfront.NewListConnectionFunctionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ConnectionFunctions {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceConnectionFunction, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}

func sweepTrustStores(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListTrustStoresInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudfront.NewListTrustStoresPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.TrustStoreList {
			sweepResources = append(sweepResources, framework.NewSweepResource(newTrustStoreResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id)),
			))
		}
	}

	return sweepResources, nil
}

func sweepVPCOrigins(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListVpcOriginsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listVPCOriginsPages(ctx, conn, &input, func(page *cloudfront.ListVpcOriginsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcOriginList.Items {
			sweepResources = append(sweepResources, framework.NewSweepResource(newVPCOriginResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id))))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
