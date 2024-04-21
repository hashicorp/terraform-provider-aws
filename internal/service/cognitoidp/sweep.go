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
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	conn := client.CognitoIDPClient(ctx)

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(50),
	}

	pages := cognitoidentityprovider.NewListUserPoolsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv2.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Cognito User Pool Domain sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Cognito User Pools: %s", err)
		}

		for _, u := range page.UserPools {
			output, err := conn.DescribeUserPool(ctx, &cognitoidentityprovider.DescribeUserPoolInput{
				UserPoolId: u.Id,
			})
			if err != nil {
				log.Printf("[ERROR] Failed describing Cognito user pool (%s): %s", aws.ToString(u.Name), err)
				continue
			}
			if output.UserPool != nil && output.UserPool.Domain != nil {
				domain := aws.ToString(output.UserPool.Domain)

				log.Printf("[INFO] Deleting Cognito user pool domain: %s", domain)
				_, err := conn.DeleteUserPoolDomain(ctx, &cognitoidentityprovider.DeleteUserPoolDomainInput{
					Domain:     output.UserPool.Domain,
					UserPoolId: u.Id,
				})
				if err != nil {
					log.Printf("[ERROR] Failed deleting Cognito user pool domain (%s): %s", domain, err)
				}
			}
		}
	}

	return nil
}

func sweepUserPools(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CognitoIDPClient(ctx)

	input := &cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(50),
	}

	pages := cognitoidentityprovider.NewListUserPoolsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			if awsv1.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Cognito User Pool sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Cognito User Pools: %w", err)
		}

		for _, userPool := range page.UserPools {
			name := aws.ToString(userPool.Name)

			log.Printf("[INFO] Deleting Cognito User Pool: %s", name)
			_, err := conn.DeleteUserPool(ctx, &cognitoidentityprovider.DeleteUserPoolInput{
				UserPoolId: userPool.Id,
			})
			if err != nil {
				log.Printf("[ERROR] Failed deleting Cognito User Pool (%s): %s", name, err)
			}
		}
	}

	return nil
}
