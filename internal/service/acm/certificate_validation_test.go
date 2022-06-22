package acm_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfacm "github.com/hashicorp/terraform-provider-aws/internal/service/acm"
)

func TestAccACMCertificateValidation_basic(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_validation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that validation succeeds
			{
				Config: testAccCertificateValidationConfig_basic(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateValidationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", certificateResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccACMCertificateValidation_timeout(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateValidationConfig_timeout(domain),
				ExpectError: regexp.MustCompile(`timeout while waiting for state to become 'ISSUED' \(last state: 'PENDING_VALIDATION'`),
			},
		},
	})
}

func TestAccACMCertificateValidation_validationRecordFQDNS(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_validation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that validation fails if given validation_fqdns don't match
			{
				Config:      testAccCertificateValidationConfig_recordFQDNsWrongFQDN(domain),
				ExpectError: regexp.MustCompile("missing .+ DNS validation record: .+"),
			},
			// Test that validation succeeds with validation
			{
				Config: testAccCertificateValidationConfig_recordFQDNsOneRoute53Record(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateValidationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", certificateResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccACMCertificateValidation_validationRecordFQDNSEmail(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCertificateValidationConfig_recordFQDNsEmail(domain),
				ExpectError: regexp.MustCompile("validation_record_fqdns is not valid for EMAIL validation"),
			},
		},
	})
}

func TestAccACMCertificateValidation_validationRecordFQDNSRoot(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_validation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateValidationConfig_recordFQDNsOneRoute53Record(rootDomain, rootDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateValidationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", certificateResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccACMCertificateValidation_validationRecordFQDNSRootAndWildcard(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_validation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateValidationConfig_recordFQDNsTwoRoute53Records(rootDomain, rootDomain, strconv.Quote(wildcardDomain)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateValidationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", certificateResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccACMCertificateValidation_validationRecordFQDNSSan(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	sanDomain := acctest.ACMCertificateRandomSubDomain(rootDomain)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_validation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateValidationConfig_recordFQDNsTwoRoute53Records(rootDomain, domain, strconv.Quote(sanDomain)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateValidationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", certificateResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccACMCertificateValidation_validationRecordFQDNSWildcard(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_validation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateValidationConfig_recordFQDNsOneRoute53Record(rootDomain, wildcardDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateValidationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", certificateResourceName, "arn"),
				),
				// ExpectNonEmptyPlan: true, // https://github.com/hashicorp/terraform-provider-aws/issues/16913
			},
		},
	})
}

func TestAccACMCertificateValidation_validationRecordFQDNSWildcardAndRoot(t *testing.T) {
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)
	certificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_acm_certificate_validation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, acm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateValidationConfig_recordFQDNsTwoRoute53Records(rootDomain, wildcardDomain, strconv.Quote(rootDomain)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCertificateValidationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", certificateResourceName, "arn"),
				),
				// ExpectNonEmptyPlan: true, // https://github.com/hashicorp/terraform-provider-aws/issues/16913
			},
		},
	})
}

func testAccCheckCertificateValidationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ACM Certificate Validation ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ACMConn

		_, err := tfacm.FindCertificateValidationByARN(conn, rs.Primary.Attributes["certificate_arn"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCertificateValidationConfig_basic(rootZoneDomain, domainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

#
# for_each acceptance testing requires SDKv2
#
# resource "aws_route53_record" "test" {
#   for_each = {
#     for dvo in aws_acm_certificate.test.domain_validation_options: dvo.domain_name => {
#       name   = dvo.resource_record_name
#       record = dvo.resource_record_value
#       type   = dvo.resource_record_type
#     }
#   }

#   allow_overwrite = true
#   name            = each.value.name
#   records         = [each.value.record]
#   ttl             = 60
#   type            = each.value.type
#   zone_id         = data.aws_route53_zone.test.zone_id
# }

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  depends_on = [aws_route53_record.test]

  certificate_arn = aws_acm_certificate.test.arn
}
`, domainName, rootZoneDomain)
}

func testAccCertificateValidationConfig_timeout(domainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn = aws_acm_certificate.test.arn

  timeouts {
    create = "5s"
  }
}
`, domainName)
}

func testAccCertificateValidationConfig_recordFQDNsEmail(domainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = "EMAIL"
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = ["wrong-validation-fqdn.example.com"]
}
`, domainName)
}

func testAccCertificateValidationConfig_recordFQDNsOneRoute53Record(rootZoneDomain, domainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

#
# for_each acceptance testing requires SDKv2
#
# resource "aws_route53_record" "test" {
#   for_each = {
#     for dvo in aws_acm_certificate.test.domain_validation_options: dvo.domain_name => {
#       name   = dvo.resource_record_name
#       record = dvo.resource_record_value
#       type   = dvo.resource_record_type
#     }
#   }

#   allow_overwrite = true
#   name            = each.value.name
#   records         = [each.value.record]
#   ttl             = 60
#   type            = each.value.type
#   zone_id         = data.aws_route53_zone.test.zone_id
# }

# resource "aws_acm_certificate_validation" "test" {
#   certificate_arn         = aws_acm_certificate.test.arn
#   validation_record_fqdns = [for record in aws_route53_record.test: record.fqdn]
# }

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}
`, domainName, rootZoneDomain)
}

func testAccCertificateValidationConfig_recordFQDNsTwoRoute53Records(rootZoneDomain, domainName, subjectAlternativeNames string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[1]q
  subject_alternative_names = [%[2]s]
  validation_method         = "DNS"
}

data "aws_route53_zone" "test" {
  name         = %[3]q
  private_zone = false
}

#
# for_each acceptance testing requires SDKv2
#
# resource "aws_route53_record" "test" {
#   for_each = {
#     for dvo in aws_acm_certificate.test.domain_validation_options: dvo.domain_name => {
#       name   = dvo.resource_record_name
#       record = dvo.resource_record_value
#       type   = dvo.resource_record_type
#     }
#   }

#   allow_overwrite = true
#   name            = each.value.name
#   records         = [each.value.record]
#   ttl             = 60
#   type            = each.value.type
#   zone_id         = data.aws_route53_zone.test.zone_id
# }

# resource "aws_acm_certificate_validation" "test" {
#   certificate_arn         = aws_acm_certificate.test.arn
#   validation_record_fqdns = [for record in aws_route53_record.test: record.fqdn]
# }

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_route53_record" "test2" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[1].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[1].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[1].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn, aws_route53_record.test2.fqdn]
}
`, domainName, subjectAlternativeNames, rootZoneDomain)
}

func testAccCertificateValidationConfig_recordFQDNsWrongFQDN(domainName string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name       = %[1]q
  validation_method = "DNS"
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = ["wrong-validation-fqdn.example.com"]
}
`, domainName)
}
