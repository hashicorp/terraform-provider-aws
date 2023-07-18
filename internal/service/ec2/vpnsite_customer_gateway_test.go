// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSiteVPNCustomerGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNCustomerGatewayConfig_basic(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`customer-gateway/cgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(rBgpAsn)),
					resource.TestCheckResourceAttr(resourceName, "certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "device_name", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "172.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ipsec.1"),
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

func TestAccSiteVPNCustomerGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNCustomerGatewayConfig_basic(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceCustomerGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSiteVPNCustomerGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNCustomerGatewayConfig_tags1(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNCustomerGatewayConfig_tags2(rBgpAsn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
			{
				Config: testAccSiteVPNCustomerGatewayConfig_tags1(rBgpAsn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccSiteVPNCustomerGateway_deviceName(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNCustomerGatewayConfig_deviceName(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "device_name", "test"),
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

func TestAccSiteVPNCustomerGateway_4ByteASN(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway ec2.CustomerGateway
	rBgpAsn := strconv.FormatInt(int64(sdkacctest.RandIntRange(64512, 65534))*10000, 10)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSiteVPNCustomerGatewayConfig_siteVPN4ByteASN(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", rBgpAsn),
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

func TestAccSiteVPNCustomerGateway_certificate(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway ec2.CustomerGateway
	var caRoot acmpca.CertificateAuthority
	var caSubordinate acmpca.CertificateAuthority
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_customer_gateway.test"
	acmRootCAResourceName := "aws_acmpca_certificate_authority.root"
	acmSubordinateCAResourceName := "aws_acmpca_certificate_authority.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	rootDomain := acctest.RandomDomainName()
	subDomain := fmt.Sprintf("%s.%s", sdkacctest.RandString(8), rootDomain)
	domain := fmt.Sprintf("%s.%s", sdkacctest.RandString(8), subDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomerGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			// We need to create and activate the CAs before issuing a certificate.
			{
				Config: testAccSiteVPNCustomerGatewayConfig_cas(rootDomain, subDomain),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(ctx, acmRootCAResourceName, &caRoot),
					acctest.CheckACMPCACertificateAuthorityExists(ctx, acmSubordinateCAResourceName, &caSubordinate),
					acctest.CheckACMPCACertificateAuthorityActivateRootCA(ctx, &caRoot),
					acctest.CheckACMPCACertificateAuthorityActivateSubordinateCA(ctx, &caRoot, &caSubordinate),
				),
			},
			{
				Config: testAccSiteVPNCustomerGatewayConfig_certificate(rName, rBgpAsn, rootDomain, subDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(ctx, resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "ip_address", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSiteVPNCustomerGatewayConfig_certificate(rName, rBgpAsn, rootDomain, subDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					// CAs must be DISABLED for deletion.
					acctest.CheckACMPCACertificateAuthorityDisableCA(ctx, &caSubordinate),
					acctest.CheckACMPCACertificateAuthorityDisableCA(ctx, &caRoot),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomerGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_customer_gateway" {
				continue
			}

			_, err := tfec2.FindCustomerGatewayByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Customer Gateway %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomerGatewayExists(ctx context.Context, n string, v *ec2.CustomerGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Customer Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindCustomerGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccSiteVPNCustomerGatewayConfig_basic(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}
`, rBgpAsn)
}

func testAccSiteVPNCustomerGatewayConfig_tags1(rBgpAsn int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"

  tags = {
    %[2]q = %[3]q
  }
}
`, rBgpAsn, tagKey1, tagValue1)
}

func testAccSiteVPNCustomerGatewayConfig_tags2(rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rBgpAsn, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccSiteVPNCustomerGatewayConfig_deviceName(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn     = %[2]d
  ip_address  = "172.0.0.1"
  type        = "ipsec.1"
  device_name = "test"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNCustomerGatewayConfig_siteVPN4ByteASN(rName, rBgpAsn string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]q
  ip_address = "172.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccSiteVPNCustomerGatewayConfig_cas(rootDomain, subDomain string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "root" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

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
  type                            = "SUBORDINATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[2]q
    }
  }
}
`, rootDomain, subDomain)
}

func testAccSiteVPNCustomerGatewayConfig_certificate(rName string, rBgpAsn int, rootDomain, subDomain, domain string) string {
	return acctest.ConfigCompose(testAccSiteVPNCustomerGatewayConfig_cas(rootDomain, subDomain), fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  domain_name               = %[3]q
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_customer_gateway" "test" {
  bgp_asn         = %[2]d
  type            = "ipsec.1"
  certificate_arn = aws_acm_certificate.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn, domain))
}
