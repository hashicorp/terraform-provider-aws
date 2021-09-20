package route53_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
)

func TestAccAwsRoute53HostedZoneDnssec_basic(t *testing.T) {
	route53ZoneResourceName := "aws_route53_zone.test"
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53HostedZoneDnssecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "hosted_zone_id", route53ZoneResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureSigning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsRoute53HostedZoneDnssec_disappears(t *testing.T) {
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53HostedZoneDnssecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceHostedZoneDNSSEC(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsRoute53HostedZoneDnssec_SigningStatus(t *testing.T) {
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckRoute53KeySigningKey(t) },
		ErrorCheck:        acctest.ErrorCheck(t, route53.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsRoute53HostedZoneDnssecDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName, domainName, tfroute53.ServeSignatureNotSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureNotSigning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName, domainName, tfroute53.ServeSignatureSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureSigning),
				),
			},
			{
				Config: testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName, domainName, tfroute53.ServeSignatureNotSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsRoute53HostedZoneDnssecExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureNotSigning),
				),
			},
		},
	})
}

func testAccCheckAwsRoute53HostedZoneDnssecDestroy(s *terraform.State) error {
	conn := testAccProviderRoute53KeySigningKey.Meta().(*conns.AWSClient).Route53Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_hosted_zone_dnssec" {
			continue
		}

		hostedZoneDnssec, err := tfroute53.FindHostedZoneDNSSEC(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeDNSSECNotFound) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHostedZone) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Route 53 Hosted Zone DNSSEC (%s): %w", rs.Primary.ID, err)
		}

		if hostedZoneDnssec != nil && hostedZoneDnssec.Status != nil && aws.StringValue(hostedZoneDnssec.Status.ServeSignature) == tfroute53.ServeSignatureSigning {
			return fmt.Errorf("Route 53 Hosted Zone DNSSEC (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsRoute53HostedZoneDnssecExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProviderRoute53KeySigningKey.Meta().(*conns.AWSClient).Route53Conn

		hostedZoneDnssec, err := tfroute53.FindHostedZoneDNSSEC(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading Route 53 Hosted Zone DNSSEC (%s): %w", rs.Primary.ID, err)
		}

		if hostedZoneDnssec == nil {
			return fmt.Errorf("Route 53 Hosted Zone DNSSEC (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsRoute53HostedZoneDnssecConfig_Base(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccRoute53KeySigningKeyRegionProviderConfig(),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  customer_master_key_spec = "ECC_NIST_P256"
  deletion_window_in_days  = 7
  key_usage                = "SIGN_VERIFY"
  policy = jsonencode({
    Statement = [
      {
        Action = [
          "kms:DescribeKey",
          "kms:GetPublicKey",
          "kms:Sign",
        ],
        Effect = "Allow"
        Principal = {
          Service = "api-service.dnssec.route53.aws.internal"
        }
        Sid = "Allow Route 53 DNSSEC Service"
      },
      {
        Action = "kms:*"
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Resource = "*"
        Sid      = "Enable IAM User Permissions"
      },
    ]
    Version = "2012-10-17"
  })
}

resource "aws_route53_zone" "test" {
  name = %[2]q
}

resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
}
`, rName, domainName))
}

func testAccAwsRoute53HostedZoneDnssecConfig(rName, domainName string) string {
	return acctest.ConfigCompose(
		testAccAwsRoute53HostedZoneDnssecConfig_Base(rName, domainName), `
resource "aws_route53_hosted_zone_dnssec" "test" {
  hosted_zone_id = aws_route53_key_signing_key.test.hosted_zone_id
}
`)
}

func testAccAwsRoute53HostedZoneDnssecConfig_SigningStatus(rName, domainName, signingStatus string) string {
	return acctest.ConfigCompose(
		testAccAwsRoute53HostedZoneDnssecConfig_Base(rName, domainName),
		fmt.Sprintf(`
resource "aws_route53_hosted_zone_dnssec" "test" {
  hosted_zone_id = aws_route53_key_signing_key.test.hosted_zone_id
  signing_status = %[1]q
}
`, signingStatus))
}
