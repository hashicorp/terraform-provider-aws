//go:build sweep
// +build sweep

package apigatewayv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

	resource.AddTestSweepers("aws_apigatewayv2_domain_name", &resource.Sweeper{
		Name: "aws_apigatewayv2_domain_name",
		F:    sweepDomainNames,
	})

	resource.AddTestSweepers("aws_apigatewayv2_vpc_link", &resource.Sweeper{
		Name: "aws_apigatewayv2_vpc_link",
		F:    sweepVPCLinks,
	})
}

func sweepAPIs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn()
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

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway v2 APIs (%s): %w", region, err)
	}

	return nil
}

func sweepDomainNames(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn()
	sweepResources := make([]sweep.Sweepable, 0)

	input := &apigatewayv2.GetDomainNamesInput{}
	err = getDomainNamesPages(ctx, conn, input, func(page *apigatewayv2.GetDomainNamesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, domainName := range page.Items {
			r := ResourceDomainName()
			d := r.Data(nil)
			d.SetId(aws.StringValue(domainName.DomainName))

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

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway v2 Domain Names (%s): %w", region, err)
	}

	return nil
}

func sweepVPCLinks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn()
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

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping API Gateway v2 VPC Links (%s): %w", region, err)
	}

	return nil
}
