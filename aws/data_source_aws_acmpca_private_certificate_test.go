package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsAcmpcaPrivateCertificate_Basic(t *testing.T) {
	resourceName := "aws_acmpca_private_certificate.test"
	datasourceName := "data.aws_acmpca_private_certificate.test"
	csr1, _ := tlsRsaX509CertificateRequestPem(4096, "terraformtest1.com")
	csr2, _ := tlsRsaX509CertificateRequestPem(4096, "terraformtest2.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsAcmpcaPrivateCertificateConfig_NonExistent,
				ExpectError: regexp.MustCompile(`(AccessDeniedException|ResourceNotFoundException)`),
			},
			{
				Config: testAccDataSourceAwsAcmpcaPrivateCertificateConfig_ARN(csr1, csr2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate", resourceName, "certificate"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_chain", resourceName, "certificate_chain"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_authority_arn", resourceName, "certificate_authority_arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsAcmpcaPrivateCertificateConfig_ARN(csr1, csr2 string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_private_certificate" "wrong" {
	certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
	certificate_signing_request = "%[1]s"
	signing_algorithm = "SHA256WITHRSA"
	validity_length = 1
	validity_unit = "YEARS"
}

resource "aws_acmpca_private_certificate" "test" {
	certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
	certificate_signing_request = "%[2]s"
	signing_algorithm = "SHA256WITHRSA"
	validity_length = 1
	validity_unit = "YEARS"
}

data "aws_acmpca_private_certificate" "test" {
	arn = aws_acmpca_private_certificate.test.arn
	certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}`, csr1, csr2)
}

const testAccDataSourceAwsAcmpcaPrivateCertificateConfig_NonExistent = `
data "aws_acmpca_private_certificate" "test" {
	arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/does-not-exist/certificate/does-not-exist"
	certificate_authority_arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/does-not-exist"
}
`
