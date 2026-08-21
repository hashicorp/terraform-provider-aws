// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_accountaccess_application", sweepApplications)
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AccountAccessClient(ctx)
	var sweepResources []sweep.Sweepable
	var nextToken *string

	for {
		output, err := conn.ListApplications(ctx, &accountaccess.ListApplicationsInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, app := range output.Applications {
			arn := aws.ToString(app.ApplicationArn)
			if arn == "" {
				continue
			}
			sweepResources = append(sweepResources, framework.NewSweepResource(
				newApplicationResource, client,
				framework.NewAttribute(names.AttrARN, arn),
				framework.NewAttribute(names.AttrID, arn),
			))
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return sweepResources, nil
}
