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

func TestAccELBListenerPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_load_balancer_listener_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", "aws_load_balancer_policy.test", "policy_name"),
				),
			},
		},
	})
}

func TestAccELBListenerPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	key := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, key, "example.com")
	resourceName := "aws_load_balancer_listener_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerPolicyConfig_update(rName, key, certificate, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", "aws_load_balancer_policy.test", "policy_name"),
				),
			},
			{
				Config: testAccListenerPolicyConfig_update(rName, key, certificate, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", "aws_load_balancer_policy.test", "policy_name"),
				),
			},
			{
				Config:   testAccListenerPolicyConfig_update(rName, key, certificate, 1),
				PlanOnly: true,
			},
		},
	})
}

func TestAccELBListenerPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_load_balancer_listener_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckListenerPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccListenerPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckListenerPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelb.ResourceListenerPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckListenerPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_load_balancer_listener_policy" {
				continue
			}

			lbName, lbPort, err := tfelb.ListenerPolicyParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tfelb.FindLoadBalancerListenerPolicyByTwoPartKey(ctx, conn, lbName, lbPort)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELB Classic Listener Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckListenerPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ELB Classic Listener Policy ID is set")
		}

		lbName, lbPort, err := tfelb.ListenerPolicyParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		_, err = tfelb.FindLoadBalancerListenerPolicyByTwoPartKey(ctx, conn, lbName, lbPort)

		return err
	}
}

func testAccListenerPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_load_balancer_policy" "test" {
  load_balancer_name = aws_elb.test.name
  policy_name        = %[1]q
  policy_type_name   = "AppCookieStickinessPolicyType"

  policy_attribute {
    name  = "CookieName"
    value = "wafer"
  }
}

resource "aws_load_balancer_listener_policy" "test" {
  load_balancer_name = aws_elb.test.name
  load_balancer_port = 80

  policy_names = [
    aws_load_balancer_policy.test.policy_name,
  ]
}
`, rName))
}

func testAccListenerPolicyConfig_update(rName, key, certificate string, certToUse int) string {
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
    instance_port      = 443
    instance_protocol  = "http"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.test[%[4]d].arn
  }
}

resource "aws_load_balancer_policy" "test" {
  load_balancer_name = aws_elb.test.name
  policy_name        = %[1]q
  policy_type_name   = "SSLNegotiationPolicyType"

  policy_attribute {
    name  = "Reference-Security-Policy"
    value = "ELBSecurityPolicy-TLS-1-2-2017-01"
  }
}

resource "aws_load_balancer_listener_policy" "test" {
  load_balancer_name = aws_elb.test.name
  load_balancer_port = 443

  policy_names = [
    aws_load_balancer_policy.test.policy_name,
  ]

  triggers = {
    certificate_arn = aws_iam_server_certificate.test[%[4]d].arn,
  }
}
`, rName, acctest.TLSPEMEscapeNewlines(certificate), acctest.TLSPEMEscapeNewlines(key), certToUse))
}
