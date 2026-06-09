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
	awsv2.Register("aws_resiliencehubv2_input_source", sweepInputSources)
	awsv2.Register("aws_resiliencehubv2_service_function", sweepServiceFunctions)
	awsv2.Register("aws_resiliencehubv2_assertion", sweepAssertions)
	awsv2.Register("aws_resiliencehubv2_user_journey", sweepUserJourneys)
	awsv2.Register("aws_resiliencehubv2_policy", sweepPolicies,
		"aws_resiliencehubv2_service",
	)
	awsv2.Register("aws_resiliencehubv2_system", sweepSystems,
		"aws_resiliencehubv2_user_journey",
	)
	awsv2.Register("aws_resiliencehubv2_service", sweepServices,
		"aws_resiliencehubv2_input_source",
		"aws_resiliencehubv2_service_function",
		"aws_resiliencehubv2_assertion",
	)
}

func sweepPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListPoliciesPaginator(conn, &resiliencehubv2.ListPoliciesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, policy := range page.PolicySummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourcePolicy, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(policy.PolicyArn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepSystems(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListSystemsPaginator(conn, &resiliencehubv2.ListSystemsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, system := range page.SystemSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceSystem, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(system.SystemArn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepServices(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	pages := resiliencehubv2.NewListServicesPaginator(conn, &resiliencehubv2.ListServicesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, svc := range page.ServiceSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceService, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(svc.ServiceArn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepInputSources(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	services := resiliencehubv2.NewListServicesPaginator(conn, &resiliencehubv2.ListServicesInput{})
	for services.HasMorePages() {
		page, err := services.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, svc := range page.ServiceSummaries {
			listInputSourcesInput := resiliencehubv2.ListInputSourcesInput{
				ServiceArn: svc.ServiceArn,
			}
			output, err := conn.ListInputSources(ctx, &listInputSourcesInput)
			if err != nil {
				continue
			}
			for _, is := range output.InputSourceSummaries {
				sweepResources = append(sweepResources, framework.NewSweepResource(newResourceInputSource, client,
					framework.NewAttribute(names.AttrID, aws.ToString(svc.ServiceArn)+","+aws.ToString(is.InputSourceId)),
					framework.NewAttribute("service_arn", aws.ToString(svc.ServiceArn)),
					framework.NewAttribute("input_source_id", aws.ToString(is.InputSourceId)),
				))
			}
		}
	}

	return sweepResources, nil
}

func sweepServiceFunctions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	services := resiliencehubv2.NewListServicesPaginator(conn, &resiliencehubv2.ListServicesInput{})
	for services.HasMorePages() {
		page, err := services.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, svc := range page.ServiceSummaries {
			listServiceFunctionsInput := resiliencehubv2.ListServiceFunctionsInput{
				ServiceArn: svc.ServiceArn,
			}
			output, err := conn.ListServiceFunctions(ctx, &listServiceFunctionsInput)
			if err != nil {
				continue
			}
			for _, sf := range output.ServiceFunctions {
				sweepResources = append(sweepResources, framework.NewSweepResource(newResourceServiceFunction, client,
					framework.NewAttribute(names.AttrID, aws.ToString(svc.ServiceArn)+","+aws.ToString(sf.ServiceFunctionId)),
					framework.NewAttribute("service_arn", aws.ToString(svc.ServiceArn)),
					framework.NewAttribute("service_function_id", aws.ToString(sf.ServiceFunctionId)),
				))
			}
		}
	}

	return sweepResources, nil
}

func sweepAssertions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	services := resiliencehubv2.NewListServicesPaginator(conn, &resiliencehubv2.ListServicesInput{})
	for services.HasMorePages() {
		page, err := services.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, svc := range page.ServiceSummaries {
			listAssertionsInput := resiliencehubv2.ListAssertionsInput{
				ServiceArn: svc.ServiceArn,
			}
			output, err := conn.ListAssertions(ctx, &listAssertionsInput)
			if err != nil {
				continue
			}
			for _, a := range output.Assertions {
				sweepResources = append(sweepResources, framework.NewSweepResource(newResourceAssertion, client,
					framework.NewAttribute(names.AttrID, aws.ToString(svc.ServiceArn)+","+aws.ToString(a.AssertionId)),
					framework.NewAttribute("service_arn", aws.ToString(svc.ServiceArn)),
					framework.NewAttribute("assertion_id", aws.ToString(a.AssertionId)),
				))
			}
		}
	}

	return sweepResources, nil
}

func sweepUserJourneys(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResilienceHubV2Client(ctx)

	var sweepResources []sweep.Sweepable

	systems := resiliencehubv2.NewListSystemsPaginator(conn, &resiliencehubv2.ListSystemsInput{})
	for systems.HasMorePages() {
		page, err := systems.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, sys := range page.SystemSummaries {
			listUserJourneysInput := resiliencehubv2.ListUserJourneysInput{
				SystemArn: sys.SystemArn,
			}
			output, err := conn.ListUserJourneys(ctx, &listUserJourneysInput)
			if err != nil {
				continue
			}
			for _, uj := range output.UserJourneySummaries {
				sweepResources = append(sweepResources, framework.NewSweepResource(newResourceUserJourney, client,
					framework.NewAttribute(names.AttrID, aws.ToString(sys.SystemArn)+","+aws.ToString(uj.UserJourneyId)),
					framework.NewAttribute("system_arn", aws.ToString(sys.SystemArn)),
					framework.NewAttribute("user_journey_id", aws.ToString(uj.UserJourneyId)),
				))
			}
		}
	}

	return sweepResources, nil
}
