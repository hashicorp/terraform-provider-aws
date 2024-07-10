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

func TestAccELBBackendServerPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	privateKey1 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	publicKey1 := acctest.TLSRSAPublicKeyPEM(t, privateKey1)
	privateKey2 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	publicKey2 := acctest.TLSRSAPublicKeyPEM(t, privateKey2)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKey1, "example.com")
	resourceName := "aws_load_balancer_backend_server_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackendServerPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackendServerPolicyConfig_basic(rName, privateKey1, certificate, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendServerPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", "aws_load_balancer_policy.test1", "policy_name"),
				),
			},
		},
	})
}

func TestAccELBBackendServerPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	privateKey1 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	publicKey1 := acctest.TLSRSAPublicKeyPEM(t, privateKey1)
	privateKey2 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	publicKey2 := acctest.TLSRSAPublicKeyPEM(t, privateKey2)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKey1, "example.com")
	resourceName := "aws_load_balancer_backend_server_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackendServerPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackendServerPolicyConfig_basic(rName, privateKey1, certificate, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendServerPolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelb.ResourceBackendServerPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBBackendServerPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	privateKey1 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	publicKey1 := acctest.TLSRSAPublicKeyPEM(t, privateKey1)
	privateKey2 := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	publicKey2 := acctest.TLSRSAPublicKeyPEM(t, privateKey2)
	certificate := acctest.TLSRSAX509SelfSignedCertificatePEM(t, privateKey1, "example.com")
	resourceName := "aws_load_balancer_backend_server_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBackendServerPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBackendServerPolicyConfig_basic(rName, privateKey1, certificate, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendServerPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", "aws_load_balancer_policy.test1", "policy_name"),
				),
			},
			{
				Config: testAccBackendServerPolicyConfig_update(rName, privateKey1, certificate, publicKey1, publicKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBackendServerPolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", "aws_load_balancer_policy.test3", "policy_name"),
				),
			},
		},
	})
}

func testAccCheckBackendServerPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_load_balancer_backend_policy" {
				continue
			}

			lbName, instancePort, err := tfelb.BackendServerPolicyParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfelb.FindLoadBalancerBackendServerPolicyByTwoPartKey(ctx, conn, lbName, instancePort)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELB Classic Backend Server Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBackendServerPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		lbName, instancePort, err := tfelb.BackendServerPolicyParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		_, err = tfelb.FindLoadBalancerBackendServerPolicyByTwoPartKey(ctx, conn, lbName, instancePort)

		return err
	}
}

func testAccBackendServerPolicyConfig_base(rName, privateKey, certificate, publicKey1, publicKey2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port      = 443
    instance_protocol  = "https"
    lb_port            = 443
    lb_protocol        = "https"
    ssl_certificate_id = aws_iam_server_certificate.test.arn
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"

  timeouts {
    delete = "30m"
  }
}

resource "aws_load_balancer_policy" "test0" {
  load_balancer_name = aws_elb.test.name
  policy_name        = "%[1]s-0"
  policy_type_name   = "PublicKeyPolicyType"

  policy_attribute {
    name  = "PublicKey"
    value = "%[4]s"
  }
}

resource "aws_load_balancer_policy" "test1" {
  load_balancer_name = aws_elb.test.name
  policy_name        = "%[1]s-1"
  policy_type_name   = "BackendServerAuthenticationPolicyType"

  policy_attribute {
    name  = "PublicKeyPolicyName"
    value = aws_load_balancer_policy.test0.policy_name
  }
}

resource "aws_load_balancer_policy" "test2" {
  load_balancer_name = aws_elb.test.name
  policy_name        = "%[1]s-2"
  policy_type_name   = "PublicKeyPolicyType"

  policy_attribute {
    name  = "PublicKey"
    value = "%[5]s"
  }
}

resource "aws_load_balancer_policy" "test3" {
  load_balancer_name = aws_elb.test.name
  policy_name        = "%[1]s-3"
  policy_type_name   = "BackendServerAuthenticationPolicyType"

  policy_attribute {
    name  = "PublicKeyPolicyName"
    value = aws_load_balancer_policy.test2.policy_name
  }
}
`,
		rName,
		acctest.TLSPEMEscapeNewlines(certificate),
		acctest.TLSPEMEscapeNewlines(privateKey),
		acctest.TLSPEMRemovePublicKeyEncapsulationBoundaries(acctest.TLSPEMRemoveNewlines(publicKey1)),
		acctest.TLSPEMRemovePublicKeyEncapsulationBoundaries(acctest.TLSPEMRemoveNewlines(publicKey2)),
	))
}

func testAccBackendServerPolicyConfig_basic(rName, privateKey, certificate, publicKey1, publicKey2 string) string {
	return acctest.ConfigCompose(testAccBackendServerPolicyConfig_base(rName, privateKey, certificate, publicKey1, publicKey2), `
resource "aws_load_balancer_backend_server_policy" "test" {
  load_balancer_name = aws_elb.test.name
  instance_port      = 443

  policy_names = [
    aws_load_balancer_policy.test1.policy_name,
  ]
}
`)
}

func testAccBackendServerPolicyConfig_update(rName, privateKey, certificate, publicKey1, publicKey2 string) string {
	return acctest.ConfigCompose(testAccBackendServerPolicyConfig_base(rName, privateKey, certificate, publicKey1, publicKey2), `
resource "aws_load_balancer_backend_server_policy" "test" {
  load_balancer_name = aws_elb.test.name
  instance_port      = 443

  policy_names = [
    aws_load_balancer_policy.test3.policy_name,
  ]
}
`)
}
