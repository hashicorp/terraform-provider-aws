package aws

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccAWSAcmCertificateImported_selfSigned(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAcmCertificateConfig_selfSigned("example"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_imported.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate_imported.cert", "domain_name", "example.com"),
					resource.TestCheckResourceAttr("aws_acm_certificate_imported.cert", "subject_alternative_names.#", "0"),
				),
			},
			resource.TestStep{
				Config: testAccAcmCertificateConfig_selfSigned("example2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_imported.cert", "arn", certificateArnRegex),
					resource.TestCheckResourceAttr("aws_acm_certificate_imported.cert", "domain_name", "example.com"),
					resource.TestCheckResourceAttr("aws_acm_certificate_imported.cert", "subject_alternative_names.#", "0"),
				),
			},
		},
	})

}

func testAccAcmCertificateConfig_selfSigned(certName string) string {
	return fmt.Sprintf(`
	resource "tls_private_key" "%[1]s" {
		algorithm = "RSA"
	}
	
	resource "tls_self_signed_cert" "%[1]s" {
		key_algorithm   = "RSA"
		private_key_pem = "${tls_private_key.%[1]s.private_key_pem}"
	
		subject {
			common_name  = "example.com"
			organization = "ACME Examples, Inc"
		}
	
		validity_period_hours = 12
	
		allowed_uses = [
			"key_encipherment",
			"digital_signature",
			"server_auth",
		]
	}
	
	resource "aws_acm_certificate_imported" "cert" {
	  private_key       = "${tls_private_key.%[1]s.private_key_pem}"
	  certificate_body  = "${tls_self_signed_cert.%[1]s.cert_pem}"
	  depends_on = ["tls_private_key.%[1]s", "tls_self_signed_cert.%[1]s"]
	}
	`, certName)
}
