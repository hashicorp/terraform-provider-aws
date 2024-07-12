// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelbv2 "github.com/hashicorp/terraform-provider-aws/internal/service/elbv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBV2ListenerCertificate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	iamServerCertificateResourceName := "aws_iam_server_certificate.test"
	lbListenerResourceName := "aws_lb_listener.test"
	resourceName := "aws_lb_listener_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerCertificateConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, iamServerCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", lbListenerResourceName, names.AttrARN),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17639
func TestAccELBV2ListenerCertificate_CertificateARN_underscores(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	iamServerCertificateResourceName := "aws_iam_server_certificate.test"
	lbListenerResourceName := "aws_lb_listener.test"
	resourceName := "aws_lb_listener_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerCertificateConfig_arnUnderscores(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerCertificateExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCertificateARN, iamServerCertificateResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "listener_arn", lbListenerResourceName, names.AttrARN),
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

func TestAccELBV2ListenerCertificate_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	keys := make([]string, 4)
	certificates := make([]string, 4)
	for i := 0; i < 4; i++ {
		keys[i] = acctest.TLSRSAPrivateKeyPEM(t, 2048)
		certificates[i] = acctest.TLSRSAX509SelfSignedCertificatePEM(t, keys[i], "example.com")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lb_listener_certificate.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerCertificateConfig_multiple(rName, keys, certificates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.default"),
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.additional_1"),
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.additional_2"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", names.AttrCertificateARN),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", names.AttrCertificateARN),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", names.AttrCertificateARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccListenerCertificateConfig_multipleAddNew(rName, keys, certificates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.default"),
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.additional_1"),
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.additional_2"),
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.additional_3"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", names.AttrCertificateARN),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", names.AttrCertificateARN),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", names.AttrCertificateARN),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_3", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_3", names.AttrCertificateARN),
				),
			},
			{
				Config: testAccListenerCertificateConfig_multiple(rName, keys, certificates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.default"),
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.additional_1"),
					testAccCheckListenerCertificateExists(ctx, "aws_lb_listener_certificate.additional_2"),
					testAccCheckListenerCertificateNotExists("aws_lb_listener_certificate.additional_3"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.default", names.AttrCertificateARN),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_1", names.AttrCertificateARN),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", "listener_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener_certificate.additional_2", names.AttrCertificateARN),
				),
			},
		},
	})
}

func TestAccELBV2ListenerCertificate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	resourceName := "aws_lb_listener_certificate.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerCertificateConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerCertificateExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelbv2.ResourceListenerCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBV2ListenerCertificate_disappears_Listener(t *testing.T) {
	ctx := acctest.Context(t)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	resourceName := "aws_lb_listener_certificate.test"
	listenerResourceName := "aws_lb_listener.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerCertificateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerCertificateConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerCertificateExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelbv2.ResourceListener(), listenerResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckListenerCertificateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_listener_certificate" && rs.Type != "aws_alb_listener_certificate" {
				continue
			}

			_, err := tfelbv2.FindListenerCertificateByTwoPartKey(ctx, conn, rs.Primary.Attributes["listener_arn"], rs.Primary.Attributes[names.AttrCertificateARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELBv2 Listener Certificate %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckListenerCertificateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

		_, err := tfelbv2.FindListenerCertificateByTwoPartKey(ctx, conn, rs.Primary.Attributes["listener_arn"], rs.Primary.Attributes[names.AttrCertificateARN])

		return err
	}
}

func testAccCheckListenerCertificateNotExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return nil
		}

		return fmt.Errorf("Not expecting but found: %s", name)
	}
}

func testAccListenerCertificateConfig_base(rName, key, certificate string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-certificate"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-lb-listener-certificate-${count.index}"
  }
}

resource "aws_lb_target_group" "test" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  internal = true
  name     = "%[1]s"
  subnets  = aws_subnet.test[*].id
}

resource "aws_iam_server_certificate" "test" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccListenerCertificateConfig_basic(rName, key, certificate string) string {
	return testAccListenerCertificateConfig_base(rName, key, certificate) + `
resource "aws_lb_listener_certificate" "test" {
  certificate_arn = aws_iam_server_certificate.test.arn
  listener_arn    = aws_lb_listener.test.arn
}
`
}

func testAccListenerCertificateConfig_arnUnderscores(rName, key, certificate string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-certificate"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-lb-listener-certificate-${count.index}"
  }
}

resource "aws_lb_target_group" "test" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id
}

resource "aws_iam_server_certificate" "test" {
  name             = replace("%[1]s", "-", "_")
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
}

resource "aws_lb_listener_certificate" "test" {
  certificate_arn = aws_iam_server_certificate.test.arn
  listener_arn    = aws_lb_listener.test.arn
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key))
}

func testAccListenerCertificateConfig_multiple(rName string, keys, certificates []string) string {
	return testAccListenerCertificateConfig_base(rName, keys[0], certificates[0]) + fmt.Sprintf(`
resource "aws_lb_listener_certificate" "default" {
  listener_arn    = aws_lb_listener.test.arn
  certificate_arn = aws_iam_server_certificate.test.arn
}

resource "aws_lb_listener_certificate" "additional_1" {
  listener_arn    = aws_lb_listener.test.arn
  certificate_arn = aws_iam_server_certificate.additional_1.arn
}

resource "aws_lb_listener_certificate" "additional_2" {
  listener_arn    = aws_lb_listener.test.arn
  certificate_arn = aws_iam_server_certificate.additional_2.arn
}

resource "aws_iam_server_certificate" "additional_1" {
  name             = "%[1]s-additional-1"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_iam_server_certificate" "additional_2" {
  name             = "%[1]s-additional-2"
  certificate_body = "%[4]s"
  private_key      = "%[5]s"
}
`, rName, acctest.TLSPEMEscapeNewlines(certificates[1]), acctest.TLSPEMEscapeNewlines(keys[1]), acctest.TLSPEMEscapeNewlines(certificates[2]), acctest.TLSPEMEscapeNewlines(keys[2]))
}

func testAccListenerCertificateConfig_multipleAddNew(rName string, keys, certificates []string) string {
	return testAccListenerCertificateConfig_multiple(rName, keys, certificates) + fmt.Sprintf(`
resource "aws_iam_server_certificate" "additional_3" {
  name             = "%[1]s-additional-3"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_lb_listener_certificate" "additional_3" {
  listener_arn    = aws_lb_listener.test.arn
  certificate_arn = aws_iam_server_certificate.additional_3.arn
}
`, rName, acctest.TLSPEMEscapeNewlines(certificates[3]), acctest.TLSPEMEscapeNewlines(keys[3]))
}
