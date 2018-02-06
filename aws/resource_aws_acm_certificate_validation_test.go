package aws

import (
	"fmt"
	"testing"

	"os"
	"regexp"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAwsAcmResource_certificateIssuingAndValidationFlow(t *testing.T) {
	if os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN") == "" {
		t.Skip("Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set")
	}

	root_zone_domain := os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN")

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, root_zone_domain)
	sanDomain := fmt.Sprintf("tf-acc-%d-san.%s", rInt1, root_zone_domain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that validation times out if certificate can't be validated
			resource.TestStep{
				Config:      testAccAcmCertificateWithValidationConfig(domain, sanDomain),
				ExpectError: regexp.MustCompile("Expected certificate to be issued but was in state PENDING_VALIDATION"),
			},
			// Test that validation fails if given validation_fqdns don't match
			resource.TestStep{
				Config:      testAccAcmCertificateWithValidationConfigAndWrongFQDN(domain, sanDomain),
				ExpectError: regexp.MustCompile("Certificate needs .* to be set but only .* was passed to validation_record_fqdns"),
			},
			// Test that validation succeeds once we provide the right DNS validation records
			resource.TestStep{
				Config: testAccAcmCertificateWithValidationAndRecordsConfig(root_zone_domain, domain, sanDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
			// Test that we can import a validated certificate
			resource.TestStep{
				ResourceName:      "aws_acm_certificate.cert",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAcmCertificateWithValidationConfig(domain string, sanDomain string) string {
	return fmt.Sprintf(`
%s

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  timeout = "20s"
}
`, testAccAcmCertificateConfig(
		domain, sanDomain,
		"Environment", "Test",
		"Foo", "Baz"))
}

func testAccAcmCertificateWithValidationConfigAndWrongFQDN(domain string, sanDomain string) string {
	return fmt.Sprintf(`
%s

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = ["some-wrong-fqdn.example.com"]
  timeout = "20s"
}
`, testAccAcmCertificateConfig(
		domain, sanDomain,
		"Environment", "Test",
		"Foo", "Baz"))
}

func testAccAcmCertificateWithValidationAndRecordsConfig(rootZoneDomain string, domain string, sanDomain string) string {
	return fmt.Sprintf(`
%s

data "aws_route53_zone" "zone" {
  name = "%s."
  private_zone = false
}

resource "aws_route53_record" "cert_validation" {
  name = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl = 60
}

resource "aws_route53_record" "cert_validation_san" {
  name = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.1.resource_record_value}"]
  ttl = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = [
	"${aws_route53_record.cert_validation.fqdn}",
	"${aws_route53_record.cert_validation_san.fqdn}"
  ]
}
`, testAccAcmCertificateConfig(
		domain, sanDomain,
		"Environment", "Test",
		"Foo", "Baz"),
		rootZoneDomain)
}
