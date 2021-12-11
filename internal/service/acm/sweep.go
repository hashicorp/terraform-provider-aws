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
	})
}

func sweepCertificates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ACMConn
	var sweeperErrs *multierror.Error

	err = conn.ListCertificatesPages(&acm.ListCertificatesInput{}, func(page *acm.ListCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, certificate := range page.CertificateSummaryList {
			arn := aws.StringValue(certificate.CertificateArn)

			output, err := conn.DescribeCertificate(&acm.DescribeCertificateInput{
				CertificateArn: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error describing ACM certificate (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if len(output.Certificate.InUseBy) > 0 {
				log.Printf("[INFO] ACM certificate (%s) is in-use, skipping", arn)
				continue
			}

			log.Printf("[INFO] Deleting ACM certificate: %s", arn)
			_, err = conn.DeleteCertificate(&acm.DeleteCertificateInput{
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
