package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSWafWebAcl_basic(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.4234791575.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
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

func TestAccAWSWafWebAcl_changeNameForceNew(t *testing.T) {
	var webACL waf.WebACL
	rName1 := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	rName2 := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Required(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.4234791575.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
				),
			},
			{
				Config: testAccAWSWafWebAclConfig_Required(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.4234791575.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", rName2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
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

func TestAccAWSWafWebAcl_DefaultAction(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_DefaultAction(rName, "ALLOW"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.4234791575.type", "ALLOW"),
				),
			},
			{
				Config: testAccAWSWafWebAclConfig_DefaultAction(rName, "BLOCK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.2267395054.type", "BLOCK"),
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

func TestAccAWSWafWebAcl_Rules(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			// Test creating with rule
			{
				Config: testAccAWSWafWebAclConfig_Rules_Single_Rule(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
				),
			},
			// Test adding rule
			{
				Config: testAccAWSWafWebAclConfig_Rules_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
				),
			},
			// Test removing rule
			{
				Config: testAccAWSWafWebAclConfig_Rules_Single_RuleGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
				),
			},
			// Test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSWafWebAcl_disappears(t *testing.T) {
	var webACL waf.WebACL
	rName := fmt.Sprintf("wafacl%s", acctest.RandString(5))
	resourceName := "aws_waf_web_acl.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWafWebAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWafWebAclConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWafWebAclExists(resourceName, &webACL),
					testAccCheckAWSWafWebAclDisappears(&webACL),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWafWebAclDisappears(v *waf.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).wafconn

		wr := newWafRetryer(conn, "global")

		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteWebACLInput{
				ChangeToken: token,
				WebACLId:    v.WebACLId,
			}
			return conn.DeleteWebACL(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF ACL: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSWafWebAclDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_web_acl" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetWebACL(
			&waf.GetWebACLInput{
				WebACLId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.WebACL.WebACLId == rs.Primary.ID {
				return fmt.Errorf("WebACL %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the WebACL is already destroyed
		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			continue
		}

		return err
	}

	return nil
}

func testAccCheckAWSWafWebAclExists(n string, v *waf.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WebACL ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafconn
		resp, err := conn.GetWebACL(&waf.GetWebACLInput{
			WebACLId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.WebACL.WebACLId == rs.Primary.ID {
			*v = *resp.WebACL
			return nil
		}

		return fmt.Errorf("WebACL (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSWafWebAclConfig_Required(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }
}
`, rName, rName)
}

func testAccAWSWafWebAclConfig_DefaultAction(rName, defaultAction string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = %q
  }
}
`, rName, rName, defaultAction)
}

func testAccAWSWafWebAclConfig_Rules_Single_Rule(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %q
  name        = %q

  predicates {
    data_id = "${aws_waf_ipset.test.id}"
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = "${aws_waf_rule.test.id}"

    action {
       type = "BLOCK"
    }
  }
}
`, rName, rName, rName, rName, rName)
}

func testAccAWSWafWebAclConfig_Rules_Single_RuleGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  metric_name = %q
  name        = %q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = "${aws_waf_rule_group.test.id}"
    type     = "GROUP"

    override_action {
       type = "NONE"
    }
  }
}
`, rName, rName, rName, rName)
}

func testAccAWSWafWebAclConfig_Rules_Multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = %q
  name        = %q

  predicates {
    data_id = "${aws_waf_ipset.test.id}"
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_rule_group" "test" {
  metric_name = %q
  name        = %q
}

resource "aws_waf_web_acl" "test" {
  metric_name = %q
  name        = %q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = "${aws_waf_rule.test.id}"

    action {
       type = "BLOCK"
    }
  }

  rules {
    priority = 2
    rule_id  = "${aws_waf_rule_group.test.id}"
    type     = "GROUP"

    override_action {
       type = "NONE"
    }
  }
}
`, rName, rName, rName, rName, rName, rName, rName)
}
