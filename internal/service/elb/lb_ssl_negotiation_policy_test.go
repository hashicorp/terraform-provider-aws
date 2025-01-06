// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBSSLNegotiationPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	resourceName := "aws_lb_ssl_negotiation_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLBSSLNegotiationPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLBSSLNegotiationPolicyConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBSSLNegotiationPolicy(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "7"),
				),
			},
		},
	})
}

func TestAccELBSSLNegotiationPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	resourceName := "aws_lb_ssl_negotiation_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLBSSLNegotiationPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLBSSLNegotiationPolicyConfig_update(rName, key, certificate, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBSSLNegotiationPolicy(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "7"),
				),
			},
			{
				Config: testAccLBSSLNegotiationPolicyConfig_update(rName, key, certificate, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBSSLNegotiationPolicy(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "7"),
				),
			},
			{
				Config:   testAccLBSSLNegotiationPolicyConfig_update(rName, key, certificate, 1),
				PlanOnly: true,
			},
		},
	})
}

func TestAccELBSSLNegotiationPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	resourceName := "aws_lb_ssl_negotiation_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLBSSLNegotiationPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLBSSLNegotiationPolicyConfig_basic(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLBSSLNegotiationPolicy(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelb.ResourceSSLNegotiationPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLBSSLNegotiationPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lb_ssl_negotiation_policy" {
				continue
			}

			lbName, lbPort, policyName, err := tfelb.SSLNegotiationPolicyParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfelb.FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELB Classic SSL Negotiation Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLBSSLNegotiationPolicy(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		lbName, lbPort, policyName, err := tfelb.SSLNegotiationPolicyParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		_, err = tfelb.FindLoadBalancerListenerPolicyByThreePartKey(ctx, conn, lbName, lbPort, policyName)

		return err
	}
}

func testAccLBSSLNegotiationPolicyConfig_basic(rName, key, certificate string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  timeouts {
    delete = "30m"
  }
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port      = 8000
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.test.arn
  }
}

resource "aws_lb_ssl_negotiation_policy" "test" {
  name          = %[1]q
  load_balancer = aws_elb.test.id
  lb_port       = 443

  attribute {
    name  = "Protocol-TLSv1"
    value = "false"
  }

  attribute {
    name  = "Protocol-TLSv1.1"
    value = "false"
  }

  attribute {
    name  = "Protocol-TLSv1.2"
    value = "true"
  }

  attribute {
    name  = "Server-Defined-Cipher-Order"
    value = "true"
  }

  attribute {
    name  = "ECDHE-RSA-AES128-GCM-SHA256"
    value = "true"
  }

  attribute {
    name  = "AES128-GCM-SHA256"
    value = "true"
  }

  attribute {
    name  = "EDH-RSA-DES-CBC3-SHA"
    value = "false"
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key)))
}

func testAccLBSSLNegotiationPolicyConfig_update(rName, key, certificate string, certToUse int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_iam_server_certificate" "test" {
  count            = 2
  name_prefix      = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  timeouts {
    delete = "30m"
  }
}

resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port      = 8000
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.test[%[4]d].arn
  }
}

resource "aws_lb_ssl_negotiation_policy" "test" {
  name          = %[1]q
  load_balancer = aws_elb.test.id
  lb_port       = 443

  attribute {
    name  = "Protocol-TLSv1"
    value = "false"
  }

  attribute {
    name  = "Protocol-TLSv1.1"
    value = "false"
  }

  attribute {
    name  = "Protocol-TLSv1.2"
    value = "true"
  }

  attribute {
    name  = "Server-Defined-Cipher-Order"
    value = "true"
  }

  attribute {
    name  = "ECDHE-RSA-AES128-GCM-SHA256"
    value = "true"
  }

  attribute {
    name  = "AES128-GCM-SHA256"
    value = "true"
  }

  attribute {
    name  = "EDH-RSA-DES-CBC3-SHA"
    value = "false"
  }

  triggers = {
    certificate_arn = aws_iam_server_certificate.test[%[4]d].arn,
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), certToUse))
}
