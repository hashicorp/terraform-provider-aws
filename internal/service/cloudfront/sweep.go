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
	// Keep distribution sweeper as an old-style sweeper.
	resource.AddTestSweepers("aws_cloudfront_distribution", &resource.Sweeper{
		Name: "aws_cloudfront_distribution",
		F:    sweepDistributions,
	})

	awsv2.Register("aws_cloudfront_cache_policy", sweepCachePolicies, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_connection_function", sweepConnectionFunctions, "aws_cloudfront_distribution")
	// DO NOT add a continuous deployment policy sweeper as these are swept as part of the distribution sweeper.
	awsv2.Register("aws_cloudfront_field_level_encryption_config", sweepFieldLevelEncryptionConfigs)
	awsv2.Register("aws_cloudfront_field_level_encryption_profile", sweepFieldLevelEncryptionProfiles, "aws_cloudfront_field_level_encryption_config")
	awsv2.Register("aws_cloudfront_function", sweepFunctions)
	awsv2.Register("aws_cloudfront_key_group", sweepKeyGroup)
	awsv2.Register("aws_cloudfront_monitoring_subscription", sweepMonitoringSubscriptions, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_origin_access_control", sweepOriginAccessControls, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_origin_request_policy", sweepOriginRequestPolicies, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_realtime_log_config", sweepRealtimeLogsConfig)
	awsv2.Register("aws_cloudfront_response_headers_policy", sweepResponseHeadersPolicies, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_trust_store", sweepTrustStores, "aws_cloudfront_distribution")
	awsv2.Register("aws_cloudfront_vpc_origin", sweepVPCOrigins)
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
	var input cloudfront.ListDistributionsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	if staging {
		log.Print("[INFO] Sweeping staging CloudFront Distributions")
	} else {
		log.Print("[INFO] Sweeping production CloudFront Distributions")
	}

	pages := cloudfront.NewListDistributionsPaginator(conn, &input)
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
	var input cloudfront.ListContinuousDeploymentPoliciesInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err = listContinuousDeploymentPoliciesPages(ctx, conn, &input, func(page *cloudfront.ListContinuousDeploymentPoliciesOutput, lastPage bool) bool {
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

func sweepCachePolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := cloudfront.ListCachePoliciesInput{
		Type: awstypes.CachePolicyTypeCustom,
	}
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listCachePoliciesPages(ctx, conn, &input, func(page *cloudfront.ListCachePoliciesOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListFunctionsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listFunctionsPages(ctx, conn, &input, func(page *cloudfront.ListFunctionsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepKeyGroup(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListKeyGroupsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listKeyGroupsPages(ctx, conn, &input, func(page *cloudfront.ListKeyGroupsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepMonitoringSubscriptions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListDistributionsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := cloudfront.NewListDistributionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.DistributionList.Items {
			r := resourceMonitoringSubscription()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepRealtimeLogsConfig(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListRealtimeLogConfigsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listRealtimeLogConfigsPages(ctx, conn, &input, func(page *cloudfront.ListRealtimeLogConfigsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}

func sweepFieldLevelEncryptionConfigs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input cloudfront.ListFieldLevelEncryptionConfigsInput
	conn := client.CloudFrontClient(ctx)
	var sweepResources []sweep.Sweepable

	err := listFieldLevelEncryptionConfigsPages(ctx, conn, &input, func(page *cloudfront.ListFieldLevelEncryptionConfigsOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
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
