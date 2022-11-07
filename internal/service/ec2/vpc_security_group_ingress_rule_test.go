package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccVPCSecurityGroupIngressRule_basic(t *testing.T) {
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
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

func TestAccVPCSecurityGroupIngressRule_disappears(t *testing.T) {
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(acctest.Provider, tfec2.ResourceSecurityGroupIngressRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_tags(t *testing.T) {
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_description(t *testing.T) {
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v2),
					testAccCheckSecurityGroupIngressRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_prefixListID(t *testing.T) {
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"
	vpcEndpoint1ResourceName := "aws_vpc_endpoint.test1"
	vpcEndpoint2ResourceName := "aws_vpc_endpoint.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_prefixListID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", vpcEndpoint1ResourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_prefixListIDUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v2),
					testAccCheckSecurityGroupIngressRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_id", vpcEndpoint2ResourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
		},
	})
}

func testAccCheckSecurityGroupIngressRuleNotRecreated(i, j *ec2.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.SecurityGroupRuleId) != aws.StringValue(j.SecurityGroupRuleId) {
			return errors.New("EC2 Security Group Rule was recreated")
		}

		return nil
	}
}

func testAccCheckSecurityGroupIngressRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_security_group_ingress_rule" {
			continue
		}

		_, err := tfec2.FindSecurityGroupIngressRuleByID(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Security Group Ingress Rule still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSecurityGroupIngressRuleExists(n string, v *ec2.SecurityGroupRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Security Group Ingress Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindSecurityGroupIngressRuleByID(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccVPCSecurityGroupIngressRuleConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCSecurityGroupIngressRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccVPCSecurityGroupIngressRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccVPCSecurityGroupIngressRuleConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/8"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080

  description = %[1]q
}
`, description))
}

func testAccVPCSecurityGroupIngressRuleConfig_prefixListIDBase(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test1" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test2" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.dynamodb"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCSecurityGroupIngressRuleConfig_prefixListID(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_prefixListIDBase(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  prefix_list_id = aws_vpc_endpoint.test1.prefix_list_id
  from_port      = 80
  ip_protocol    = "tcp"
  to_port        = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_prefixListIDUpdated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_prefixListIDBase(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  prefix_list_id = aws_vpc_endpoint.test2.prefix_list_id
  from_port      = 80
  ip_protocol    = "tcp"
  to_port        = 8080
}
`)
}
