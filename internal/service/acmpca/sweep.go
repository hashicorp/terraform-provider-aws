//go:build sweep
// +build sweep

package acmpca

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_acmpca_certificate_authority", &resource.Sweeper{
		Name: "aws_acmpca_certificate_authority",
		F:    sweepCertificateAuthorities,
	})
}

func sweepCertificateAuthorities(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ACMPCAConn

	certificateAuthorities, err := listCertificateAuthorities(conn)
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ACM PCA Certificate Authorities sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("retrieving ACM PCA Certificate Authorities: %w", err)
	}
	if len(certificateAuthorities) == 0 {
		log.Print("[DEBUG] No ACM PCA Certificate Authorities to sweep")
		return nil
	}

	var sweeperErrs *multierror.Error

	for _, certificateAuthority := range certificateAuthorities {
		arn := aws.StringValue(certificateAuthority.Arn)

		if aws.StringValue(certificateAuthority.Status) == acmpca.CertificateAuthorityStatusActive {
			log.Printf("[INFO] Disabling ACM PCA Certificate Authority: %s", arn)
			_, err := conn.UpdateCertificateAuthority(&acmpca.UpdateCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(arn),
				Status:                  aws.String(acmpca.CertificateAuthorityStatusDisabled),
			})
			if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error disabling ACM PCA Certificate Authority (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		log.Printf("[INFO] Deleting ACM PCA Certificate Authority: %s", arn)
		_, err := conn.DeleteCertificateAuthority(&acmpca.DeleteCertificateAuthorityInput{
			CertificateAuthorityArn:     aws.String(arn),
			PermanentDeletionTimeInDays: aws.Int64(7),
		})
		if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting ACM PCA Certificate Authority (%s): %w", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func listCertificateAuthorities(conn *acmpca.ACMPCA) ([]*acmpca.CertificateAuthority, error) {
	certificateAuthorities := []*acmpca.CertificateAuthority{}
	input := &acmpca.ListCertificateAuthoritiesInput{}

	for {
		output, err := conn.ListCertificateAuthorities(input)
		if err != nil {
			return certificateAuthorities, err
		}
		certificateAuthorities = append(certificateAuthorities, output.CertificateAuthorities...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return certificateAuthorities, nil
}
