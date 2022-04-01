//go:build sweep
// +build sweep

package apigateway

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_api_gateway_rest_api", &resource.Sweeper{
		Name: "aws_api_gateway_rest_api",
		F:    sweepRestAPIs,
	})

	resource.AddTestSweepers("aws_api_gateway_vpc_link", &resource.Sweeper{
		Name: "aws_api_gateway_vpc_link",
		F:    sweepVPCLinks,
	})
}

func sweepRestAPIs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayConn

	err = conn.GetRestApisPages(&apigateway.GetRestApisInput{}, func(page *apigateway.GetRestApisOutput, lastPage bool) bool {
		for _, item := range page.Items {
			input := &apigateway.DeleteRestApiInput{
				RestApiId: item.Id,
			}
			log.Printf("[INFO] Deleting API Gateway REST API: %s", input)
			// TooManyRequestsException: Too Many Requests can take over a minute to resolve itself
			err := resource.Retry(2*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteRestApi(input)
				if err != nil {
					if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeTooManyRequestsException) {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete API Gateway REST API %s: %s", *item.Name, err)
				continue
			}
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway REST API sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving API Gateway REST APIs: %s", err)
	}

	return nil
}

func sweepVPCLinks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).APIGatewayConn

	sweepResources := make([]*sweep.SweepResource, 0)
	var sweeperErrs *multierror.Error

	err = conn.GetVpcLinksPages(&apigateway.GetVpcLinksInput{}, func(page *apigateway.GetVpcLinksOutput, lastPage bool) bool {
		for _, item := range page.Items {
			id := aws.StringValue(item.Id)

			log.Printf("[INFO] Deleting API Gateway VPC Link (%s)", id)
			r := ResourceVPCLink()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping API Gateway VPC Link sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error retrieving API Gateway VPC Links: %w", err)
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping API Gateway VPC Links: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
