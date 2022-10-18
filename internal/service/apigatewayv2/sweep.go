//go:build sweep
// +build sweep

package apigatewayv2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
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

	resource.AddTestSweepers("aws_apigatewayv2_api_mapping", &resource.Sweeper{
		Name: "aws_apigatewayv2_api_mapping",
		F:    sweepAPIMappings,
	})

	resource.AddTestSweepers("aws_apigatewayv2_stage", &resource.Sweeper{
		Name: "aws_apigatewayv2_stage",
		F:    sweepStages,
	})
}

func sweepAPIs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn
	input := &apigatewayv2.GetApisInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetApis(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway v2 API sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving API Gateway v2 APIs: %s", err)
		}

		for _, api := range output.Items {
			log.Printf("[INFO] Deleting API Gateway v2 API: %s", aws.StringValue(api.ApiId))
			_, err := conn.DeleteApi(&apigatewayv2.DeleteApiInput{
				ApiId: api.ApiId,
			})
			if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting API Gateway v2 API (%s): %w", aws.StringValue(api.ApiId), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDomainNames(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn
	input := &apigatewayv2.GetDomainNamesInput{}
	var sweeperErrs *multierror.Error

	err = getDomainNamesPages(conn, input, func(page *apigatewayv2.GetDomainNamesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, domainName := range page.Items {
			r := ResourceDomainName()
			d := r.Data(nil)
			d.SetId(aws.StringValue(domainName.DomainName))
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping API Gateway v2 domain names sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing API Gateway v2 domain names: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepVPCLinks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn
	input := &apigatewayv2.GetVpcLinksInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetVpcLinks(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway v2 VPC Link sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving API Gateway v2 VPC Links: %s", err)
		}

		for _, link := range output.Items {
			log.Printf("[INFO] Deleting API Gateway v2 VPC Link: %s", aws.StringValue(link.VpcLinkId))
			_, err := conn.DeleteVpcLink(&apigatewayv2.DeleteVpcLinkInput{
				VpcLinkId: link.VpcLinkId,
			})
			if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting API Gateway v2 VPC Link (%s): %w", aws.StringValue(link.VpcLinkId), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = WaitVPCLinkDeleted(conn, aws.StringValue(link.VpcLinkId))
			if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error waiting for API Gateway v2 VPC Link (%s) deletion: %w", aws.StringValue(link.VpcLinkId), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepAPIMappings(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn
	var sweeperErrs *multierror.Error
	log.Printf("[INFO] API Gateway v2 API Mapping")

	err = getDomainNamesPages(conn, &apigatewayv2.GetDomainNamesInput{}, func(page *apigatewayv2.GetDomainNamesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, domainName := range page.Items {
			err = getAPIMappingsPages(conn, &apigatewayv2.GetApiMappingsInput{
				DomainName: domainName.DomainName,
			}, func(page *apigatewayv2.GetApiMappingsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, apiMapping := range page.Items {
					log.Printf("[INFO] API Gateway v2 API Mapping: %+v", apiMapping)
				}

				return !lastPage
			})

		}

		return !lastPage
	})

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing API Gateway v2 domain names: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepStages(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayV2Conn
	var sweeperErrs *multierror.Error
	log.Printf("[INFO] API Gateway v2 Stages")

	err = getAPIsPages(conn, &apigatewayv2.GetApisInput{}, func(page *apigatewayv2.GetApisOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, api := range page.Items {
			err = getStagesPages(conn, &apigatewayv2.GetStagesInput{
				ApiId: api.ApiId,
			}, func(page *apigatewayv2.GetStagesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, stage := range page.Items {
					log.Printf("[INFO] API Gateway v2 Stage: %+v", stage)
				}

				return !lastPage
			})

		}

		return !lastPage
	})

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing API Gateway v2 APIs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
