// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53HostedZoneDNSSEC_basic(t *testing.T) {
	ctx := acctest.Context(t)
	route53ZoneResourceName := "aws_route53_zone.test"
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedZoneDNSSECDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneDNSSECConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccHostedZoneDNSSECExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrHostedZoneID, route53ZoneResourceName, names.AttrID),
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

func TestAccRoute53HostedZoneDNSSEC_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedZoneDNSSECDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneDNSSECConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccHostedZoneDNSSECExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceHostedZoneDNSSEC(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53HostedZoneDNSSEC_signingStatus(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_hosted_zone_dnssec.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedZoneDNSSECDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedZoneDNSSECConfig_signingStatus(rName, domainName, tfroute53.ServeSignatureNotSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccHostedZoneDNSSECExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureNotSigning),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHostedZoneDNSSECConfig_signingStatus(rName, domainName, tfroute53.ServeSignatureSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccHostedZoneDNSSECExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureSigning),
				),
			},
			{
				Config: testAccHostedZoneDNSSECConfig_signingStatus(rName, domainName, tfroute53.ServeSignatureNotSigning),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccHostedZoneDNSSECExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "signing_status", tfroute53.ServeSignatureNotSigning),
				),
			},
		},
	})
}

func testAccCheckHostedZoneDNSSECDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_hosted_zone_dnssec" {
				continue
			}

			_, err := tfroute53.FindHostedZoneDNSSECByZoneID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 Hosted Zone DNSSEC %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccHostedZoneDNSSECExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		_, err := tfroute53.FindHostedZoneDNSSECByZoneID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccHostedZoneDNSSECConfig_base(rName, domainName string) string {
	return fmt.Sprintf(`
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
`, rName, domainName)
}

func testAccHostedZoneDNSSECConfig_basic(rName, domainName string) string {
	return acctest.ConfigCompose(testAccHostedZoneDNSSECConfig_base(rName, domainName), `
resource "aws_route53_hosted_zone_dnssec" "test" {
  hosted_zone_id = aws_route53_key_signing_key.test.hosted_zone_id
}
`)
}

func testAccHostedZoneDNSSECConfig_signingStatus(rName, domainName, signingStatus string) string {
	return acctest.ConfigCompose(testAccHostedZoneDNSSECConfig_base(rName, domainName),
		fmt.Sprintf(`
resource "aws_route53_hosted_zone_dnssec" "test" {
  hosted_zone_id = aws_route53_key_signing_key.test.hosted_zone_id
  signing_status = %[1]q
}
`, signingStatus))
}
