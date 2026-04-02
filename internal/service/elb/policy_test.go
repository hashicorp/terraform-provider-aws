// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := acctest.RandInt(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
				),
			},
		},
	})
}

func TestAccELBPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := acctest.RandInt(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelb.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBPolicy_LBCookieStickinessPolicyType_computedAttributesOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policyTypeName := "LBCookieStickinessPolicyType"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_typeNameOnly(rName, policyTypeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", policyTypeName),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "1"),
				),
			},
		},
	})
}

func TestAccELBPolicy_SSLNegotiationPolicyType_computedAttributesOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_typeNameOnly(rName, "SSLNegotiationPolicyType"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_attribute"), tfknownvalue.ListNotEmpty()),
				},
			},
		},
	})
}

func TestAccELBPolicy_SSLNegotiationPolicyType_customPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_customSSLSecurity(rName, "Protocol-TLSv1.1", "DHE-RSA-AES256-SHA256"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						names.AttrName:  "Protocol-TLSv1.1",
						names.AttrValue: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						names.AttrName:  "DHE-RSA-AES256-SHA256",
						names.AttrValue: acctest.CtTrue,
					}),
				),
			},
			{
				Config: testAccPolicyConfig_customSSLSecurity(rName, "Protocol-TLSv1.2", "ECDHE-ECDSA-AES128-GCM-SHA256"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						names.AttrName:  "Protocol-TLSv1.2",
						names.AttrValue: acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						names.AttrName:  "ECDHE-ECDSA-AES128-GCM-SHA256",
						names.AttrValue: acctest.CtTrue,
					}),
				),
			},
		},
	})
}

func TestAccELBPolicy_SSLSecurityPolicy_predefined(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	predefinedSecurityPolicy := "ELBSecurityPolicy-TLS-1-2-2017-01"
	predefinedSecurityPolicyUpdated := "ELBSecurityPolicy-2016-08"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predefinedSSLSecurity(rName, predefinedSecurityPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						names.AttrName:  "Reference-Security-Policy",
						names.AttrValue: predefinedSecurityPolicy,
					}),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
				),
			},
			{
				Config: testAccPolicyConfig_predefinedSSLSecurity(rName, predefinedSecurityPolicyUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						names.AttrName:  "Reference-Security-Policy",
						names.AttrValue: predefinedSecurityPolicyUpdated,
					}),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
				),
			},
		},
	})
}

func TestAccELBPolicy_updateWhileAssigned(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := acctest.RandInt(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_updateWhileAssigned0(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
				),
			},
			{
				Config: testAccPolicyConfig_updateWhileAssigned1(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
				),
			},
		},
	})
}

func testAccCheckPolicyExists(ctx context.Context, t *testing.T, n string, v *awstypes.PolicyDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		lbName, policyName, err := tfelb.PolicyParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.ProviderMeta(ctx, t).ELBClient(ctx)

		output, err := tfelb.FindLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

		if err != nil {
			return err
		}

		*output = *v

		return nil
	}
}

func testAccCheckPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_load_balancer_policy" {
				continue
			}

			lbName, policyName, err := tfelb.PolicyParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfelb.FindLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ELB Classic Load Balancer Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPolicyConfig_basic(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test-lb" {
  name               = "test-lb-%[1]d"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test-policy" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-policy-%[1]d"
  policy_type_name   = "AppCookieStickinessPolicyType"

  policy_attribute {
    name  = "CookieName"
    value = "magic_cookie"
  }
}
`, rInt))
}

func testAccPolicyConfig_typeNameOnly(rName, policyType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elb" "test" {
  name               = %[1]q
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test" {
  load_balancer_name = aws_elb.test.name
  policy_name        = %[1]q
  policy_type_name   = %[2]q
}
`, rName, policyType))
}

func testAccPolicyConfig_customSSLSecurity(rName, protocol, cipher string) string {
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

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test" {
  load_balancer_name = aws_elb.test.name
  policy_name        = %[1]q
  policy_type_name   = "SSLNegotiationPolicyType"

  policy_attribute {
    name  = %[2]q
    value = "true"
  }

  policy_attribute {
    name  = %[3]q
    value = "true"
  }
}
`, rName, protocol, cipher))
}

func testAccPolicyConfig_predefinedSSLSecurity(rName, securityPolicy string) string {
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

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test" {
  load_balancer_name = aws_elb.test.name
  policy_name        = %[1]q
  policy_type_name   = "SSLNegotiationPolicyType"

  policy_attribute {
    name  = "Reference-Security-Policy"
    value = %[2]q
  }
}
`, rName, securityPolicy))
}

func testAccPolicyConfig_updateWhileAssigned0(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test-lb" {
  name               = "test-lb-%[1]d"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test-policy" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-policy-%[1]d"
  policy_type_name   = "AppCookieStickinessPolicyType"

  policy_attribute {
    name  = "CookieName"
    value = "magic_cookie"
  }
}

resource "aws_load_balancer_listener_policy" "test-lb-test-policy-80" {
  load_balancer_name = aws_elb.test-lb.name
  load_balancer_port = 80

  policy_names = [
    aws_load_balancer_policy.test-policy.policy_name,
  ]
}
`, rInt))
}

func testAccPolicyConfig_updateWhileAssigned1(rInt int) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elb" "test-lb" {
  name               = "test-lb-%[1]d"
  availability_zones = [data.aws_availability_zones.available.names[0]]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_load_balancer_policy" "test-policy" {
  load_balancer_name = aws_elb.test-lb.name
  policy_name        = "test-policy-%[1]d"
  policy_type_name   = "AppCookieStickinessPolicyType"

  policy_attribute {
    name  = "CookieName"
    value = "unicorn_cookie"
  }
}

resource "aws_load_balancer_listener_policy" "test-lb-test-policy-80" {
  load_balancer_name = aws_elb.test-lb.name
  load_balancer_port = 80

  policy_names = [
    aws_load_balancer_policy.test-policy.policy_name,
  ]
}
`, rInt))
}
