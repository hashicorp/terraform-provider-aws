// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_appsync_api", sweepAPIs)
	awsv2.Register("aws_appsync_graphql_api", sweepGraphQLAPIs, "aws_appsync_domain_name_api_association")
	awsv2.Register("aws_appsync_domain_name", sweepDomainNames, "aws_appsync_domain_name_api_association")
	awsv2.Register("aws_appsync_domain_name_api_association", sweepDomainNameAssociations)
}

func sweepAPIs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppSyncClient(ctx)
	var input appsync.ListApisInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := listAPIsPages(ctx, conn, &input, func(page *appsync.ListApisOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Apis {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAPIResource, client,
				framework.NewAttribute("api_id", aws.ToString(v.ApiId))))
		}

		return !lastPage
	})

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return sweepResources, nil
}

func sweepGraphQLAPIs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppSyncClient(ctx)
	var input appsync.ListGraphqlApisInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := listGraphQLAPIsPages(ctx, conn, &input, func(page *appsync.ListGraphqlApisOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return sweepResources, nil
}

func sweepDomainNames(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppSyncClient(ctx)
	var input appsync.ListDomainNamesInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := listDomainNamesPages(ctx, conn, &input, func(page *appsync.ListDomainNamesOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return sweepResources, nil
}

func sweepDomainNameAssociations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppSyncClient(ctx)
	var input appsync.ListDomainNamesInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := listDomainNamesPages(ctx, conn, &input, func(page *appsync.ListDomainNamesOutput, lastPage bool) bool {
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

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	return sweepResources, nil
}
