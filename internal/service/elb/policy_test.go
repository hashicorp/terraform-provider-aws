// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccELBPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
				),
			},
		},
	})
}

func TestAccELBPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelb.ResourcePolicy(), resourceName),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyTypeName := "LBCookieStickinessPolicyType"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_typeNameOnly(rName, policyTypeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", policyTypeName),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccELBPolicy_SSLNegotiationPolicyType_computedAttributesOnly(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_typeNameOnly(rName, "SSLNegotiationPolicyType"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
					resource.TestMatchResourceAttr(resourceName, "policy_attribute.#", regexache.MustCompile(`[^0]+`)),
				),
			},
		},
	})
}

func TestAccELBPolicy_SSLNegotiationPolicyType_customPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var policy awstypes.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_customSSLSecurity(rName, "Protocol-TLSv1.1", "DHE-RSA-AES256-SHA256"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", acctest.Ct2),
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
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", "SSLNegotiationPolicyType"),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", acctest.Ct2),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	predefinedSecurityPolicy := "ELBSecurityPolicy-TLS-1-2-2017-01"
	predefinedSecurityPolicyUpdated := "ELBSecurityPolicy-2016-08"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predefinedSSLSecurity(rName, predefinedSecurityPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", acctest.Ct1),
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
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", acctest.Ct1),
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
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ELBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_updateWhileAssigned0(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
				),
			},
			{
				Config: testAccPolicyConfig_updateWhileAssigned1(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
				),
			},
		},
	})
}

func testAccCheckPolicyExists(ctx context.Context, n string, v *awstypes.PolicyDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		lbName, policyName, err := tfelb.PolicyParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		output, err := tfelb.FindLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

		if err != nil {
			return err
		}

		*output = *v

		return nil
	}
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_load_balancer_policy" {
				continue
			}

			lbName, policyName, err := tfelb.PolicyParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tfelb.FindLoadBalancerPolicyByTwoPartKey(ctx, conn, lbName, policyName)

			if tfresource.NotFound(err) {
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
