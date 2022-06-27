//go:build sweep
// +build sweep

package cloudfront

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_cloudfront_cache_policy", &resource.Sweeper{
		Name: "aws_cloudfront_cache_policy",
		F:    sweepCachePolicies,
		Dependencies: []string{
			"aws_cloudfront_distribution",
		},
	})

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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListCachePoliciesInput{
		Type: aws.String(cloudfront.ResponseHeadersPolicyTypeCustom),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = ListCachePoliciesPages(conn, input, func(page *cloudfront.ListCachePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CachePolicyList.Items {
			id := aws.StringValue(v.CachePolicy.Id)

			output, err := FindCachePolicyByID(conn, id)

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
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Cache Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Cache Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Cache Policies (%s): %w", region, err)
	}

	return nil
}

func sweepDistributions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListDistributionsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListDistributionsPages(input, func(page *cloudfront.ListDistributionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DistributionList.Items {
			id := aws.StringValue(v.Id)

			output, err := FindDistributionByID(conn, id)

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

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Distribution sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Distributions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Distributions (%s): %w", region, err)
	}

	return nil
}

func sweepFunctions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListFunctionsInput{}
	var sweeperErrs *multierror.Error

	err = ListFunctionsPages(conn, input, func(page *cloudfront.ListFunctionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.FunctionList.Items {
			name := aws.StringValue(item.Name)

			output, err := FindFunctionByNameAndStage(conn, name, cloudfront.FunctionStageDevelopment)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error reading CloudFront Function (%s): %w", name, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			r := ResourceFunction()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("etag", output.ETag)

			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Function sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFront Functions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepKeyGroup(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	var sweeperErrs *multierror.Error

	input := &cloudfront.ListKeyGroupsInput{}

	for {
		output, err := conn.ListKeyGroups(input)
		if err != nil {
			if sweep.SkipSweepError(err) {
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
			out, err := conn.GetKeyGroup(&cloudfront.GetKeyGroupInput{
				Id: id,
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error reading CloudFront key group %s: %w", aws.StringValue(id), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = conn.DeleteKeyGroup(&cloudfront.DeleteKeyGroupInput{
				Id:      id,
				IfMatch: out.ETag,
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error sweeping CloudFront key group %s: %w", aws.StringValue(id), err)
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	var sweeperErrs *multierror.Error

	distributionSummaries := make([]*cloudfront.DistributionSummary, 0)

	input := &cloudfront.ListDistributionsInput{}
	err = conn.ListDistributionsPages(input, func(page *cloudfront.ListDistributionsOutput, lastPage bool) bool {
		distributionSummaries = append(distributionSummaries, page.DistributionList.Items...)
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
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

		_, err := conn.GetMonitoringSubscription(&cloudfront.GetMonitoringSubscriptionInput{
			DistributionId: distributionSummary.Id,
		})
		if err != nil {
			return fmt.Errorf("error reading CloudFront Monitoring Subscription %s: %s", aws.StringValue(distributionSummary.Id), err)
		}

		_, err = conn.DeleteMonitoringSubscription(&cloudfront.DeleteMonitoringSubscriptionInput{
			DistributionId: distributionSummary.Id,
		})
		if err != nil {
			return fmt.Errorf("error deleting CloudFront Monitoring Subscription %s: %s", aws.StringValue(distributionSummary.Id), err)
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRealtimeLogsConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListRealtimeLogConfigsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListRealtimeLogConfigs(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Real-time Log Configs sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudFront Real-time Log Configs: %w", err))
			return sweeperErrs
		}

		for _, config := range output.RealtimeLogConfigs.Items {
			id := aws.StringValue(config.ARN)

			log.Printf("[INFO] Deleting CloudFront Real-time Log Config: %s", id)
			r := ResourceRealtimeLogConfig()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if aws.StringValue(output.RealtimeLogConfigs.NextMarker) == "" {
			break
		}
		input.Marker = output.RealtimeLogConfigs.NextMarker
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFieldLevelEncryptionConfigs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListFieldLevelEncryptionConfigsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = ListFieldLevelEncryptionConfigsPages(conn, input, func(page *cloudfront.ListFieldLevelEncryptionConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FieldLevelEncryptionList.Items {
			id := aws.StringValue(v.Id)

			output, err := FindFieldLevelEncryptionConfigByID(conn, id)

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
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Field-level Encryption Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Field-level Encryption Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Field-level Encryption Configs (%s): %w", region, err)
	}

	return nil
}

func sweepFieldLevelEncryptionProfiles(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListFieldLevelEncryptionProfilesInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = ListFieldLevelEncryptionProfilesPages(conn, input, func(page *cloudfront.ListFieldLevelEncryptionProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FieldLevelEncryptionProfileList.Items {
			id := aws.StringValue(v.Id)

			output, err := FindFieldLevelEncryptionProfileByID(conn, id)

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
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Field-level Encryption Profile sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Field-level Encryption Profiles (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Field-level Encryption Profiles (%s): %w", region, err)
	}

	return nil
}

func sweepOriginRequestPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListOriginRequestPoliciesInput{
		Type: aws.String(cloudfront.ResponseHeadersPolicyTypeCustom),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = ListOriginRequestPoliciesPages(conn, input, func(page *cloudfront.ListOriginRequestPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OriginRequestPolicyList.Items {
			id := aws.StringValue(v.OriginRequestPolicy.Id)

			output, err := FindOriginRequestPolicyByID(conn, id)

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
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Origin Request Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Origin Request Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Origin Request Policies (%s): %w", region, err)
	}

	return nil
}

func sweepResponseHeadersPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	input := &cloudfront.ListResponseHeadersPoliciesInput{
		Type: aws.String(cloudfront.ResponseHeadersPolicyTypeCustom),
	}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = ListResponseHeadersPoliciesPages(conn, input, func(page *cloudfront.ListResponseHeadersPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResponseHeadersPolicyList.Items {
			id := aws.StringValue(v.ResponseHeadersPolicy.Id)

			output, err := FindResponseHeadersPolicyByID(conn, id)

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
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFront Response Headers Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFront Response Headers Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudFront Response Headers Policies (%s): %w", region, err)
	}

	return nil
}
