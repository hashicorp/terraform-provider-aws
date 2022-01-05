package ec2_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2CustomerGateway_basic(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`customer-gateway/cgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(rBgpAsn)),
					resource.TestCheckResourceAttr(resourceName, "certificate_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "device_name", ""),
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

func TestAccEC2CustomerGateway_disappears(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceCustomerGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2CustomerGateway_tags(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfigTags1(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomerGatewayConfigTags2(rBgpAsn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
			{
				Config: testAccCustomerGatewayConfigTags1(rBgpAsn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccEC2CustomerGateway_deviceName(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfigDeviceName(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
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

func TestAccEC2CustomerGateway_4ByteASN(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := strconv.FormatInt(int64(sdkacctest.RandIntRange(64512, 65534))*10000, 10)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfig4ByteAsn(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
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

func TestAccEC2CustomerGateway_certificate(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_customer_gateway.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfigCertificate(rName, rBgpAsn, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
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

func testAccCheckCustomerGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_customer_gateway" {
			continue
		}

		_, err := tfec2.FindCustomerGatewayByID(conn, rs.Primary.ID)

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

func testAccCheckCustomerGatewayExists(n string, v *ec2.CustomerGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Customer Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindCustomerGatewayByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCustomerGatewayConfig(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}
`, rBgpAsn)
}

func testAccCustomerGatewayConfigTags1(rBgpAsn int, tagKey1, tagValue1 string) string {
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

func testAccCustomerGatewayConfigTags2(rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccCustomerGatewayConfigDeviceName(rName string, rBgpAsn int) string {
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

func testAccCustomerGatewayConfig4ByteAsn(rName, rBgpAsn string) string {
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

func testAccCustomerGatewayConfigCertificate(rName string, rBgpAsn int, domain string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

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

resource "aws_acmpca_certificate" "root" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = aws_acmpca_certificate_authority.root.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 2
  }
}

resource "aws_acmpca_certificate_authority_certificate" "root" {
  certificate_authority_arn = aws_acmpca_certificate_authority.root.arn

  certificate       = aws_acmpca_certificate.root.certificate
  certificate_chain = aws_acmpca_certificate.root.certificate_chain
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "SUBORDINATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "sub.%[1]s"
    }
  }
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.root.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/SubordinateCACertificate_PathLen0/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

resource "aws_acm_certificate" "test" {
  domain_name               = "test.sub.%[1]s"
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_customer_gateway" "test" {
  bgp_asn         = %[2]d
  ip_address      = "172.0.0.1"
  type            = "ipsec.1"
  certificate_arn = aws_acm_certificate.test.arn

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}
