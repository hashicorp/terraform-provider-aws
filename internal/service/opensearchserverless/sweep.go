// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_opensearchserverless_access_policy", &resource.Sweeper{
		Name: "aws_opensearchserverless_access_policy",
		F:    sweepAccessPolicies,
	})
	resource.AddTestSweepers("aws_opensearchserverless_collection", &resource.Sweeper{
		Name: "aws_opensearchserverless_collection",
		F:    sweepCollections,
	})
	resource.AddTestSweepers("aws_opensearchserverless_security_config", &resource.Sweeper{
		Name: "aws_opensearchserverless_security_config",
		F:    sweepSecurityConfigs,
	})
	resource.AddTestSweepers("aws_opensearchserverless_security_policy", &resource.Sweeper{
		Name: "aws_opensearchserverless_security_policy",
		F:    sweepSecurityPolicies,
	})
	resource.AddTestSweepers("aws_opensearchserverless_vpc_endpoint", &resource.Sweeper{
		Name: "aws_opensearchserverless_vpc_endpoint",
		F:    sweepVPCEndpoints,
	})
}

func sweepAccessPolicies(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping OpenSearch Serverless Access Policy sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchServerlessClient(ctx)
	input := &opensearchserverless.ListAccessPoliciesInput{
		Type: types.AccessPolicyTypeData,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := opensearchserverless.NewListAccessPoliciesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping OpenSearch Serverless Access Policies sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving OpenSearch Serverless Access Policies: %w", err)
		}

		for _, ap := range page.AccessPolicySummaries {
			name := aws.ToString(ap.Name)

			log.Printf("[INFO] Deleting OpenSearch Serverless Access Policy: %s", name)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceAccessPolicy, client,
				framework.NewAttribute(names.AttrID, name),
				framework.NewAttribute(names.AttrName, name),
				framework.NewAttribute(names.AttrType, ap.Type),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping OpenSearch Serverless Access Policies for %s: %w", region, err)
	}

	return nil
}

func sweepCollections(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping OpenSearch Serverless Collection sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchServerlessClient(ctx)
	input := &opensearchserverless.ListCollectionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := opensearchserverless.NewListCollectionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping OpenSearch Serverless Collections sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving OpenSearch Serverless Collections: %w", err)
		}

		for _, collection := range page.CollectionSummaries {
			id := aws.ToString(collection.Id)

			log.Printf("[INFO] Deleting OpenSearch Serverless Collection: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceCollection, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping OpenSearch Serverless Collections for %s: %w", region, err)
	}

	return nil
}

func sweepSecurityConfigs(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping OpenSearch Serverless Security Config sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchServerlessClient(ctx)
	input := &opensearchserverless.ListSecurityConfigsInput{
		Type: types.SecurityConfigTypeSaml,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := opensearchserverless.NewListSecurityConfigsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping OpenSearch Serverless Security Configs sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving OpenSearch Serverless Security Configs: %w", err)
		}

		for _, sc := range page.SecurityConfigSummaries {
			id := aws.ToString(sc.Id)

			log.Printf("[INFO] Deleting OpenSearch Serverless Security Config: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceCollection, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping OpenSearch Serverless Security Configs for %s: %w", region, err)
	}

	return nil
}

func sweepSecurityPolicies(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping OpenSearch Serverless Security Policy sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchServerlessClient(ctx)
	inputEncryption := &opensearchserverless.ListSecurityPoliciesInput{
		Type: types.SecurityPolicyTypeEncryption,
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pagesEncryption := opensearchserverless.NewListSecurityPoliciesPaginator(conn, inputEncryption)

	for pagesEncryption.HasMorePages() {
		page, err := pagesEncryption.NextPage(ctx)
		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping OpenSearch Serverless Security Policies sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving OpenSearch Serverless Security Policies: %w", err)
		}

		for _, sp := range page.SecurityPolicySummaries {
			name := aws.ToString(sp.Name)

			log.Printf("[INFO] Deleting OpenSearch Serverless Security Policy: %s", name)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceCollection, client,
				framework.NewAttribute(names.AttrID, name),
				framework.NewAttribute(names.AttrName, name),
				framework.NewAttribute(names.AttrType, sp.Type),
			))
		}
	}

	inputNetwork := &opensearchserverless.ListSecurityPoliciesInput{
		Type: types.SecurityPolicyTypeNetwork,
	}
	pagesNetwork := opensearchserverless.NewListSecurityPoliciesPaginator(conn, inputNetwork)

	for pagesNetwork.HasMorePages() {
		page, err := pagesNetwork.NextPage(ctx)
		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping OpenSearch Serverless Security Policies sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving OpenSearch Serverless Security Policies: %w", err)
		}

		for _, sp := range page.SecurityPolicySummaries {
			name := aws.ToString(sp.Name)

			log.Printf("[INFO] Deleting OpenSearch Serverless Security Policy: %s", name)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceCollection, client,
				framework.NewAttribute(names.AttrID, name),
				framework.NewAttribute(names.AttrName, name),
				framework.NewAttribute(names.AttrType, sp.Type),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping OpenSearch Serverless Security Policies for %s: %w", region, err)
	}

	return nil
}

func sweepVPCEndpoints(region string) error {
	ctx := sweep.Context(region)
	if region == names.USWest1RegionID {
		log.Printf("[WARN] Skipping OpenSearch Serverless Security Policy sweep for region: %s", region)
		return nil
	}
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.OpenSearchServerlessClient(ctx)
	input := &opensearchserverless.ListVpcEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := opensearchserverless.NewListVpcEndpointsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) || skipSweepErr(err) {
			log.Printf("[WARN] Skipping OpenSearch Serverless VPC Endpoints sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving OpenSearch Serverless VPC Endpoints: %w", err)
		}

		for _, endpoint := range page.VpcEndpointSummaries {
			id := aws.ToString(endpoint.Id)

			log.Printf("[INFO] Deleting OpenSearch Serverless VPC Endpoint: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceCollection, client,
				framework.NewAttribute(names.AttrID, id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping OpenSearch Serverless VPC Endpoints for %s: %w", region, err)
	}

	return nil
}

func skipSweepErr(err error) bool {
	// OpenSearch Serverless returns this error when the service is not supported in the region
	return tfawserr.ErrMessageContains(err, "AccessDeniedException", "UnknownError")
}
