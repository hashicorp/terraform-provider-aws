package elb_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
)

func TestAccELBPolicy_basic(t *testing.T) {
	var policy elb.PolicyDescription
	loadBalancerResourceName := "aws_elb.test-lb"
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					testAccCheckPolicyState(loadBalancerResourceName, resourceName),
				),
			},
		},
	})
}

func TestAccELBPolicy_disappears(t *testing.T) {
	var loadBalancer elb.LoadBalancerDescription
	var policy elb.PolicyDescription
	loadBalancerResourceName := "aws_elb.test-lb"
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(loadBalancerResourceName, &loadBalancer),
					testAccCheckPolicyExists(resourceName, &policy),
					testAccCheckPolicyDisappears(&loadBalancer, &policy),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccPolicyConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(loadBalancerResourceName, &loadBalancer),
					testAccCheckPolicyExists(resourceName, &policy),
					testAccCheckLoadBalancerDisappears(&loadBalancer),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccELBPolicy_LBCookieStickinessPolicyType_computedAttributesOnly(t *testing.T) {
	var policy elb.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyTypeName := "LBCookieStickinessPolicyType"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_policyTypeNameOnly(rName, policyTypeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", policyTypeName),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "1"),
				),
			},
		},
	})
}

func TestAccELBPolicy_SSLNegotiationPolicyType_computedAttributesOnly(t *testing.T) {
	var policy elb.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_policyTypeNameOnly(rName, tfelb.SSLNegotiationPolicyType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", tfelb.SSLNegotiationPolicyType),
					resource.TestMatchResourceAttr(resourceName, "policy_attribute.#", regexp.MustCompile(`[^0]+`)),
				),
			},
		},
	})
}

func TestAccELBPolicy_SSLNegotiationPolicyType_customPolicy(t *testing.T) {
	var policy elb.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_customSSLSecurityPolicy(rName, "Protocol-TLSv1.1", "DHE-RSA-AES256-SHA256"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", tfelb.SSLNegotiationPolicyType),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						"name":  "Protocol-TLSv1.1",
						"value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						"name":  "DHE-RSA-AES256-SHA256",
						"value": "true",
					}),
				),
			},
			{
				Config: testAccPolicyConfig_customSSLSecurityPolicy(rName, "Protocol-TLSv1.2", "ECDHE-ECDSA-AES128-GCM-SHA256"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", rName),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", tfelb.SSLNegotiationPolicyType),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						"name":  "Protocol-TLSv1.2",
						"value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						"name":  "ECDHE-ECDSA-AES128-GCM-SHA256",
						"value": "true",
					}),
				),
			},
		},
	})
}

func TestAccELBPolicy_SSLSecurityPolicy_predefined(t *testing.T) {
	var policy elb.PolicyDescription
	resourceName := "aws_load_balancer_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	predefinedSecurityPolicy := "ELBSecurityPolicy-TLS-1-2-2017-01"
	predefinedSecurityPolicyUpdated := "ELBSecurityPolicy-2016-08"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_predefinedSSLSecurityPolicy(rName, predefinedSecurityPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						"name":  tfelb.ReferenceSecurityPolicy,
						"value": predefinedSecurityPolicy,
					}),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", tfelb.SSLNegotiationPolicyType),
				),
			},
			{
				Config: testAccPolicyConfig_predefinedSSLSecurityPolicy(rName, predefinedSecurityPolicyUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "policy_attribute.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "policy_attribute.*", map[string]string{
						"name":  tfelb.ReferenceSecurityPolicy,
						"value": predefinedSecurityPolicyUpdated,
					}),
					resource.TestCheckResourceAttr(resourceName, "policy_type_name", tfelb.SSLNegotiationPolicyType),
				),
			},
		},
	})
}

func TestAccELBPolicy_updateWhileAssigned(t *testing.T) {
	var policy elb.PolicyDescription
	loadBalancerResourceName := "aws_elb.test-lb"
	resourceName := "aws_load_balancer_policy.test-policy"
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_updateWhileAssigned0(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					testAccCheckPolicyState(loadBalancerResourceName, resourceName),
				),
			},
			{
				Config: testAccPolicyConfig_updateWhileAssigned1(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &policy),
					testAccCheckPolicyState(loadBalancerResourceName, resourceName),
				),
			},
		},
	})
}

func testAccCheckPolicyExists(resourceName string, policyDescription *elb.PolicyDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer Policy ID is set for %s", resourceName)
		}

		loadBalancerName, policyName := tfelb.PolicyParseID(rs.Primary.ID)

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

		input := &elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(loadBalancerName),
			PolicyNames:      []*string{aws.String(policyName)},
		}

		output, err := conn.DescribeLoadBalancerPolicies(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.PolicyDescriptions) == 0 {
			return fmt.Errorf("Load Balancer Policy (%s) not found", rs.Primary.ID)
		}

		*policyDescription = *output.PolicyDescriptions[0]

		return nil
	}
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_load_balancer_policy" {
			continue
		}

		loadBalancerName, policyName := tfelb.PolicyParseID(rs.Primary.ID)
		out, err := conn.DescribeLoadBalancerPolicies(
			&elb.DescribeLoadBalancerPoliciesInput{
				LoadBalancerName: aws.String(loadBalancerName),
				PolicyNames:      []*string{aws.String(policyName)},
			})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && (ec2err.Code() == "PolicyNotFound" || ec2err.Code() == "LoadBalancerNotFound") {
				continue
			}
			return err
		}

		if len(out.PolicyDescriptions) > 0 {
			return fmt.Errorf("Policy still exists")
		}
	}
	return nil
}

func testAccCheckPolicyDisappears(loadBalancer *elb.LoadBalancerDescription, policy *elb.PolicyDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn

		input := elb.DeleteLoadBalancerPolicyInput{
			LoadBalancerName: loadBalancer.LoadBalancerName,
			PolicyName:       policy.PolicyName,
		}
		_, err := conn.DeleteLoadBalancerPolicy(&input)

		return err
	}
}

func testAccCheckPolicyState(elbResource string, policyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[elbResource]
		if !ok {
			return fmt.Errorf("Not found: %s", elbResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[policyResource]
		if !ok {
			return fmt.Errorf("Not found: %s", policyResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ELBConn
		loadBalancerName, policyName := tfelb.PolicyParseID(policy.Primary.ID)
		loadBalancerPolicies, err := conn.DescribeLoadBalancerPolicies(&elb.DescribeLoadBalancerPoliciesInput{
			LoadBalancerName: aws.String(loadBalancerName),
			PolicyNames:      []*string{aws.String(policyName)},
		})

		if err != nil {
			return err
		}

		for _, loadBalancerPolicy := range loadBalancerPolicies.PolicyDescriptions {
			if *loadBalancerPolicy.PolicyName == policyName {
				if *loadBalancerPolicy.PolicyTypeName != policy.Primary.Attributes["policy_type_name"] {
					return fmt.Errorf("PolicyTypeName does not match")
				}
				policyAttributeCount, err := strconv.Atoi(policy.Primary.Attributes["policy_attribute.#"])
				if err != nil {
					return err
				}
				if len(loadBalancerPolicy.PolicyAttributeDescriptions) != policyAttributeCount {
					return fmt.Errorf("PolicyAttributeDescriptions length mismatch")
				}
				policyAttributes := make(map[string]string)
				for k, v := range policy.Primary.Attributes {
					if strings.HasPrefix(k, "policy_attribute.") && strings.HasSuffix(k, ".name") {
						key := v
						value_key := fmt.Sprintf("%s.value", strings.TrimSuffix(k, ".name"))
						policyAttributes[key] = policy.Primary.Attributes[value_key]
					}
				}
				for _, policyAttribute := range loadBalancerPolicy.PolicyAttributeDescriptions {
					if *policyAttribute.AttributeValue != policyAttributes[*policyAttribute.AttributeName] {
						return fmt.Errorf("PollicyAttribute Value mismatch %s != %s: %s", *policyAttribute.AttributeValue, policyAttributes[*policyAttribute.AttributeName], policyAttributes)
					}
				}
			}
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

func testAccPolicyConfig_policyTypeNameOnly(rName, policyType string) string {
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

func testAccPolicyConfig_customSSLSecurityPolicy(rName, protocol, cipher string) string {
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

func testAccPolicyConfig_predefinedSSLSecurityPolicy(rName, securityPolicy string) string {
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
