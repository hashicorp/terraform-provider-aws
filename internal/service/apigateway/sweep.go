// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_api_gateway_rest_api", &resource.Sweeper{
		Name: "aws_api_gateway_rest_api",
		F:    sweepRestAPIs,
	})

	resource.AddTestSweepers("aws_api_gateway_vpc_link", &resource.Sweeper{
		Name: "aws_api_gateway_vpc_link",
		F:    sweepVPCLinks,
	})

	resource.AddTestSweepers("aws_api_gateway_client_certificate", &resource.Sweeper{
		Name: "aws_api_gateway_client_certificate",
		F:    sweepClientCertificates,
	})

	resource.AddTestSweepers("aws_api_gateway_usage_plan", &resource.Sweeper{
		Name: "aws_api_gateway_usage_plan",
		F:    sweepUsagePlans,
	})

	resource.AddTestSweepers("aws_api_gateway_api_key", &resource.Sweeper{
		Name: "aws_api_gateway_api_key",
		F:    sweepAPIKeys,
		Dependencies: []string{
			"aws_api_gateway_usage_plan",
		},
	})

	resource.AddTestSweepers("aws_api_gateway_domain_name", &resource.Sweeper{
		Name: "aws_api_gateway_domain_name",
		F:    sweepDomainNames,
	})
}

func sweepRestAPIs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &apigateway.GetRestApisInput{}
	conn := client.APIGatewayClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apigateway.NewGetRestApisPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway REST API sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing API Gateway REST APIs (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := resourceRestAPI()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway REST APIs (%s): %w", region, err)
	}

	return nil
}

func sweepVPCLinks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.APIGatewayClient(ctx)
	input := &apigateway.GetVpcLinksInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apigateway.NewGetVpcLinksPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway VPC Link sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing API Gateway VPC Links (%s): %w", region, err)
		}

		for _, v := range page.Items {
			id := aws.ToString(v.Id)

			if v.Status == types.VpcLinkStatusFailed {
				log.Printf("[INFO] Skipping API Gateway VPC Link %s: Status=%s", id, string(v.Status))
				continue
			}

			r := resourceVPCLink()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway VPC Links (%s): %w", region, err)
	}

	return nil
}

func sweepClientCertificates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.APIGatewayClient(ctx)
	input := &apigateway.GetClientCertificatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apigateway.NewGetClientCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway Client Certificate sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing API Gateway Client Certificates (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := resourceClientCertificate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ClientCertificateId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway Client Certificates (%s): %w", region, err)
	}

	return nil
}

func sweepUsagePlans(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.APIGatewayClient(ctx)
	input := &apigateway.GetUsagePlansInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apigateway.NewGetUsagePlansPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway Usage Plan sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing API Gateway Usage Plans (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := resourceUsagePlan()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))
			d.Set("api_stages", flattenAPIStages(v.ApiStages))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway Usage Plans (%s): %w", region, err)
	}

	return nil
}

func sweepAPIKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.APIGatewayClient(ctx)
	input := &apigateway.GetApiKeysInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apigateway.NewGetApiKeysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway API Key sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing API Gateway API Keys (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := resourceAPIKey()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway API Keys (%s): %w", region, err)
	}

	return nil
}

func sweepDomainNames(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.APIGatewayClient(ctx)
	input := &apigateway.GetDomainNamesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apigateway.NewGetDomainNamesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway Domain Name sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing API Gateway Domain Names (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := resourceDomainName()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DomainName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway Domain Names (%s): %w", region, err)
	}

	return nil
}
