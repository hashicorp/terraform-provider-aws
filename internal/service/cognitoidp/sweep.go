// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cognito_user_pool_domain", &resource.Sweeper{
		Name: "aws_cognito_user_pool_domain",
		F:    sweepUserPoolDomains,
	})

	resource.AddTestSweepers("aws_cognito_user_pool", &resource.Sweeper{
		Name: "aws_cognito_user_pool",
		F:    sweepUserPools,
		Dependencies: []string{
			"aws_cognito_user_pool_domain",
		},
	})
}

func sweepUserPoolDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(50),
	}
	conn := client.CognitoIDPClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cognitoidentityprovider.NewListUserPoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cognito User Pool Domain sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Cognito User Pools (%s): %w", region, err)
		}

		for _, v := range page.UserPools {
			userPoolID := aws.ToString(v.Id)
			userPool, err := findUserPoolByID(ctx, conn, userPoolID)

			if err != nil {
				log.Printf("[ERROR] Reading Cognito User Pool (%s): %s", userPoolID, err)
				continue
			}

			if domain := aws.ToString(userPool.Domain); domain != "" {
				r := resourceUserPoolDomain()
				d := r.Data(nil)
				d.SetId(domain)
				d.Set(names.AttrUserPoolID, userPoolID)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cognito User Pool Domains (%s): %w", region, err)
	}

	return nil
}

func sweepUserPools(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(50),
	}
	conn := client.CognitoIDPClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cognitoidentityprovider.NewListUserPoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cognito User Pool sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Cognito User Pools (%s): %w", region, err)
		}

		for _, v := range page.UserPools {
			r := resourceUserPool()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cognito User Pools (%s): %w", region, err)
	}

	return nil
}
