// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2EIPDomainName_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_eip_domain_name.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDomainNameConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPDomainNameExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "ptr_record"),
				),
			},
		},
	})
}

func TestAccEC2EIPDomainName_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_eip_domain_name.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDomainNameConfig_basic(rName, rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPDomainNameExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfec2.ResourceEIPDomainName, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EIPDomainName_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_eip_domain_name.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rootDomain := acctest.ACMCertificateDomainFromEnv(t)
	domain1 := acctest.ACMCertificateRandomSubDomain(rootDomain)
	domain2 := acctest.ACMCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEIPDomainNameDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEIPDomainNameConfig_original(rName, rootDomain, domain1, domain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPDomainNameExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "ptr_record"),
				),
			},
			{
				Config: testAccEIPDomainNameConfig_updated(rName, rootDomain, domain1, domain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPDomainNameExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "ptr_record"),
				),
			},
		},
	})
}

func testAccCheckEIPDomainNameExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindEIPDomainNameAttributeByAllocationID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckEIPDomainNameDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eip_domain_name" {
				continue
			}

			_, err := tfec2.FindEIPDomainNameAttributeByAllocationID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 EIP Domain Name %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccEIPDomainNameConfig_basic(rName, rootDomain, domain string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

resource "aws_route53_record" "test" {
  zone_id = data.aws_route53_zone.test.zone_id
  ttl     = 60
  type    = "A"
  name    = %[3]q

  records = [aws_eip.test.public_ip]
}

resource "aws_eip_domain_name" "test" {
  allocation_id = aws_eip.test.allocation_id
  domain_name   = aws_route53_record.test.fqdn
}
`, rName, rootDomain, domain)
}

func testAccEIPDomainNameConfig_original(rName, rootDomain, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

resource "aws_route53_record" "test1" {
  zone_id = data.aws_route53_zone.test.zone_id
  ttl     = 60
  type    = "A"
  name    = %[3]q

  records = [aws_eip.test.public_ip]
}

resource "aws_route53_record" "test2" {
  zone_id = data.aws_route53_zone.test.zone_id
  ttl     = 60
  type    = "A"
  name    = %[4]q

  records = [aws_eip.test.public_ip]
}

resource "aws_eip_domain_name" "test" {
  allocation_id = aws_eip.test.allocation_id
  domain_name   = aws_route53_record.test1.fqdn
}
`, rName, rootDomain, domain1, domain2)
}

func testAccEIPDomainNameConfig_updated(rName, rootDomain, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

data "aws_route53_zone" "test" {
  name         = %[2]q
  private_zone = false
}

resource "aws_route53_record" "test1" {
  zone_id = data.aws_route53_zone.test.zone_id
  ttl     = 60
  type    = "A"
  name    = %[3]q

  records = [aws_eip.test.public_ip]
}

resource "aws_route53_record" "test2" {
  zone_id = data.aws_route53_zone.test.zone_id
  ttl     = 60
  type    = "A"
  name    = %[4]q

  records = [aws_eip.test.public_ip]
}

resource "aws_eip_domain_name" "test" {
  allocation_id = aws_eip.test.allocation_id
  domain_name   = aws_route53_record.test2.fqdn
}
`, rName, rootDomain, domain1, domain2)
}
