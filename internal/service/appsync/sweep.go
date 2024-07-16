// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_appsync_graphql_api", &resource.Sweeper{
		Name: "aws_appsync_graphql_api",
		F:    sweepGraphQLAPIs,
	})

	resource.AddTestSweepers("aws_appsync_domain_name", &resource.Sweeper{
		Name: "aws_appsync_domain_name",
		F:    sweepDomainNames,
		Dependencies: []string{
			"aws_appsync_domain_name_api_association",
		},
	})

	resource.AddTestSweepers("aws_appsync_domain_name_api_association", &resource.Sweeper{
		Name: "aws_appsync_domain_name_api_association",
		F:    sweepDomainNameAssociations,
	})
}

func sweepGraphQLAPIs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	input := &appsync.ListGraphqlApisInput{}
	conn := client.AppSyncClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = listGraphQLAPIsPages(ctx, conn, input, func(page *appsync.ListGraphqlApisOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GraphqlApis {
			r := resourceGraphQLAPI()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ApiId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppSync GraphQL API sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppSync GraphQL APIs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppSync GraphQL APIs (%s): %w", region, err)
	}

	return nil
}

func sweepDomainNames(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	input := &appsync.ListDomainNamesInput{}
	conn := client.AppSyncClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = listDomainNamesPages(ctx, conn, input, func(page *appsync.ListDomainNamesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DomainNameConfigs {
			r := resourceDomainName()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DomainName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppSync Domain Name sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppSync Domain Names (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppSync Domain Names (%s): %w", region, err)
	}

	return nil
}

func sweepDomainNameAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	input := &appsync.ListDomainNamesInput{}
	conn := client.AppSyncClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = listDomainNamesPages(ctx, conn, input, func(page *appsync.ListDomainNamesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DomainNameConfigs {
			r := resourceDomainNameAPIAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DomainName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppSync Domain Name API Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppSync Domain Names (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppSync Domain Name API Associations (%s): %w", region, err)
	}

	return nil
}
