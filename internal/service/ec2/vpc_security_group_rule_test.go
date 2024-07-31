// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSecurityGroupRuleCreateID(t *testing.T) {
	t.Parallel()

	simple := awstypes.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(80),
		ToPort:     aws.Int32(8000),
		IpRanges: []awstypes.IpRange{
			{
				CidrIp: aws.String("10.0.0.0/8"),
			},
		},
	}

	egress := awstypes.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(80),
		ToPort:     aws.Int32(8000),
		IpRanges: []awstypes.IpRange{
			{
				CidrIp: aws.String("10.0.0.0/8"),
			},
		},
	}

	egress_all := awstypes.IpPermission{
		IpProtocol: aws.String("-1"),
		IpRanges: []awstypes.IpRange{
			{
				CidrIp: aws.String("10.0.0.0/8"),
			},
		},
	}

	vpc_security_group_source := awstypes.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(80),
		ToPort:     aws.Int32(8000),
		UserIdGroupPairs: []awstypes.UserIdGroupPair{
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

	security_group_source := awstypes.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(80),
		ToPort:     aws.Int32(8000),
		UserIdGroupPairs: []awstypes.UserIdGroupPair{
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
		Input  awstypes.IpPermission
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
		actual := tfec2.SecurityGroupRuleCreateID("sg-12345", tc.Type, &tc.Input)
		if actual != tc.Output {
			t.Errorf("input: %s - %#v\noutput: %s", tc.Type, tc.Input, actual)
		}
	}
}

func TestAccVPCSecurityGroupRule_Ingress_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressSourceAccountID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "some description"),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestMatchResourceAttr(resourceName, "source_security_group_id", regexache.MustCompile("^[0-9]{12}/sg-[0-9a-z]{17}$")),
					resource.TestCheckResourceAttr(resourceName, "to_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_protocol(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressProtocol(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressIcmpv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "icmpv6"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_egress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "egress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_selfReference(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCSecurityGroupRuleConfig_expectInvalidType(rName),
				ExpectError: regexache.MustCompile(`expected type to be one of \[\"egress\" \"ingress\"\]`),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_expectInvalidCIDR(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCSecurityGroupRuleConfig_invalidIPv4CIDR(rName),
				ExpectError: regexache.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccVPCSecurityGroupRuleConfig_invalidIPv6CIDR(rName),
				ExpectError: regexache.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

// testing partial match implementation
func TestAccVPCSecurityGroupRule_PartialMatching_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_security_group_rule.test1"
	resource2Name := "aws_security_group_rule.test2"
	resource3Name := "aws_security_group_rule.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_partialMatching(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.0", "10.0.2.0/24"),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.1", "10.0.3.0/24"),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.2", "10.0.4.0/24"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.0", "10.0.5.0/24"),
					resource.TestCheckResourceAttr(resource3Name, "cidr_blocks.#", acctest.Ct3),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_security_group_rule.test1"
	resource2Name := "aws_security_group_rule.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_partialMatchingSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resource1Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.#", acctest.Ct3),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_issue5310(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "65535"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sgResourceName := "aws_security_group.test"
	n := 50

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_race(rName, n),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					testAccCheckSecurityGroupRuleCount(ctx, &group, n, n),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_selfSource(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_selfInSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "source_security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "to_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"
	vpceResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_prefixListEgress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_list_ids.0", vpceResourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "egress"),
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

// https://github.com/hashicorp/terraform-provider-aws/issues/26191.
func TestAccVPCSecurityGroupRule_prefixListEmptyString(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCSecurityGroupRuleConfig_prefixListEmptyString(rName),
				ExpectError: regexache.MustCompile(`prefix_list_ids.0 must not be empty`),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_ingressDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TF acceptance test ingress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_egressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TF acceptance test egress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "egress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TF acceptance test ingress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
				),
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressUpdateDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TF acceptance test ingress rule updated"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_egressDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TF acceptance test egress rule"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "egress"),
				),
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_egressUpdateDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "TF acceptance test egress rule updated"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "egress"),
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
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_descriptionAllPorts(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_DescriptionAllPorts_nonZeroPorts(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_descriptionAllPortsNonZeroPorts(rName, "description1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
					resource.TestCheckResourceAttr(resourceName, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6416
func TestAccVPCSecurityGroupRule_MultipleRuleSearching_allProtocolCrash(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_security_group_rule.test1"
	resource2Name := "aws_security_group_rule.test2"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_multipleSearchingAllProtocolCrash(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.0", "10.0.0.0/8"),
					resource.TestCheckNoResourceAttr(resource1Name, names.AttrDescription),
					resource.TestCheckResourceAttr(resource1Name, "from_port", acctest.Ct0),
					resource.TestCheckResourceAttr(resource1Name, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource1Name, names.AttrProtocol, "-1"),
					resource.TestCheckResourceAttr(resource1Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource1Name, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resource1Name, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resource1Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource1Name, "to_port", "65535"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrType, "ingress"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.0", "172.168.0.0/16"),
					resource.TestCheckNoResourceAttr(resource2Name, names.AttrDescription),
					resource.TestCheckResourceAttr(resource2Name, "from_port", "443"),
					resource.TestCheckResourceAttr(resource2Name, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource2Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource2Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource2Name, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resource2Name, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resource2Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource2Name, "to_port", "443"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrType, "ingress"),
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

func TestAccVPCSecurityGroupRule_multiDescription(t *testing.T) {
	ctx := acctest.Context(t)
	var group1, group2 awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_security_group_rule.test1"
	resource2Name := "aws_security_group_rule.test2"
	resource3Name := "aws_security_group_rule.test3"
	resource4Name := "aws_security_group_rule.test4"
	sg1ResourceName := "aws_security_group.test.0"
	sg2ResourceName := "aws_security_group.test.1"
	vpceResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_multiDescription(rName, "ingress"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sg1ResourceName, &group1),
					testAccCheckSecurityGroupExists(ctx, sg2ResourceName, &group2),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrDescription, "CIDR Description"),
					resource.TestCheckResourceAttr(resource1Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource1Name, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource1Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource1Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource1Name, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resource1Name, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resource1Name, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resource1Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource1Name, "to_port", "22"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrType, "ingress"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource2Name, names.AttrDescription, "IPv6 CIDR Description"),
					resource.TestCheckResourceAttr(resource2Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource2Name, "ipv6_cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resource2Name, "ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource2Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource2Name, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resource2Name, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resource2Name, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resource2Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource2Name, "to_port", "22"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrType, "ingress"),
					resource.TestCheckResourceAttr(resource3Name, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource3Name, names.AttrDescription, "Third Description"),
					resource.TestCheckResourceAttr(resource3Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource3Name, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource3Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource3Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource3Name, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resource3Name, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resource3Name, "self", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resource3Name, "source_security_group_id", sg2ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resource3Name, "to_port", "22"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrType, "ingress"),
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
			{
				Config: testAccVPCSecurityGroupRuleConfig_multiDescription(rName, "egress"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sg1ResourceName, &group1),
					testAccCheckSecurityGroupExists(ctx, sg2ResourceName, &group2),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resource1Name, "cidr_blocks.0", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrDescription, "CIDR Description"),
					resource.TestCheckResourceAttr(resource1Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource1Name, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource1Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource1Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource1Name, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resource1Name, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resource1Name, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resource1Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource1Name, "to_port", "22"),
					resource.TestCheckResourceAttr(resource1Name, names.AttrType, "egress"),
					resource.TestCheckResourceAttr(resource2Name, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource2Name, names.AttrDescription, "IPv6 CIDR Description"),
					resource.TestCheckResourceAttr(resource2Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource2Name, "ipv6_cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resource2Name, "ipv6_cidr_blocks.0", "::/0"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource2Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource2Name, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resource2Name, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resource2Name, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resource2Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource2Name, "to_port", "22"),
					resource.TestCheckResourceAttr(resource2Name, names.AttrType, "egress"),
					resource.TestCheckResourceAttr(resource3Name, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource3Name, names.AttrDescription, "Third Description"),
					resource.TestCheckResourceAttr(resource3Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource3Name, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource3Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource3Name, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resource3Name, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resource3Name, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resource3Name, "self", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resource3Name, "source_security_group_id", sg2ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resource3Name, "to_port", "22"),
					resource.TestCheckResourceAttr(resource3Name, names.AttrType, "egress"),
					resource.TestCheckResourceAttr(resource4Name, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource4Name, names.AttrDescription, "Prefix List Description"),
					resource.TestCheckResourceAttr(resource4Name, "from_port", "22"),
					resource.TestCheckResourceAttr(resource4Name, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resource4Name, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resource4Name, "prefix_list_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resource4Name, "prefix_list_ids.0", vpceResourceName, "prefix_list_id"),
					resource.TestCheckResourceAttrPair(resource4Name, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resource4Name, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resource4Name, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resource4Name, "source_security_group_id"),
					resource.TestCheckResourceAttr(resource4Name, "to_port", "22"),
					resource.TestCheckResourceAttr(resource4Name, names.AttrType, "egress"),
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
			{
				ResourceName:      resource4Name,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resource4Name),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_multipleIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressMultipleIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.0", "2001:db8:85a3::/64"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.1", "2001:db8:85a3:2::/64"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_rule_id", ""),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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

func TestAccVPCSecurityGroupRule_Ingress_multiplePrefixLists(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressMultiplePrefixLists(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_rule_id", ""),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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

func TestAccVPCSecurityGroupRule_Ingress_peeredVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressPeeredVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_rule_id"),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckResourceAttrSet(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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

func TestAccVPCSecurityGroupRule_Ingress_ipv4AndIPv6(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_ingressIPv4AndIPv6(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.0", "10.2.0.0/16"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.0", "2001:db8:85a3::/64"),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_rule_id", ""),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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

func TestAccVPCSecurityGroupRule_Ingress_prefixListAndSelf(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_prefixListAndSelf(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sgResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_rule_id", ""),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtTrue),
					resource.TestCheckNoResourceAttr(resourceName, "source_security_group_id"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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

func TestAccVPCSecurityGroupRule_Ingress_prefixListAndSource(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	sg1ResourceName := "aws_security_group.test.0"
	sg2ResourceName := "aws_security_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_prefixListAndSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sg1ResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "from_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, "prefix_list_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", sg1ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "security_group_rule_id", ""),
					resource.TestCheckResourceAttr(resourceName, "self", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "source_security_group_id", sg2ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "to_port", "8000"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
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

func TestAccVPCSecurityGroupRule_protocolChange(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_security_group_rule.test"
	resourceName2 := "aws_security_group_rule.test2"
	sgName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSecurityGroupRuleConfig_protocolChange(rName, "tcp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
					resource.TestCheckResourceAttr(resourceName2, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrType, "ingress"),
				),
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_protocolChange(rName, "udp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "udp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
					resource.TestCheckResourceAttr(resourceName2, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrProtocol, "udp"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrType, "ingress"),
				),
			},
			{
				Config: testAccVPCSecurityGroupRuleConfig_protocolChange(rName, "tcp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSecurityGroupExists(ctx, sgName, &group),
					resource.TestCheckResourceAttr(resourceName, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ingress"),
					resource.TestCheckResourceAttr(resourceName2, "cidr_blocks.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrProtocol, "tcp"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrType, "ingress"),
				),
			},
		},
	})
}

func testAccSecurityGroupRuleImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		sgID := rs.Primary.Attributes["security_group_id"]
		ruleType := rs.Primary.Attributes[names.AttrType]
		protocol := rs.Primary.Attributes[names.AttrProtocol]
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

		if rs.Primary.Attributes["self"] == acctest.CtTrue {
			parts = append(parts, "self")
		}

		return strings.Join(parts, "_"), nil
	}
}

func testAccSecurityGroupRuleImportGetAttrs(attrs map[string]string, key string) (*[]string, error) {
	var values []string
	if countStr, ok := attrs[fmt.Sprintf("%s.#", key)]; ok && countStr != acctest.Ct0 {
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

func testAccVPCSecurityGroupRuleConfig_multiDescription(rName, ruleType string) string {
	config := fmt.Sprintf(`
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
  security_group_id = aws_security_group.test[0].id
  description       = "CIDR Description"
  type              = %[2]q
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "test2" {
  security_group_id = aws_security_group.test[0].id
  description       = "IPv6 CIDR Description"
  type              = %[2]q
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  ipv6_cidr_blocks  = ["::/0"]
}

resource "aws_security_group_rule" "test3" {
  security_group_id        = aws_security_group.test[0].id
  description              = "Third Description"
  type                     = %[2]q
  protocol                 = "tcp"
  from_port                = 22
  to_port                  = 22
  source_security_group_id = aws_security_group.test[1].id
}
`, rName, ruleType)

	if ruleType == "egress" {
		config = acctest.ConfigCompose(config, fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test4" {
  security_group_id = aws_security_group.test[0].id
  description       = "Prefix List Description"
  type              = %[2]q
  protocol          = "tcp"
  from_port         = 22
  to_port           = 22
  prefix_list_ids   = [aws_vpc_endpoint.test.prefix_list_id]
}
`, rName, ruleType))
	}

	return config
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

func testAccVPCSecurityGroupRuleConfig_prefixListEmptyString(rName string) string {
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
  type              = "egress"
  from_port         = 443
  to_port           = 443
  protocol          = "TCP"
  prefix_list_ids   = [""]
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
  name = %[1]q

  tags = {
    Name = %[1]q
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

func testAccVPCSecurityGroupRuleConfig_ingressMultipleIPv6(rName string) string {
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
  type             = "ingress"
  protocol         = "6"
  from_port        = 80
  to_port          = 8000
  ipv6_cidr_blocks = ["2001:db8:85a3::/64", "2001:db8:85a3:2::/64"]

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressMultiplePrefixLists(rName string) string {
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

resource "aws_ec2_managed_prefix_list" "test" {
  count = 2

  address_family = "IPv4"
  max_entries    = 1
  name           = "%[1]s-${count.index}"
}

resource "aws_security_group_rule" "test" {
  type            = "ingress"
  protocol        = "6"
  from_port       = 80
  to_port         = 8000
  prefix_list_ids = aws_ec2_managed_prefix_list.test[*].id

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_ingressPeeredVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
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

resource "aws_vpc" "other" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "other" {
  provider = "awsalternate"

  vpc_id = aws_vpc.other.id
  name   = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "other" {
  provider = "awsalternate"
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id        = aws_vpc.test.id
  peer_vpc_id   = aws_vpc.other.id
  peer_owner_id = data.aws_caller_identity.other.account_id
  peer_region   = %[2]q
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection_accepter" "other" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  type      = "ingress"
  protocol  = "6"
  from_port = 80
  to_port   = 8000

  source_security_group_id = "${data.aws_caller_identity.other.account_id}/${aws_security_group.other.id}"

  security_group_id = aws_security_group.test.id

  depends_on = [aws_vpc_peering_connection_accepter.other]
}
`, rName, acctest.Region()))
}

func testAccVPCSecurityGroupRuleConfig_ingressIPv4AndIPv6(rName string) string {
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
  type             = "ingress"
  protocol         = "6"
  from_port        = 80
  to_port          = 8000
  cidr_blocks      = ["10.2.0.0/16"]
  ipv6_cidr_blocks = ["2001:db8:85a3::/64"]

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_prefixListAndSelf(rName string) string {
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

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv6"
  max_entries    = 2
  name           = %[1]q
}

resource "aws_security_group_rule" "test" {
  type            = "ingress"
  protocol        = "6"
  from_port       = 80
  to_port         = 8000
  prefix_list_ids = [aws_ec2_managed_prefix_list.test.id]
  self            = true

  security_group_id = aws_security_group.test.id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_prefixListAndSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id
  name   = "%[1]s-${count.index}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_security_group_rule" "test" {
  type            = "ingress"
  protocol        = "6"
  from_port       = 80
  to_port         = 8000
  prefix_list_ids = [aws_ec2_managed_prefix_list.test.id]

  source_security_group_id = aws_security_group.test[1].id

  security_group_id = aws_security_group.test[0].id
}
`, rName)
}

func testAccVPCSecurityGroupRuleConfig_protocolChange(rName, protocol string) string {
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
  type              = "ingress"
  from_port         = 9443
  to_port           = 9443
  protocol          = %[2]q
  cidr_blocks       = [aws_vpc.test.cidr_block]
  security_group_id = aws_security_group.test.id
}

resource "aws_security_group_rule" "test2" {
  type              = "ingress"
  from_port         = 8989
  to_port           = 8989
  protocol          = %[2]q
  cidr_blocks       = [aws_vpc.test.cidr_block]
  security_group_id = aws_security_group.test.id
}
`, rName, protocol)
}
