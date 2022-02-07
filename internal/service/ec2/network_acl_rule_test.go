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

func TestAccEC2NetworkACLRule_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.test1"),
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.test2"),
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.test3"),
				),
			},
			{
				ResourceName:      "aws_network_acl_rule.test1",
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc("aws_network_acl_rule.test1", "tcp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_disappears(t *testing.T) {
	resourceName := "aws_network_acl_rule.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkACLRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_Disappears_networkACL(t *testing.T) {
	resourceName := "aws_network_acl_rule.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkACL(), "aws_network_acl.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_Disappears_ingressEgressSameNumber(t *testing.T) {
	resourceName := "aws_network_acl_rule.test1"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIngressEgressSameNumberMissingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkACLRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_ipv6(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIPv6Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.baz"),
				),
			},
			{
				ResourceName:      "aws_network_acl_rule.baz",
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc("aws_network_acl_rule.baz", "tcp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_ipv6ICMP(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIPv6ICMPConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resourceName, "58"),
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6710
func TestAccEC2NetworkACLRule_ipv6VPCAssignGeneratedIPv6CIDRBlockUpdate(t *testing.T) {
	var vpc ec2.Vpc
	vpcResourceName := "aws_vpc.test"
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIPv6VPCNotAssignGeneratedIpv6CIDRBlockUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(vpcResourceName, &vpc),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(vpcResourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccNetworkACLRuleIPv6VPCAssignGeneratedIpv6CIDRBlockUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(vpcResourceName, &vpc),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestMatchResourceAttr(vpcResourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
					testAccCheckNetworkACLRuleExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resourceName, "tcp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_allProtocol(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccNetworkACLRuleAllProtocolConfig,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testAccNetworkACLRuleAllProtocolNoRealUpdateConfig,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_tcpProtocol(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccNetworkACLRuleTCPProtocolConfig,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testAccNetworkACLRuleTCPProtocolNoRealUpdateConfig,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckNetworkACLRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_network_acl_rule" {
			continue
		}

		egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])

		if err != nil {
			return err
		}

		naclID := rs.Primary.Attributes["network_acl_id"]

		ruleNumber, err := strconv.Atoi(rs.Primary.Attributes["rule_number"])

		if err != nil {
			return err
		}

		_, err = tfec2.FindNetworkACLEntry(conn, naclID, egress, ruleNumber)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Network ACL Rule %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckNetworkACLRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Network ACL Rule ID is set: %s", n)
		}

		egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])

		if err != nil {
			return err
		}

		naclID := rs.Primary.Attributes["network_acl_id"]

		ruleNumber, err := strconv.Atoi(rs.Primary.Attributes["rule_number"])

		if err != nil {
			return err
		}

		_, err = tfec2.FindNetworkACLEntry(conn, naclID, egress, ruleNumber)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccNetworkACLRuleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test1" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_network_acl_rule" "test2" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 300
  protocol       = "icmp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  icmp_type      = 0
  icmp_code      = -1
}

resource "aws_network_acl_rule" "test3" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 400
  protocol       = "icmp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  icmp_type      = -1
  icmp_code      = -1
}
`, rName)
}

const testAccNetworkACLRuleAllProtocolNoRealUpdateConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-all-proto-no-real-upd"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-no-real-update"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "all"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleTCPProtocolNoRealUpdateConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "testAccNetworkACLRuleTCPProtocolNoRealUpdateConfig"
  }
}
resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id
}
resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleAllProtocolConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-proto"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-all-protocol"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "-1"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleTCPProtocolConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "testAccNetworkACLRuleTCPProtocolConfig"
  }
}
resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id
}
resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "6"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleIPv6Config = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-ipv6"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-ipv6"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id  = aws_network_acl.bar.id
  rule_number     = 150
  egress          = false
  protocol        = "tcp"
  rule_action     = "allow"
  ipv6_cidr_block = "::/0"
  from_port       = 22
  to_port         = 22
}
`

func testAccNetworkACLRuleIngressEgressSameNumberMissingConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_acl_rule" "test1" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 100
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_network_acl_rule" "test2" {
  network_acl_id = aws_network_acl.test.id
  rule_number    = 100
  egress         = true
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`, rName)
}

func testAccNetworkACLRuleIPv6ICMPConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %q
  }
}

resource "aws_network_acl_rule" "test" {
  icmp_code       = -1
  icmp_type       = -1
  ipv6_cidr_block = "::/0"
  network_acl_id  = aws_network_acl.test.id
  protocol        = 58
  rule_action     = "allow"
  rule_number     = 150
}
`, rName, rName)
}

func testAccNetworkACLRuleIPv6VPCAssignGeneratedIpv6CIDRBlockUpdateConfig() string {
	return `
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.3.0.0/16"

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-enabled"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-enabled"
  }
}

resource "aws_network_acl_rule" "test" {
  from_port       = 22
  ipv6_cidr_block = aws_vpc.test.ipv6_cidr_block
  network_acl_id  = aws_network_acl.test.id
  protocol        = "tcp"
  rule_action     = "allow"
  rule_number     = 150
  to_port         = 22
}
`
}

func testAccNetworkACLRuleIPv6VPCNotAssignGeneratedIpv6CIDRBlockUpdateConfig() string {
	return `
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = false
  cidr_block                       = "10.3.0.0/16"

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-not-enabled"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-not-enabled"
  }
}`
}

func testAccNetworkACLRuleImportStateIdFunc(resourceName, resourceProtocol string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		networkAclId := rs.Primary.Attributes["network_acl_id"]
		ruleNumber := rs.Primary.Attributes["rule_number"]
		protocol := rs.Primary.Attributes["protocol"]
		// Ensure the resource's ID will be determined from the original protocol value set in the resource's config
		if protocol != resourceProtocol {
			protocol = resourceProtocol
		}
		egress := rs.Primary.Attributes["egress"]

		return fmt.Sprintf("%s:%s:%s:%s", networkAclId, ruleNumber, protocol, egress), nil
	}
}
