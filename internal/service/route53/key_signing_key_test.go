// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53KeySigningKey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	kmsKeyResourceName := "aws_kms_key.test"
	route53ZoneResourceName := "aws_route53_zone.test"
	resourceName := "aws_route53_key_signing_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeySigningKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeySigningKeyConfig_name(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccKeySigningKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "digest_algorithm_mnemonic", "SHA-256"),
					resource.TestCheckResourceAttr(resourceName, "digest_algorithm_type", acctest.Ct2),
					resource.TestMatchResourceAttr(resourceName, "digest_value", regexache.MustCompile(`^[0-9A-F]+$`)),
					resource.TestMatchResourceAttr(resourceName, "dnskey_record", regexache.MustCompile(`^257 [0-9]+ [0-9]+ [0-9A-Za-z+/]+={0,3}$`)),
					resource.TestMatchResourceAttr(resourceName, "ds_record", regexache.MustCompile(`^[0-9]+ [0-9]+ [0-9]+ [0-9A-F]+$`)),
					resource.TestCheckResourceAttr(resourceName, "flag", "257"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrHostedZoneID, route53ZoneResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "key_management_service_arn", kmsKeyResourceName, names.AttrARN),
					resource.TestMatchResourceAttr(resourceName, "key_tag", regexache.MustCompile(`^[0-9]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, names.AttrPublicKey, regexache.MustCompile(`^[0-9A-Za-z+/]+={0,3}$`)),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm_mnemonic", "ECDSAP256SHA256"),
					resource.TestCheckResourceAttr(resourceName, "signing_algorithm_type", "13"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, tfroute53.KeySigningKeyStatusActive),
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

func TestAccRoute53KeySigningKey_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_key_signing_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeySigningKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeySigningKeyConfig_name(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccKeySigningKeyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceKeySigningKey(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53KeySigningKey_status(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_key_signing_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeySigningKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKeySigningKeyConfig_status(rName, domainName, tfroute53.KeySigningKeyStatusInactive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccKeySigningKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, tfroute53.KeySigningKeyStatusInactive),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeySigningKeyConfig_status(rName, domainName, tfroute53.KeySigningKeyStatusActive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccKeySigningKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, tfroute53.KeySigningKeyStatusActive),
				),
			},
			{
				Config: testAccKeySigningKeyConfig_status(rName, domainName, tfroute53.KeySigningKeyStatusInactive),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccKeySigningKeyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, tfroute53.KeySigningKeyStatusInactive),
				),
			},
		},
	})
}

func testAccCheckKeySigningKeyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_key_signing_key" {
				continue
			}

			_, err := tfroute53.FindKeySigningKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrHostedZoneID], rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 Key Signing Key %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccKeySigningKeyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		_, err := tfroute53.FindKeySigningKeyByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrHostedZoneID], rs.Primary.Attributes[names.AttrName])

		return err
	}
}

func testAccKeySigningKeyConfig_base(rName, domainName string) string {
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
`, rName, domainName)
}

func testAccKeySigningKeyConfig_name(rName, domainName string) string {
	return acctest.ConfigCompose(testAccKeySigningKeyConfig_base(rName, domainName), fmt.Sprintf(`
resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
}
`, rName))
}

func testAccKeySigningKeyConfig_status(rName, domainName, status string) string {
	return acctest.ConfigCompose(testAccKeySigningKeyConfig_base(rName, domainName), fmt.Sprintf(`
resource "aws_route53_key_signing_key" "test" {
  hosted_zone_id             = aws_route53_zone.test.id
  key_management_service_arn = aws_kms_key.test.arn
  name                       = %[1]q
  status                     = %[2]q
}
`, rName, status))
}
