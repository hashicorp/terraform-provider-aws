// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package apigatewayv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_apigatewayv2_api", &resource.Sweeper{
		Name: "aws_apigatewayv2_api",
		F:    sweepAPIs,
		Dependencies: []string{
			"aws_apigatewayv2_domain_name",
		},
	})

	resource.AddTestSweepers("aws_apigatewayv2_api_mapping", &resource.Sweeper{
		Name: "aws_apigatewayv2_api_mapping",
		F:    sweepAPIMappings,
	})

	resource.AddTestSweepers("aws_apigatewayv2_domain_name", &resource.Sweeper{
		Name: "aws_apigatewayv2_domain_name",
		F:    sweepDomainNames,
		Dependencies: []string{
			"aws_apigatewayv2_api_mapping",
		},
	})

	resource.AddTestSweepers("aws_apigatewayv2_vpc_link", &resource.Sweeper{
		Name: "aws_apigatewayv2_vpc_link",
		F:    sweepVPCLinks,
	})
}

func sweepAPIs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.APIGatewayV2Conn(ctx)
	input := &apigatewayv2.GetApisInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getAPIsPages(ctx, conn, input, func(page *apigatewayv2.GetApisOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			r := ResourceAPI()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ApiId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping API Gateway v2 API sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing API Gateway v2 APIs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway v2 APIs (%s): %w", region, err)
	}

	return nil
}

func sweepAPIMappings(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.APIGatewayV2Conn(ctx)
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &apigatewayv2.GetDomainNamesInput{}
	err = getDomainNamesPages(ctx, conn, input, func(page *apigatewayv2.GetDomainNamesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			domainName := aws.StringValue(v.DomainName)
			input := &apigatewayv2.GetApiMappingsInput{
				DomainName: aws.String(domainName),
			}

			err := getAPIMappingsPages(ctx, conn, input, func(page *apigatewayv2.GetApiMappingsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Items {
					r := ResourceAPIMapping()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.ApiMappingId))
					d.Set("domain_name", domainName)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing API Gateway v2 API Mappings (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping API Gateway v2 API Mapping sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing API Gateway v2 Domain Names (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping API Gateway v2 API Mappings (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDomainNames(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.APIGatewayV2Conn(ctx)
	input := &apigatewayv2.GetDomainNamesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getDomainNamesPages(ctx, conn, input, func(page *apigatewayv2.GetDomainNamesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			r := ResourceDomainName()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DomainName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping API Gateway v2 Domain Name sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing API Gateway v2 Domain Names (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway v2 Domain Names (%s): %w", region, err)
	}

	return nil
}

func sweepVPCLinks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.APIGatewayV2Conn(ctx)
	input := &apigatewayv2.GetVpcLinksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = getVPCLinksPages(ctx, conn, input, func(page *apigatewayv2.GetVpcLinksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			r := ResourceVPCLink()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.VpcLinkId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping API Gateway v2 VPC Link sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing API Gateway v2 VPC Links (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway v2 VPC Links (%s): %w", region, err)
	}

	return nil
}
