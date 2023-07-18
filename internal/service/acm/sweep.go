// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package acm

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
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
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ACMClient(ctx)
	input := &acm.ListCertificatesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := acm.NewListCertificatesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ACM Certificate sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ACM Certificates (%s): %w", region, err)
		}

		for _, v := range page.CertificateSummaryList {
			arn := aws.ToString(v.CertificateArn)
			input := &acm.DescribeCertificateInput{
				CertificateArn: aws.String(arn),
			}
			certificate, err := findCertificate(ctx, conn, input)

			if err != nil {
				log.Printf("[ERROR] Reading ACM Certificate (%s): %s", arn, err)
				continue
			}

			if n := len(certificate.InUseBy); n > 0 {
				log.Printf("[INFO] ACM Certificate (%s) skipped, in use by, e.g., (%d tot):", arn, n)
				m := make(map[string]string)
				for _, iub := range certificate.InUseBy {
					if len(iub) < 77 {
						m[iub] = ""
					} else {
						m[iub[:77]] = ""
					}
				}
				for k := range m {
					log.Printf("[INFO]  %s...", k)
				}
				continue
			}

			r := resourceCertificate()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ACM Certificates (%s): %w", region, err)
	}

	return nil
}
