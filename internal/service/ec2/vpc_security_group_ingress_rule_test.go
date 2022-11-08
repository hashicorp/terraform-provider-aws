package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestNormalizeIPProtocol(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plannedValue  attr.Value
		currentValue  attr.Value
		expectedValue attr.Value
		expectError   bool
	}
	tests := map[string]testCase{
		"planned name, current number (equivalent)": {
			plannedValue:  types.String{Value: "icmp"},
			currentValue:  types.String{Value: "1"},
			expectedValue: types.String{Value: "1"},
		},
		"planned number, current name (equivalent)": {
			plannedValue:  types.String{Value: "1"},
			currentValue:  types.String{Value: "icmp"},
			expectedValue: types.String{Value: "icmp"},
		},
		"planned name, current number (not equivalent)": {
			plannedValue:  types.String{Value: "icmp"},
			currentValue:  types.String{Value: "2"},
			expectedValue: types.String{Value: "icmp"},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			request := tfsdk.ModifyAttributePlanRequest{
				AttributePath:  path.Root("test"),
				AttributePlan:  test.plannedValue,
				AttributeState: test.currentValue,
			}
			response := tfsdk.ModifyAttributePlanResponse{
				AttributePlan: request.AttributePlan,
			}
			tfec2.NormalizeIPProtocol().Modify(ctx, request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}

			if diff := cmp.Diff(response.AttributePlan, test.expectedValue); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

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
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
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

func TestAccVPCSecurityGroupIngressRule_cidrIPv4(t *testing.T) {
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
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "0.0.0.0/0"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "53"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "udp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "53"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4Updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v2),
					testAccCheckSecurityGroupIngressRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv4", "10.0.0.0/16"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_cidrIPv6(t *testing.T) {
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
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv6", "2001:db8:85a3::/64"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
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
				Config: testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6Updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v2),
					testAccCheckSecurityGroupIngressRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckResourceAttr(resourceName, "cidr_ipv6", "2001:db8:85a3:2::/64"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckNoResourceAttr(resourceName, "from_port"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "icmpv6"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckNoResourceAttr(resourceName, "to_port"),
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
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
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
					resource.TestCheckNoResourceAttr(resourceName, "referenced_security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_referencedSecurityGroupID(t *testing.T) {
	var v1, v2 ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"
	securityGroup1ResourceName := "aws_security_group.test"
	securityGroup2ResourceName := "aws_security_group.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(resourceName, "referenced_security_group_id", securityGroup1ResourceName, "id"),
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
				Config: testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v2),
					testAccCheckSecurityGroupIngressRuleNotRecreated(&v2, &v1),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(resourceName, "referenced_security_group_id", securityGroup2ResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckNoResourceAttr(resourceName, "tags"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8080"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupIngressRule_ReferencedSecurityGroupID_peerVPC(t *testing.T) {
	var v ec2.SecurityGroupRule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_security_group_ingress_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(t),
		CheckDestroy:             testAccCheckSecurityGroupIngressRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDPeerVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupIngressRuleExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "cidr_ipv6"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ip_protocol", "tcp"),
					resource.TestCheckNoResourceAttr(resourceName, "prefix_list_id"),
					resource.TestMatchResourceAttr(resourceName, "referenced_security_group_id", regexp.MustCompile("^[0-9]{12}/sg-[0-9a-z]{17}$")),
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

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "0.0.0.0/0"
  from_port   = 53
  ip_protocol = "udp"
  to_port     = 53
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv4Updated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv4   = "10.0.0.0/16"
  from_port   = -1
  ip_protocol = "1"
  to_port     = -1
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv6   = "2001:db8:85a3::/64"
  from_port   = 80
  ip_protocol = "tcp"
  to_port     = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_cidrIPv6Updated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  cidr_ipv6   = "2001:db8:85a3:2::/64"
  ip_protocol = "icmpv6"
}
`)
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

func testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupID(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), `
resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  referenced_security_group_id = aws_security_group.test.id
  from_port                    = 80
  ip_protocol                  = "tcp"
  to_port                      = 8080
}
`)
}

func testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDUpdated(rName string) string {
	return acctest.ConfigCompose(testAccVPCSecurityGroupIngressRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_security_group" "test1" {
  vpc_id = aws_vpc.test.id
  name   = "%[1]s-1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  referenced_security_group_id = aws_security_group.test1.id
  from_port                    = 80
  ip_protocol                  = "tcp"
  to_port                      = 8080
}
`, rName))
}

func testAccVPCSecurityGroupIngressRuleConfig_referencedSecurityGroupIDPeerVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), testAccVPCSecurityGroupIngressRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "peer" {
  provider = "awsalternate"

  vpc_id = aws_vpc.peer.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "test" {
  vpc_id        = aws_vpc.test.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = %[2]q
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_security_group_ingress_rule" "test" {
  security_group_id = aws_security_group.test.id

  referenced_security_group_id = "${data.aws_caller_identity.peer.account_id}/${aws_security_group.peer.id}"
  from_port                    = 80
  ip_protocol                  = "tcp"
  to_port                      = 8080

  depends_on = [aws_vpc_peering_connection_accepter.peer]
}
`, rName, acctest.Region()))
}
