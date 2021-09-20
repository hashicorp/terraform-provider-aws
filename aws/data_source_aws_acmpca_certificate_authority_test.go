package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccDataSourceAwsAcmpcaCertificateAuthority_basic(t *testing.T) {
	resourceName := "aws_acmpca_certificate_authority.test"
	datasourceName := "data.aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_NonExistent,
				ExpectError: regexp.MustCompile(`(AccessDeniedException|ResourceNotFoundException)`),
			},
			{
				Config: testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_ARN(commonName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate", resourceName, "certificate"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_chain", resourceName, "certificate_chain"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_signing_request", resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_after", resourceName, "not_after"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_before", resourceName, "not_before"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.#", resourceName, "revocation_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.#", resourceName, "revocation_configuration.0.crl_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.enabled", resourceName, "revocation_configuration.0.crl_configuration.0.enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "serial", resourceName, "serial"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "type", resourceName, "type"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAcmpcaCertificateAuthority_S3ObjectAcl(t *testing.T) {
	resourceName := "aws_acmpca_certificate_authority.test"
	datasourceName := "data.aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_NonExistent,
				ExpectError: regexp.MustCompile(`(AccessDeniedException|ResourceNotFoundException)`),
			},
			{
				Config: testAccDataSourceAwsAcmpcaCertificateAuthorityConfigS3ObjectAcl_ARN(commonName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate", resourceName, "certificate"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_chain", resourceName, "certificate_chain"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_signing_request", resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_after", resourceName, "not_after"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_before", resourceName, "not_before"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.#", resourceName, "revocation_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.#", resourceName, "revocation_configuration.0.crl_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.enabled", resourceName, "revocation_configuration.0.crl_configuration.0.enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl"),
					resource.TestCheckResourceAttrPair(datasourceName, "serial", resourceName, "serial"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "type", resourceName, "type"),
				),
			},
		},
	})
}

func testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_ARN(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "wrong" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_acmpca_certificate_authority" "test" {
  arn = aws_acmpca_certificate_authority.test.arn
}
`, commonName)
}

func testAccDataSourceAwsAcmpcaCertificateAuthorityConfigS3ObjectAcl_ARN(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "wrong" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

data "aws_acmpca_certificate_authority" "test" {
  arn = aws_acmpca_certificate_authority.test.arn
}
`, commonName)
}

//lintignore:AWSAT003,AWSAT005
const testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_NonExistent = `
data "aws_acmpca_certificate_authority" "test" {
  arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/tf-acc-test-does-not-exist"
}
`
