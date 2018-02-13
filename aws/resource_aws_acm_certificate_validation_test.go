package aws

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAwsAcmCertificateValidation_basic(t *testing.T) {
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
				Config:      testAccAcmCertificateValidation_basic(domain),
				ExpectError: regexp.MustCompile("Expected certificate to be issued but was in state PENDING_VALIDATION"),
			},
			// Test that validation fails if given validation_fqdns don't match
			resource.TestStep{
				Config:      testAccAcmCertificateValidation_validationRecordFqdns(domain, acm.ValidationMethodDns, `"some-wrong-fqdn.example.com"`),
				ExpectError: regexp.MustCompile("Certificate needs .* to be set but only .* was passed to validation_record_fqdns"),
			},
			// Test that validation fails if not DNS validation
			resource.TestStep{
				Config:      testAccAcmCertificateValidation_validationRecordFqdns(domain, acm.ValidationMethodEmail, `"some-wrong-fqdn.example.com"`),
				ExpectError: regexp.MustCompile("validation_record_fqdns is only valid for DNS validation"),
			},
			// Test that validation succeeds once we provide the right DNS validation records
			resource.TestStep{
				Config: testAccAcmCertificateValidation_withRoute53Records(root_zone_domain, domain, strconv.Quote(sanDomain)),
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

func testAccAcmCertificateValidation_basic(domainName string) string {
	return fmt.Sprintf(`
%s

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  timeouts {
    create = "20s"
  }
}
`, testAccAcmCertificateConfig(domainName, acm.ValidationMethodDns))
}

func testAccAcmCertificateValidation_validationRecordFqdns(domainName, validationMethod, validationRecordFqdns string) string {
	return fmt.Sprintf(`
%s

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = [%s]
  timeouts {
    create = "20s"
  }
}
`, testAccAcmCertificateConfig(domainName, validationMethod), validationRecordFqdns)
}

func testAccAcmCertificateValidation_withRoute53Records(rootZoneDomain, domainName, subjectAlternativeNames string) string {
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
`, testAccAcmCertificateConfig_subjectAlternativeNames(domainName, subjectAlternativeNames, acm.ValidationMethodDns), rootZoneDomain)
}
