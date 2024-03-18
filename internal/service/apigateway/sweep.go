// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
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
	conn := client.APIGatewayClient(ctx)

	pages := apigateway.NewGetRestApisPaginator(conn, &apigateway.GetRestApisInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			if awsv1.SkipSweepError(err) {
				log.Printf("[WARN] Skipping API Gateway REST API sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("retrieving API Gateway REST APIs: %s", err)
		}

		for _, item := range page.Items {
			input := &apigateway.DeleteRestApiInput{
				RestApiId: item.Id,
			}
			log.Printf("[INFO] Deleting API Gateway REST API: %+v", input)
			// TooManyRequestsException: Too Many Requests can take over a minute to resolve itself
			err := retry.RetryContext(ctx, 2*time.Minute, func() *retry.RetryError {
				_, err := conn.DeleteRestApi(ctx, input)
				if err != nil {
					if errs.IsA[*awstypes.TooManyRequestsException](err) {
						return retry.RetryableError(err)
					}
					return retry.NonRetryableError(err)
				}
				return nil
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete API Gateway REST API %s: %s", *item.Name, err)
				continue
			}
		}
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
		if err != nil {
			if awsv1.SkipSweepError(err) {
				log.Printf("[WARN] Skipping API Gateway VPC Link sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("error listing API Gateway VPC Links (%s): %w", region, err)
		}

		for _, v := range page.Items {
			id := aws.ToString(v.Id)

			if v.Status == awstypes.VpcLinkStatusFailed {
				log.Printf("[INFO] Skipping API Gateway VPC Link %s: Status=%s", id, string(v.Status))
				continue
			}

			r := ResourceVPCLink()
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
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := apigateway.NewGetClientCertificatesPaginator(conn, &apigateway.GetClientCertificatesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("describing API Gateway Client Certificates for %s: %w", region, err))
		}

		for _, clientCertificate := range page.Items {
			r := ResourceClientCertificate()
			d := r.Data(nil)
			d.SetId(aws.ToString(clientCertificate.ClientCertificateId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping API Gateway Client Certificates for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping API Gateway Client Certificate sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepUsagePlans(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	log.Printf("[INFO] Sweeping API Gateway Usage Plans for %s", region)

	conn := client.APIGatewayClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := apigateway.NewGetUsagePlansPaginator(conn, &apigateway.GetUsagePlansInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("describing API Gateway Usage Plans for %s: %w", region, err))
		}

		log.Printf("[INFO] API Gateway Usage Plans: %d", len(page.Items))

		for _, up := range page.Items {
			r := ResourceUsagePlan()
			d := r.Data(nil)
			d.SetId(aws.ToString(up.Id))
			d.Set("api_stages", flattenAPIStages(up.ApiStages))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping API Gateway Usage Plans for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping API Gateway Usage Plan sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepAPIKeys(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	log.Printf("[INFO] Sweeping API Gateway API Keys for %s", region)

	conn := client.APIGatewayClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := apigateway.NewGetApiKeysPaginator(conn, &apigateway.GetApiKeysInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("describing API Gateway API Keys for %s: %w", region, err))
		}

		log.Printf("[INFO] API Gateway API Keys: %d", len(page.Items))

		for _, ak := range page.Items {
			r := ResourceAPIKey()
			d := r.Data(nil)
			d.SetId(aws.ToString(ak.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping API Gateway API Keys for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping API Gateway API Key sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepDomainNames(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}

	log.Printf("[INFO] Sweeping API Gateway Domain Names for %s", region)

	conn := client.APIGatewayClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := apigateway.NewGetDomainNamesPaginator(conn, &apigateway.GetDomainNamesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("describing API Gateway Domain Names for %s: %w", region, err))
		}

		log.Printf("[INFO] API Gateway Domain Names: %d", len(page.Items))

		for _, dn := range page.Items {
			r := ResourceDomainName()
			d := r.Data(nil)
			d.SetId(aws.ToString(dn.DomainName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping API Gateway Domain Names for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping API Gateway Domain Name sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
