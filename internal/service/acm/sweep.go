//go:build sweep
// +build sweep

package acm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_acm_certificate", &resource.Sweeper{
		Name: "aws_acm_certificate",
		F:    sweepCertificates,
		Dependencies: []string{
			"aws_api_gateway_api_key",
			"aws_api_gateway_client_certificate",
			"aws_api_gateway_domain_name",
			"aws_api_gateway_rest_api",
			"aws_api_gateway_usage_plan",
			"aws_api_gateway_vpc_link",
			"aws_apigatewayv2_api",
			"aws_apigatewayv2_api_mapping",
			"aws_apigatewayv2_domain_name",
			"aws_apigatewayv2_stage",
			"aws_apigatewayv2_vpc_link",
			"aws_elb",
			"aws_iam_server_certificate",
			"aws_iam_signing_certificate",
			"aws_lb",
			"aws_lb_listener",
		},
	})
}

func sweepCertificates(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ACMConn()
	var sweeperErrs *multierror.Error

	err = conn.ListCertificatesPagesWithContext(ctx, &acm.ListCertificatesInput{}, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, certificate := range page.CertificateSummaryList {
			arn := aws.StringValue(certificate.CertificateArn)

			output, err := conn.DescribeCertificateWithContext(ctx, &acm.DescribeCertificateInput{
				CertificateArn: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error describing ACM certificate (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if len(output.Certificate.InUseBy) > 0 {
				log.Printf("[INFO] ACM Certificate (%s) skipped, in use by, e.g., (%d tot):", arn, len(output.Certificate.InUseBy))
				m := make(map[string]string)
				for _, iub := range output.Certificate.InUseBy {
					if len(aws.StringValue(iub)) < 77 {
						m[aws.StringValue(iub)] = ""
					} else {
						m[aws.StringValue(iub)[:77]] = ""
					}
				}
				for k := range m {
					log.Printf("[INFO]  %s...", k)
				}
				continue
			}

			log.Printf("[INFO] Deleting ACM certificate: %s", arn)
			_, err = conn.DeleteCertificateWithContext(ctx, &acm.DeleteCertificateInput{
				CertificateArn: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting ACM certificate (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ACM certificate sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		return fmt.Errorf("error retrieving ACM certificates: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}
