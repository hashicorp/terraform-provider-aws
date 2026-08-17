// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_resiliencehubv2_assertion", sweepAssertions)
	awsv2.Register("aws_resiliencehubv2_input_source", sweepInputSources)
	awsv2.Register("aws_resiliencehubv2_policy", sweepPolicies,
		"aws_resiliencehubv2_service",
	)
	awsv2.Register("aws_resiliencehubv2_service", sweepServices,
		"aws_resiliencehubv2_assertion",
		"aws_resiliencehubv2_input_source",
		"aws_resiliencehubv2_service_function",
	)
	awsv2.Register("aws_resiliencehubv2_service_function", sweepServiceFunctions)
	awsv2.Register("aws_resiliencehubv2_system", sweepSystems,
		"aws_resiliencehubv2_user_journey",
	)
	awsv2.Register("aws_resiliencehubv2_user_journey", sweepUserJourneys)
}

func sweepPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)
	var input resiliencehubv2.ListPoliciesInput
	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListPoliciesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.PolicySummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newPolicyResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.PolicyArn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepSystems(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)
	var input resiliencehubv2.ListSystemsInput
	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListSystemsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.SystemSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newSystemResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.SystemArn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepServices(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)
	var input resiliencehubv2.ListServicesInput
	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListServicesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ServiceSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newServiceResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.ServiceArn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepInputSources(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)
	var input resiliencehubv2.ListServicesInput
	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListServicesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ServiceSummaries {
			serviceARN := aws.ToString(v.ServiceArn)
			input := resiliencehubv2.ListInputSourcesInput{
				ServiceArn: aws.String(serviceARN),
			}

			pages := resiliencehubv2.NewListInputSourcesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.InputSourceSummaries {
					sweepResources = append(sweepResources, framework.NewSweepResource(newInputSourceResource, client,
						framework.NewAttribute("service_arn", serviceARN),
						framework.NewAttribute("input_source_id", aws.ToString(v.InputSourceId)),
					))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepServiceFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)
	var input resiliencehubv2.ListServicesInput
	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListServicesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ServiceSummaries {
			serviceARN := aws.ToString(v.ServiceArn)
			input := resiliencehubv2.ListServiceFunctionsInput{
				ServiceArn: aws.String(serviceARN),
			}

			pages := resiliencehubv2.NewListServiceFunctionsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.ServiceFunctions {
					sweepResources = append(sweepResources, framework.NewSweepResource(newServiceFunctionResource, client,
						framework.NewAttribute("service_arn", serviceARN),
						framework.NewAttribute("service_function_id", aws.ToString(v.ServiceFunctionId)),
					))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepAssertions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)
	var input resiliencehubv2.ListServicesInput
	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListServicesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.ServiceSummaries {
			serviceARN := aws.ToString(v.ServiceArn)
			input := resiliencehubv2.ListAssertionsInput{
				ServiceArn: aws.String(serviceARN),
			}

			pages := resiliencehubv2.NewListAssertionsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.Assertions {
					sweepResources = append(sweepResources, framework.NewSweepResource(newAssertionResource, client,
						framework.NewAttribute("service_arn", serviceARN),
						framework.NewAttribute("assertion_id", aws.ToString(v.AssertionId)),
					))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepUserJourneys(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)
	var input resiliencehubv2.ListSystemsInput
	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListSystemsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.SystemSummaries {
			systemARN := aws.ToString(v.SystemArn)
			input := resiliencehubv2.ListUserJourneysInput{
				SystemArn: aws.String(systemARN),
			}

			pages := resiliencehubv2.NewListUserJourneysPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.UserJourneySummaries {
					sweepResources = append(sweepResources, framework.NewSweepResource(newResourceUserJourney, client,
						framework.NewAttribute("system_arn", systemARN),
						framework.NewAttribute("user_journey_id", aws.ToString(v.UserJourneyId)),
					))
				}
			}
		}
	}

	return sweepResources, nil
}
