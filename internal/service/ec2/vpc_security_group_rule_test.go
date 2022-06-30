package ec2_test

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestIPPermissionIDHash(t *testing.T) {
	simple := &ec2.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(80),
		ToPort:     aws.Int64(8000),
		IpRanges: []*ec2.IpRange{
			{
				CidrIp: aws.String("10.0.0.0/8"),
			},
		},
	}

	egress := &ec2.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(80),
		ToPort:     aws.Int64(8000),
		IpRanges: []*ec2.IpRange{
			{
				CidrIp: aws.String("10.0.0.0/8"),
			},
		},
	}

	egress_all := &ec2.IpPermission{
		IpProtocol: aws.String("-1"),
		IpRanges: []*ec2.IpRange{
			{
				CidrIp: aws.String("10.0.0.0/8"),
			},
		},
	}

	vpc_security_group_source := &ec2.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(80),
		ToPort:     aws.Int64(8000),
		UserIdGroupPairs: []*ec2.UserIdGroupPair{
			{
				UserId:  aws.String("987654321"),
				GroupId: aws.String("sg-12345678"),
			},
			{
				UserId:  aws.String("123456789"),
				GroupId: aws.String("sg-987654321"),
			},
			{
				UserId:  aws.String("123456789"),
				GroupId: aws.String("sg-12345678"),
			},
		},
	}

	security_group_source := &ec2.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(80),
		ToPort:     aws.Int64(8000),
		UserIdGroupPairs: []*ec2.UserIdGroupPair{
			{
				UserId:    aws.String("987654321"),
				GroupName: aws.String("my-security-group"),
			},
			{
				UserId:    aws.String("123456789"),
				GroupName: aws.String("my-security-group"),
			},
			{
				UserId:    aws.String("123456789"),
				GroupName: aws.String("my-other-security-group"),
			},
		},
	}

	// hardcoded hashes, to detect future change
	cases := []struct {
		Input  *ec2.IpPermission
		Type   string
		Output string
	}{
		{simple, "ingress", "sgrule-3403497314"},
		{egress, "egress", "sgrule-1173186295"},
		{egress_all, "egress", "sgrule-766323498"},
		{vpc_security_group_source, "egress", "sgrule-351225364"},
		{security_group_source, "egress", "sgrule-2198807188"},
	}

	for _, tc := range cases {
		actual := tfec2.IPPermissionIDHash("sg-12345", tc.Type, tc.Input)
		if actual != tc.Output {
			t.Errorf("input: %s - %s\noutput: %s", tc.Type, tc.Input, actual)
		}
	}
}

func TestAccVPCSecurityGroupRule_Ingress_vpc(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_IngressSourceWithAccount_id(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressSourceAccountID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "some description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestMatchResourceAttr(resourceName, "source_security_group_id", regexp.MustCompile("^[0-9]{12}/sg-[0-9a-z]{17}$")),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_protocol(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_icmpv6(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressIcmpv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "icmpv6"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_ipv6(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_classic(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressClassic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupEC2ClassicExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_egress(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_egress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "egress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_selfReference(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_selfReference(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_expectInvalidTypeError(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCSecurityGroupRuleConfig_expectInvalidType(rName),
				ExpectError: regexp.MustCompile(`expected type to be one of \[ingress egress\]`),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_expectInvalidCIDR(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCSecurityGroupRuleConfig_invalidIPv4CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccVPCSecurityGroupRuleConfig_invalidIPv6CIDR(rName),
				ExpectError: regexp.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

// testing partial match implementation
func TestAccVPCSecurityGroupRule_PartialMatching_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_security_group_rule.test1"
	resource2Name := "aws_security_group_rule.test2"
	resource3Name := "aws_security_group_rule.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_partialMatching(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.#", "3"),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.0", "10.0.2.0/24"),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.1", "10.0.3.0/24"),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.2", "10.0.4.0/24"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.0", "10.0.5.0/24"),
					resource.TestCheckResourceAttr(resource3Name, "cidr_blocks.#", "3"),
					resource.TestCheckResourceAttr(resource3Name, "cidr_blocks.0", "10.0.2.0/24"),
					resource.TestCheckResourceAttr(resource3Name, "cidr_blocks.1", "10.0.3.0/24"),
					resource.TestCheckResourceAttr(resource3Name, "cidr_blocks.2", "10.0.4.0/24"),
				),
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resource1Name),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resource2Name),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource3Name,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resource3Name),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_PartialMatching_source(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_security_group_rule.test1"
	resource2Name := "aws_security_group_rule.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_partialMatchingSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttrSet(resource1Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.#", "3"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.0", "10.0.2.0/24"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.1", "10.0.3.0/24"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.2", "10.0.4.0/24"),
					resource.TestCheckNoResourceAttr(resource2Name, "source_security_group_id"),
				),
			},
			{
				ResourceName:      resource1Name,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resource1Name),
				ImportStateVerify: true,
			},
			{
				ResourceName:      resource2Name,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resource2Name),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_issue5310(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_issue5310(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_race(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sgResourceName := "aws_security_group.test"
	n := 50

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_race(rName, n),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					testAccCheckSecurityGroupRuleCount(&group, n, n),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_selfSource(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_selfInSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "source_security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_prefixListEgress(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"
	vpceResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_prefixListEgress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "0"),
					resource.TestCheckNoResourceAttr(resourceName, "description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_ids.0", vpceResourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "egress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_ingressDescription(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "TF acceptance test ingress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_egressDescription(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_egressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "TF acceptance test egress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "egress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_IngressDescription_updates(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "TF acceptance test ingress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressUpdateDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "TF acceptance test ingress rule updated"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_EgressDescription_updates(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_egressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "TF acceptance test egress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "egress"),
				),
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_egressUpdateDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, "description", "TF acceptance test egress rule updated"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, "type", "egress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Description_allPorts(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_descriptionAllPorts(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_descriptionAllPorts(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_DescriptionAllPorts_nonZeroPorts(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_descriptionAllPortsNonZeroPorts(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_descriptionAllPortsNonZeroPorts(rName, "description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "self", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6416
func TestAccVPCSecurityGroupRule_MultipleRuleSearching_allProtocolCrash(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	securityGroupResourceName := "aws_security_group.test"
	resourceName1 := "aws_security_group_rule.test1"
	resourceName2 := "aws_security_group_rule.test2"

	rule1 := ec2.IpPermission{
		IpProtocol: aws.String("-1"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("10.0.0.0/8")},
		},
	}

	rule2 := ec2.IpPermission{
		FromPort:   aws.Int64(443),
		ToPort:     aws.Int64(443),
		IpProtocol: aws.String("tcp"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("172.168.0.0/16")},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_multipleSearchingAllProtocolCrash(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(securityGroupResourceName, &group),
					testAccCheckSecurityGroupRuleAttributes(resourceName1, &group, &rule1, "ingress"),
					testAccCheckSecurityGroupRuleAttributes(resourceName2, &group, &rule2, "ingress"),
					resource.TestCheckResourceAttr(resourceName1, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName1, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName1, "to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName2, "from_port", "443"),
					resource.TestCheckResourceAttr(resourceName2, "protocol", "tcp"),
					resource.TestCheckResourceAttr(resourceName2, "to_port", "443"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_multiDescription(t *testing.T) {
	var group ec2.SecurityGroup
	var nat ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	rule1 := ec2.IpPermission{
		FromPort:   aws.Int64(22),
		ToPort:     aws.Int64(22),
		IpProtocol: aws.String("tcp"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("0.0.0.0/0"), Description: aws.String("CIDR Description")},
		},
	}

	rule2 := ec2.IpPermission{
		FromPort:   aws.Int64(22),
		ToPort:     aws.Int64(22),
		IpProtocol: aws.String("tcp"),
		Ipv6Ranges: []*ec2.Ipv6Range{
			{CidrIpv6: aws.String("::/0"), Description: aws.String("IPv6 CIDR Description")},
		},
	}

	var rule3 ec2.IpPermission

	// This function creates the expected IPPermission with the group id from an
	// external security group, needed because Security Group IDs are generated on
	// AWS side and can't be known ahead of time.
	setupSG := func(*terraform.State) error {
		if nat.GroupId == nil {
			return fmt.Errorf("Error: nat group has nil GroupID")
		}

		rule3 = ec2.IpPermission{
			FromPort:   aws.Int64(22),
			ToPort:     aws.Int64(22),
			IpProtocol: aws.String("tcp"),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{GroupId: nat.GroupId, Description: aws.String("NAT SG Description")},
			},
		}

		return nil
	}

	var endpoint ec2.VpcEndpoint
	var rule4 ec2.IpPermission

	// This function creates the expected IPPermission with the prefix list ID from
	// the VPC Endpoint created in the test
	setupPL := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		prefixListInput := &ec2.DescribePrefixListsInput{
			Filters: []*ec2.Filter{
				{Name: aws.String("prefix-list-name"), Values: []*string{endpoint.ServiceName}},
			},
		}

		log.Printf("[DEBUG] Reading VPC Endpoint prefix list: %s", prefixListInput)
		prefixListsOutput, err := conn.DescribePrefixLists(prefixListInput)

		if err != nil {
			return fmt.Errorf("error reading VPC Endpoint prefix list: %w", err)
		}

		if len(prefixListsOutput.PrefixLists) != 1 {
			return fmt.Errorf("unexpected multiple prefix lists associated with the service: %s", prefixListsOutput)
		}

		rule4 = ec2.IpPermission{
			FromPort:   aws.Int64(22),
			ToPort:     aws.Int64(22),
			IpProtocol: aws.String("tcp"),
			PrefixListIds: []*ec2.PrefixListId{
				{PrefixListId: prefixListsOutput.PrefixLists[0].PrefixListId, Description: aws.String("Prefix List Description")},
			},
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_multidescription(rInt, "ingress"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("aws_security_group.worker", &group),
					testAccCheckSecurityGroupExists("aws_security_group.nat", &nat),
					testAccCheckVPCEndpointExists("aws_vpc_endpoint.s3_endpoint", &endpoint),

					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.rule_1", &group, &rule1, "ingress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.rule_1", "description", "CIDR Description"),

					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.rule_2", &group, &rule2, "ingress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.rule_2", "description", "IPv6 CIDR Description"),

					setupSG,
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.rule_3", &group, &rule3, "ingress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.rule_3", "description", "NAT SG Description"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.rule_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.rule_1"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_security_group_rule.rule_2",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.rule_2"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_security_group_rule.rule_3",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.rule_3"),
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_multidescription(rInt, "egress"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists("aws_security_group.worker", &group),
					testAccCheckSecurityGroupExists("aws_security_group.nat", &nat),
					testAccCheckVPCEndpointExists("aws_vpc_endpoint.s3_endpoint", &endpoint),

					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.rule_1", &group, &rule1, "egress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.rule_1", "description", "CIDR Description"),

					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.rule_2", &group, &rule2, "egress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.rule_2", "description", "IPv6 CIDR Description"),

					setupSG,
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.rule_3", &group, &rule3, "egress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.rule_3", "description", "NAT SG Description"),

					setupPL,
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.rule_4", &group, &rule4, "egress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.rule_4", "description", "Prefix List Description"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.rule_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.rule_1"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_security_group_rule.rule_2",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.rule_2"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_security_group_rule.rule_3",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.rule_3"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_security_group_rule.rule_4",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.rule_4"),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckSecurityGroupRuleAttributes(n string, group *ec2.SecurityGroup, p *ec2.IpPermission, ruleType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Security Group Rule Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Group Rule is set")
		}

		if p == nil {
			p = &ec2.IpPermission{
				FromPort:   aws.Int64(80),
				ToPort:     aws.Int64(8000),
				IpProtocol: aws.String("tcp"),
				IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("10.0.0.0/8")}},
			}
		}

		var matchingRule *ec2.IpPermission
		var rules []*ec2.IpPermission
		if ruleType == "ingress" {
			rules = group.IpPermissions
		} else {
			rules = group.IpPermissionsEgress
		}

		if len(rules) == 0 {
			return fmt.Errorf("No IPPerms")
		}

		for _, r := range rules {
			if p.ToPort != nil && r.ToPort != nil && *p.ToPort != *r.ToPort {
				continue
			}

			if p.FromPort != nil && r.FromPort != nil && *p.FromPort != *r.FromPort {
				continue
			}

			if p.IpProtocol != nil && r.IpProtocol != nil && *p.IpProtocol != *r.IpProtocol {
				continue
			}

			remaining := len(p.IpRanges)
			for _, ip := range p.IpRanges {
				for _, rip := range r.IpRanges {
					if ip.CidrIp == nil || rip.CidrIp == nil {
						continue
					}
					if *ip.CidrIp == *rip.CidrIp {
						remaining--
					}
				}
			}

			if remaining > 0 {
				continue
			}

			remaining = len(p.Ipv6Ranges)
			for _, ip := range p.Ipv6Ranges {
				for _, rip := range r.Ipv6Ranges {
					if ip.CidrIpv6 == nil || rip.CidrIpv6 == nil {
						continue
					}
					if *ip.CidrIpv6 == *rip.CidrIpv6 {
						remaining--
					}
				}
			}

			if remaining > 0 {
				continue
			}

			remaining = len(p.UserIdGroupPairs)
			for _, ip := range p.UserIdGroupPairs {
				for _, rip := range r.UserIdGroupPairs {
					if ip.GroupId == nil || rip.GroupId == nil {
						continue
					}
					if *ip.GroupId == *rip.GroupId {
						remaining--
					}
				}
			}

			if remaining > 0 {
				continue
			}

			remaining = len(p.PrefixListIds)
			for _, pip := range p.PrefixListIds {
				for _, rpip := range r.PrefixListIds {
					if pip.PrefixListId == nil || rpip.PrefixListId == nil {
						continue
					}
					if *pip.PrefixListId == *rpip.PrefixListId {
						remaining--
					}
				}
			}

			if remaining > 0 {
				continue
			}

			matchingRule = r
		}

		if matchingRule != nil {
			log.Printf("[DEBUG] Matching rule found : %s", matchingRule)
			return nil
		}

		return fmt.Errorf("Error here\n\tlooking for %s, wasn't found in %s", p, rules)
	}
}

func testAccSecurityGroupRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		sgID := rs.Primary.Attributes["security_group_id"]
		ruleType := rs.Primary.Attributes["type"]
		protocol := rs.Primary.Attributes["protocol"]
		fromPort := rs.Primary.Attributes["from_port"]
		toPort := rs.Primary.Attributes["to_port"]

		cidrs, err := testAccSecurityGroupRuleImportGetAttrs(rs.Primary.Attributes, "cidr_blocks")
		if err != nil {
			return "", err
		}

		ipv6CIDRs, err := testAccSecurityGroupRuleImportGetAttrs(rs.Primary.Attributes, "ipv6_cidr_blocks")
		if err != nil {
			return "", err
		}

		prefixes, err := testAccSecurityGroupRuleImportGetAttrs(rs.Primary.Attributes, "prefix_list_ids")
		if err != nil {
			return "", err
		}

		var parts []string
		parts = append(parts, sgID)
		parts = append(parts, ruleType)
		parts = append(parts, protocol)
		parts = append(parts, fromPort)
		parts = append(parts, toPort)
		parts = append(parts, *cidrs...)
		parts = append(parts, *ipv6CIDRs...)
		parts = append(parts, *prefixes...)

		if sgSource, ok := rs.Primary.Attributes["source_security_group_id"]; ok {
			parts = append(parts, sgSource)
		}

		if rs.Primary.Attributes["self"] == "true" {
			parts = append(parts, "self")
		}

		return strings.Join(parts, "_"), nil
	}
}

func testAccSecurityGroupRuleImportGetAttrs(attrs map[string]string, key string) (*[]string, error) {
	var values []string
	if countStr, ok := attrs[fmt.Sprintf("%s.#", key)]; ok && countStr != "0" {
		count, err := strconv.Atoi(countStr)
		if err != nil {
			return nil, err
		}
		for i := 0; i < count; i++ {
			values = append(values, attrs[fmt.Sprintf("%s.%d", key, i)])
		}
	}
	return &values, nil
}

func testAccVPCSecurityGroupRuleConfig_ingress(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressIcmpv6(rName string) string {
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

resource "aws_security_group_rule" "test" {
  security_group_id = aws_security_group.test.id
  type              = "ingress"
  from_port         = -1
  to_port           = -1
  protocol          = "icmpv6"
  ipv6_cidr_blocks  = ["::/0"]
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressIPv6(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "tftest" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.tftest.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type             = "ingress"
  protocol         = "6"
  from_port        = 80
  to_port          = 8000
  ipv6_cidr_blocks = ["::/0"]

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressProtocol(rName string) string {
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

resource "aws_security_group_rule" "test" {
  type        = "ingress"
  protocol    = "6"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_issue5310(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type              = "ingress"
  from_port         = 0
  to_port           = 65535
  protocol          = "tcp"
  security_group_id = aws_security_group.test.id
  self              = true
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressClassic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigEC2ClassicRegionProvider(), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.test.id
}
`, rName))
}

func testAccVPCSecurityGroupRuleConfig_egress(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_multidescription(rInt int, rType string) string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf(`
resource "aws_vpc" "tf_sgrule_description_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-multi-desc"
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "s3_endpoint" {
  vpc_id       = aws_vpc.tf_sgrule_description_test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

resource "aws_security_group" "worker" {
  name        = "terraform_test_%[1]d"
  vpc_id      = aws_vpc.tf_sgrule_description_test.id
  description = "Used in the terraform acceptance tests"

  tags = { Name = "tf-sg-rule-description" }
}

resource "aws_security_group" "nat" {
  name        = "terraform_test_%[1]d_nat"
  vpc_id      = aws_vpc.tf_sgrule_description_test.id
  description = "Used in the terraform acceptance tests"

  tags = { Name = "tf-sg-rule-description" }
}

resource "aws_security_group_rule" "rule_1" {
  security_group_id = aws_security_group.worker.id
  description       = "CIDR Description"
  type              = "%[2]s"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "rule_2" {
  security_group_id = aws_security_group.worker.id
  description       = "IPv6 CIDR Description"
  type              = "%[2]s"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  ipv6_cidr_blocks  = ["::/0"]
}

resource "aws_security_group_rule" "rule_3" {
  security_group_id        = aws_security_group.worker.id
  description              = "NAT SG Description"
  type                     = "%[2]s"
  protocol                 = "tcp"
  from_port                = 22
  to_port                  = 22
  source_security_group_id = aws_security_group.nat.id
}
`, rInt, rType))

	if rType == "egress" {
		b.WriteString(`
resource "aws_security_group_rule" "rule_4" {
  security_group_id = aws_security_group.worker.id
  description       = "Prefix List Description"
  type              = "egress"
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  prefix_list_ids   = [aws_vpc_endpoint.s3_endpoint.prefix_list_id]
}
`)
	}

	return b.String()
}

// check for GH-1985 regression
func testAccVPCSecurityGroupRuleConfig_selfReference(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type              = "ingress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  self              = true
  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_partialMatching(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test1" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24", "10.0.4.0/24"]

  security_group_id = aws_security_group.test[0].id
}

resource "aws_security_group_rule" "test2" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.5.0/24"]

  security_group_id = aws_security_group.test[0].id
}

# same a above, but different group, to guard against bad hashing
resource "aws_security_group_rule" "test3" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24", "10.0.4.0/24"]

  security_group_id = aws_security_group.test[1].id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_partialMatchingSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  name   = "%[1]s-${count.index}"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test1" {
  type      = "ingress"
  from_port = 80
  to_port   = 80
  protocol  = "tcp"

  source_security_group_id = aws_security_group.test[0].id
  security_group_id        = aws_security_group.test[1].id
}

resource "aws_security_group_rule" "test2" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24", "10.0.4.0/24"]

  security_group_id = aws_security_group.test[0].id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_prefixListEgress(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]

  tags = {
    Name = %[1]q
  }

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowAll",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  prefix_list_ids   = [aws_vpc_endpoint.test.prefix_list_id]
  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test ingress rule"

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressUpdateDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test ingress rule updated"

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_egressDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test egress rule"

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_egressUpdateDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type        = "egress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test egress rule updated"

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_descriptionAllPorts(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  cidr_blocks       = ["0.0.0.0/0"]
  description       = %[2]q
  from_port         = 0
  protocol          = -1
  security_group_id = aws_security_group.test.id
  to_port           = 0
  type              = "ingress"
}
`, rName, description)
}

func testAccVPCSecurityGroupRuleConfig_descriptionAllPortsNonZeroPorts(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  cidr_blocks       = ["0.0.0.0/0"]
  description       = %[2]q
  from_port         = -1
  protocol          = -1
  security_group_id = aws_security_group.test.id
  to_port           = -1
  type              = "ingress"
}
`, rName, description)
}

func testAccVPCSecurityGroupRuleConfig_multipleSearchingAllProtocolCrash(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %q

  tags = {
    Name = "tf-acc-test-ec2-security-group-rule"
  }
}

resource "aws_security_group_rule" "test1" {
  cidr_blocks       = ["10.0.0.0/8"]
  from_port         = 0
  protocol          = -1
  security_group_id = aws_security_group.test.id
  to_port           = 65535
  type              = "ingress"
}

resource "aws_security_group_rule" "test2" {
  cidr_blocks       = ["172.168.0.0/16"]
  from_port         = 443
  protocol          = "tcp"
  security_group_id = aws_security_group.test.id
  to_port           = 443
  type              = "ingress"
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_race(rName string, n int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test_ingress" {
  count = %[2]d

  security_group_id = aws_security_group.test.id
  type              = "ingress"
  from_port         = count.index
  to_port           = count.index
  protocol          = "tcp"
  cidr_blocks       = ["10.0.0.${count.index}/32"]
}

resource "aws_security_group_rule" "test_egress" {
  count = %[2]d

  security_group_id = aws_security_group.test.id
  type              = "egress"
  from_port         = count.index
  to_port           = count.index
  protocol          = "tcp"
  cidr_blocks       = ["10.0.0.${count.index}/32"]
}
`, rName, n)
}

func testAccVPCSecurityGroupRuleConfig_selfInSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.test.id
  source_security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressSourceAccountID(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  description              = "some description"
  security_group_id        = aws_security_group.test.id
  source_security_group_id = "${data.aws_caller_identity.current.account_id}/${aws_security_group.test.id}"
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_expectInvalidType(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type                     = "invalid"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.test.id
  source_security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_invalidIPv4CIDR(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type              = "ingress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["1.2.3.4/33"]
  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_invalidIPv6CIDR(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  ipv6_cidr_blocks  = ["::/244"]
  security_group_id = aws_security_group.test.id
}
`, rName)
}
