// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package opensearchserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.OpenSearchServerlessClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &opensearchserverless.ListAccessPoliciesInput{
		Type: types.AccessPolicyTypeData,
	}

	pages := opensearchserverless.NewListAccessPoliciesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) {
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
				framework.NewAttribute("id", name),
				framework.NewAttribute("name", name),
				framework.NewAttribute("type", ap.Type),
			))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping OpenSearch Serverless Access Policies for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Serverless Access Policies sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepCollections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.OpenSearchServerlessClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &opensearchserverless.ListCollectionsInput{}

	pages := opensearchserverless.NewListCollectionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) {
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
				framework.NewAttribute("id", id),
			))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping OpenSearch Serverless Collections for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Serverless Collections sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSecurityConfigs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.OpenSearchServerlessClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &opensearchserverless.ListSecurityConfigsInput{
		Type: types.SecurityConfigTypeSaml,
	}
	pages := opensearchserverless.NewListSecurityConfigsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) {
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
				framework.NewAttribute("id", id),
			))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping OpenSearch Serverless Security Configs for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Serverless Security Configs sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepSecurityPolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.OpenSearchServerlessClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	inputEncryption := &opensearchserverless.ListSecurityPoliciesInput{
		Type: types.SecurityPolicyTypeEncryption,
	}
	pagesEncryption := opensearchserverless.NewListSecurityPoliciesPaginator(conn, inputEncryption)

	for pagesEncryption.HasMorePages() {
		page, err := pagesEncryption.NextPage(ctx)
		if sweep.SkipSweepError(err) {
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
				framework.NewAttribute("id", name),
				framework.NewAttribute("name", name),
				framework.NewAttribute("type", sp.Type),
			))
		}
	}

	inputNetwork := &opensearchserverless.ListSecurityPoliciesInput{
		Type: types.SecurityPolicyTypeNetwork,
	}
	pagesNetwork := opensearchserverless.NewListSecurityPoliciesPaginator(conn, inputNetwork)

	for pagesNetwork.HasMorePages() {
		page, err := pagesNetwork.NextPage(ctx)
		if sweep.SkipSweepError(err) {
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
				framework.NewAttribute("id", name),
				framework.NewAttribute("name", name),
				framework.NewAttribute("type", sp.Type),
			))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping OpenSearch Serverless Security Policies for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Serverless Security Policies sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepVPCEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.OpenSearchServerlessClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &opensearchserverless.ListVpcEndpointsInput{}

	pages := opensearchserverless.NewListVpcEndpointsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if sweep.SkipSweepError(err) {
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
				framework.NewAttribute("id", id),
			))
		}
	}

	if err := sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping OpenSearch Serverless VPC Endpoints for %s: %w", region, err))
	}
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping OpenSearch Serverless VPC Endpoint sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
