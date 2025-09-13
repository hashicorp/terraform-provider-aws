// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_cognito_user_pool_domain", sweepUserPoolDomains)
	awsv2.Register("aws_cognito_user_pool", sweepUserPools, "aws_cognito_user_pool_domain")
}

func sweepUserPoolDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CognitoIDPClient(ctx)
	input := cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(50),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cognitoidentityprovider.NewListUserPoolsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.UserPools {
			userPoolID := aws.ToString(v.Id)
			userPool, err := findUserPoolByID(ctx, conn, userPoolID)

			if err != nil {
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

	return sweepResources, nil
}

func sweepUserPools(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CognitoIDPClient(ctx)
	input := cognitoidentityprovider.ListUserPoolsInput{
		MaxResults: aws.Int32(50),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cognitoidentityprovider.NewListUserPoolsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.UserPools {
			userPoolID := aws.ToString(v.Id)
			userPool, err := findUserPoolByID(ctx, conn, userPoolID)

			if err != nil {
				continue
			}

			if deletionProtection := userPool.DeletionProtection; deletionProtection == awstypes.DeletionProtectionTypeActive {
				log.Printf("[INFO] Skipping Cognito User Pool %s: DeletionProtection=%s", userPoolID, deletionProtection)
				continue
			}

			r := resourceUserPool()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
