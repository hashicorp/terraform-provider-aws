// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_appfabric_app_bundle", sweepAppBundles, "aws_appfabric_app_authorization")
	awsv2.Register("aws_appfabric_app_authorization", sweepAppAuthorizations)
}

func sweepAppBundles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppFabricClient(ctx)
	input := &appfabric.ListAppBundlesInput{}
	var sweepResources []sweep.Sweepable

	pages := appfabric.NewListAppBundlesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AppBundleSummaryList {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAppBundleResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Arn))))
		}
	}

	return sweepResources, nil
}

func sweepAppAuthorizations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppFabricClient(ctx)
	input := &appfabric.ListAppBundlesInput{}
	var sweepResources []sweep.Sweepable

	pages := appfabric.NewListAppBundlesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AppBundleSummaryList {
			appBundleARN := aws.ToString(v.Arn)
			input := &appfabric.ListAppAuthorizationsInput{
				AppBundleIdentifier: aws.String(appBundleARN),
			}

			pages := appfabric.NewListAppAuthorizationsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.AppAuthorizationSummaryList {
					sweepResources = append(sweepResources, framework.NewSweepResource(newAppAuthorizationResource, client,
						framework.NewAttribute("app_bundle_arn", appBundleARN),
						framework.NewAttribute(names.AttrARN, aws.ToString(v.AppAuthorizationArn))))
				}
			}
		}
	}

	return sweepResources, nil
}
