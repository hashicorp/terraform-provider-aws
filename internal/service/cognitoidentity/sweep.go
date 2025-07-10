// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cognito_identity_pool", &resource.Sweeper{
		Name: "aws_cognito_identity_pool",
		F:    sweepIdentityPools,
	})
}

func sweepIdentityPools(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	input := &cognitoidentity.ListIdentityPoolsInput{
		MaxResults: aws.Int32(50),
	}
	conn := client.CognitoIdentityClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cognitoidentity.NewListIdentityPoolsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cognito Identity Pool sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Cognito Identity Pools (%s): %w", region, err)
		}

		for _, v := range page.IdentityPools {
			r := resourcePool()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.IdentityPoolId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cognito Identity Pools (%s): %w", region, err)
	}

	return nil
}
