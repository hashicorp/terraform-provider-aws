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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestIpPermissionIDHash(t *testing.T) {
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
	rInt := sdkacctest.RandInt()

	testRuleCount := func(*terraform.State) error {
		if len(group.IpPermissions) != 1 {
			return fmt.Errorf("Wrong Security Group rule count, expected %d, got %d",
				1, len(group.IpPermissions))
		}

		rule := group.IpPermissions[0]
		if aws.Int64Value(rule.FromPort) != int64(80) {
			return fmt.Errorf("Wrong Security Group port setting, expected %d, got %d",
				80, aws.Int64Value(rule.FromPort))
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIngressConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.ingress_1", &group, nil, "ingress"),
					resource.TestCheckResourceAttr(
						"aws_security_group_rule.ingress_1", "from_port", "80"),
					testRuleCount,
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_IngressSourceWithAccount_id(t *testing.T) {
	var group ec2.SecurityGroup

	rInt := sdkacctest.RandInt()

	ruleName := "aws_security_group_rule.allow_self"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRule_Ingress_Source_with_AccountID(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					resource.TestCheckResourceAttrPair(
						ruleName, "security_group_id", "aws_security_group.web", "id"),
					resource.TestMatchResourceAttr(
						ruleName, "source_security_group_id", regexp.MustCompile("^[0-9]{12}/sg-[0-9a-z]{17}$")),
					resource.TestCheckResourceAttr(
						ruleName, "description", "some description"),
					resource.TestCheckResourceAttr(
						ruleName, "from_port", "0"),
					resource.TestCheckResourceAttr(
						ruleName, "to_port", "0"),
					resource.TestCheckResourceAttr(
						ruleName, "protocol", "-1"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_protocol(t *testing.T) {
	var group ec2.SecurityGroup

	testRuleCount := func(*terraform.State) error {
		if len(group.IpPermissions) != 1 {
			return fmt.Errorf("Wrong Security Group rule count, expected %d, got %d",
				1, len(group.IpPermissions))
		}

		rule := group.IpPermissions[0]
		if *rule.FromPort != int64(80) {
			return fmt.Errorf("Wrong Security Group port setting, expected %d, got %d",
				80, aws.Int64Value(rule.FromPort))
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIngress_protocolConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.ingress_1", &group, nil, "ingress"),
					resource.TestCheckResourceAttr(
						"aws_security_group_rule.ingress_1", "from_port", "80"),
					testRuleCount,
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_icmpv6(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_security_group_rule.test"
	sgResourceName := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIngress_icmpv6Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists(sgResourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "icmpv6"),
					resource.TestCheckResourceAttr(resourceName, "type", "ingress"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ipv6_cidr_blocks.0", "::/0"),
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

	testRuleCount := func(*terraform.State) error {
		if len(group.IpPermissions) != 1 {
			return fmt.Errorf("Wrong Security Group rule count, expected %d, got %d",
				1, len(group.IpPermissions))
		}

		rule := group.IpPermissions[0]
		if *rule.FromPort != int64(80) {
			return fmt.Errorf("Wrong Security Group port setting, expected %d, got %d",
				80, aws.Int64Value(rule.FromPort))
		}

		ipv6Address := rule.Ipv6Ranges[0]
		if *ipv6Address.CidrIpv6 != "::/0" {
			return fmt.Errorf("Wrong Security Group IPv6 address, expected %s, got %s",
				"::/0", *ipv6Address.CidrIpv6)
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIngress_ipv6Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testRuleCount,
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Ingress_classic(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	testRuleCount := func(*terraform.State) error {
		if len(group.IpPermissions) != 1 {
			return fmt.Errorf("Wrong Security Group rule count, expected %d, got %d",
				1, len(group.IpPermissions))
		}

		rule := group.IpPermissions[0]
		if *rule.FromPort != int64(80) {
			return fmt.Errorf("Wrong Security Group port setting, expected %d, got %d",
				80, aws.Int64Value(rule.FromPort))
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIngressClassicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.ingress_1", &group, nil, "ingress"),
					resource.TestCheckResourceAttr(
						"aws_security_group_rule.ingress_1", "from_port", "80"),
					testRuleCount,
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_multiIngress(t *testing.T) {
	var group ec2.SecurityGroup

	testMultiRuleCount := func(*terraform.State) error {
		if len(group.IpPermissions) != 2 {
			return fmt.Errorf("Wrong Security Group rule count, expected %d, got %d",
				2, len(group.IpPermissions))
		}

		var rule *ec2.IpPermission
		for _, r := range group.IpPermissions {
			if *r.FromPort == int64(80) {
				rule = r
			}
		}

		if *rule.ToPort != int64(8000) {
			return fmt.Errorf("Wrong Security Group port 2 setting, expected %d, got %d",
				8000, aws.Int64Value(rule.ToPort))
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleMultiIngressConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testMultiRuleCount,
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress_2",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress_2"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_egress(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleEgressConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.egress_1", &group, nil, "egress"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.egress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.egress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_selfReference(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleSelfReferenceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.self",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.self"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_expectInvalidTypeError(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecurityGroupRuleExpectInvalidType(rInt),
				ExpectError: regexp.MustCompile(`expected type to be one of \[ingress egress\]`),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_expectInvalidCIDR(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccSecurityGroupRuleInvalidIPv4CIDR(rInt),
				ExpectError: regexp.MustCompile("invalid CIDR address: 1.2.3.4/33"),
			},
			{
				Config:      testAccSecurityGroupRuleInvalidIPv6CIDR(rInt),
				ExpectError: regexp.MustCompile("invalid CIDR address: ::/244"),
			},
		},
	})
}

// testing partial match implementation
func TestAccVPCSecurityGroupRule_PartialMatching_basic(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	p := ec2.IpPermission{
		FromPort:   aws.Int64(80),
		ToPort:     aws.Int64(80),
		IpProtocol: aws.String("tcp"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("10.0.2.0/24")},
			{CidrIp: aws.String("10.0.3.0/24")},
			{CidrIp: aws.String("10.0.4.0/24")},
		},
	}

	o := ec2.IpPermission{
		FromPort:   aws.Int64(80),
		ToPort:     aws.Int64(80),
		IpProtocol: aws.String("tcp"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("10.0.5.0/24")},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulePartialMatchingConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.ingress", &group, &p, "ingress"),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.other", &group, &o, "ingress"),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.nat_ingress", &group, &o, "ingress"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_security_group_rule.other",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.other"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_security_group_rule.nat_ingress",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.nat_ingress"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_PartialMatching_source(t *testing.T) {
	var group ec2.SecurityGroup
	var nat ec2.SecurityGroup
	var p ec2.IpPermission
	rInt := sdkacctest.RandInt()

	// This function creates the expected IPPermission with the group id from an
	// external security group, needed because Security Group IDs are generated on
	// AWS side and can't be known ahead of time.
	setupSG := func(*terraform.State) error {
		if nat.GroupId == nil {
			return fmt.Errorf("Error: nat group has nil GroupID")
		}

		p = ec2.IpPermission{
			FromPort:   aws.Int64(80),
			ToPort:     aws.Int64(80),
			IpProtocol: aws.String("tcp"),
			UserIdGroupPairs: []*ec2.UserIdGroupPair{
				{GroupId: nat.GroupId},
			},
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulePartialMatching_SourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleExists("aws_security_group.nat", &nat),
					setupSG,
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.source_ingress", &group, &p, "ingress"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.source_ingress",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.source_ingress"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_issue5310(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIssue5310,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.issue_5310", &group),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.issue_5310",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.issue_5310"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_race(t *testing.T) {
	var group ec2.SecurityGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleRace,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.race", &group),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_selfSource(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleSelfInSource(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.allow_self",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.allow_self"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_prefixListEgress(t *testing.T) {
	var group ec2.SecurityGroup
	var endpoint ec2.VpcEndpoint
	var p ec2.IpPermission

	// This function creates the expected IPPermission with the prefix list ID from
	// the VPC Endpoint created in the test
	setupSG := func(*terraform.State) error {
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

		p = ec2.IpPermission{
			IpProtocol: aws.String("-1"),
			PrefixListIds: []*ec2.PrefixListId{
				{PrefixListId: prefixListsOutput.PrefixLists[0].PrefixListId},
			},
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRulePrefixListEgressConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.egress", &group),
					// lookup info on the VPC Endpoint created, to populate the expected
					// IP Perm
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.s3_endpoint", &endpoint),
					setupSG,
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.egress_1", &group, &p, "egress"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.egress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.egress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_ingressDescription(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIngressDescriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.ingress_1", &group, nil, "ingress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.ingress_1", "description", "TF acceptance test ingress rule"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_egressDescription(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleEgressDescriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.egress_1", &group, nil, "egress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.egress_1", "description", "TF acceptance test egress rule"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.egress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.egress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_IngressDescription_updates(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleIngressDescriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.ingress_1", &group, nil, "ingress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.ingress_1", "description", "TF acceptance test ingress rule"),
				),
			},

			{
				Config: testAccSecurityGroupRuleIngress_updateDescriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.ingress_1", &group, nil, "ingress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.ingress_1", "description", "TF acceptance test ingress rule updated"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.ingress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.ingress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_EgressDescription_updates(t *testing.T) {
	var group ec2.SecurityGroup
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleEgressDescriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.egress_1", &group, nil, "egress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.egress_1", "description", "TF acceptance test egress rule"),
				),
			},

			{
				Config: testAccSecurityGroupRuleEgress_updateDescriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.web", &group),
					testAccCheckSecurityGroupRuleAttributes("aws_security_group_rule.egress_1", &group, nil, "egress"),
					resource.TestCheckResourceAttr("aws_security_group_rule.egress_1", "description", "TF acceptance test egress rule updated"),
				),
			},
			{
				ResourceName:      "aws_security_group_rule.egress_1",
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc("aws_security_group_rule.egress_1"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_Description_allPorts(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	securityGroupResourceName := "aws_security_group.test"
	resourceName := "aws_security_group_rule.test"

	rule1 := ec2.IpPermission{
		IpProtocol: aws.String("-1"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("0.0.0.0/0"), Description: aws.String("description1")},
		},
	}

	rule2 := ec2.IpPermission{
		IpProtocol: aws.String("-1"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("0.0.0.0/0"), Description: aws.String("description2")},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleDescriptionAllPortsConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists(securityGroupResourceName, &group),
					testAccCheckSecurityGroupRuleAttributes(resourceName, &group, &rule1, "ingress"),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityGroupRuleDescriptionAllPortsConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists(securityGroupResourceName, &group),
					testAccCheckSecurityGroupRuleAttributes(resourceName, &group, &rule2, "ingress"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
				),
			},
		},
	})
}

func TestAccVPCSecurityGroupRule_DescriptionAllPorts_nonZeroPorts(t *testing.T) {
	var group ec2.SecurityGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	securityGroupResourceName := "aws_security_group.test"
	resourceName := "aws_security_group_rule.test"

	rule1 := ec2.IpPermission{
		IpProtocol: aws.String("-1"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("0.0.0.0/0"), Description: aws.String("description1")},
		},
	}

	rule2 := ec2.IpPermission{
		IpProtocol: aws.String("-1"),
		IpRanges: []*ec2.IpRange{
			{CidrIp: aws.String("0.0.0.0/0"), Description: aws.String("description2")},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleDescriptionAllPortsNonZeroPortsConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists(securityGroupResourceName, &group),
					testAccCheckSecurityGroupRuleAttributes(resourceName, &group, &rule1, "ingress"),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "-1"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "-1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccSecurityGroupRuleImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccSecurityGroupRuleDescriptionAllPortsNonZeroPortsConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists(securityGroupResourceName, &group),
					testAccCheckSecurityGroupRuleAttributes(resourceName, &group, &rule2, "ingress"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
					resource.TestCheckResourceAttr(resourceName, "from_port", "0"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "-1"),
					resource.TestCheckResourceAttr(resourceName, "to_port", "0"),
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
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleMultipleRuleSearchingAllProtocolCrashConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists(securityGroupResourceName, &group),
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
		CheckDestroy:      testAccCheckSecurityGroupRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupRuleMultiDescription(rInt, "ingress"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.worker", &group),
					testAccCheckSecurityGroupRuleExists("aws_security_group.nat", &nat),
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.s3_endpoint", &endpoint),

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
				Config: testAccSecurityGroupRuleMultiDescription(rInt, "egress"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityGroupRuleExists("aws_security_group.worker", &group),
					testAccCheckSecurityGroupRuleExists("aws_security_group.nat", &nat),
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.s3_endpoint", &endpoint),

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

func testAccCheckSecurityGroupRuleDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_security_group" {
			continue
		}

		_, err := tfec2.FindSecurityGroupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Security Group (%s) still exists.", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSecurityGroupRuleExists(n string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Group is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		sg, err := tfec2.FindSecurityGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*group = *sg

		return nil
	}
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

func testAccSecurityGroupRuleIngressConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name        = "terraform_test_%d"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "ingress_1" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.web.id
}
`, rInt)
}

const testAccSecurityGroupRuleIngress_icmpv6Config = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group_rule" "test" {
  security_group_id = aws_security_group.test.id
  type              = "ingress"
  from_port         = -1
  to_port           = -1
  protocol          = "icmpv6"
  ipv6_cidr_blocks  = ["::/0"]
}
`

const testAccSecurityGroupRuleIngress_ipv6Config = `
resource "aws_vpc" "tftest" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-ingress-ipv6"
  }
}

resource "aws_security_group" "web" {
  vpc_id = aws_vpc.tftest.id

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "ingress_1" {
  type             = "ingress"
  protocol         = "6"
  from_port        = 80
  to_port          = 8000
  ipv6_cidr_blocks = ["::/0"]

  security_group_id = aws_security_group.web.id
}
`

const testAccSecurityGroupRuleIngress_protocolConfig = `
resource "aws_vpc" "tftest" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-ingress-protocol"
  }
}

resource "aws_security_group" "web" {
  vpc_id = aws_vpc.tftest.id

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "ingress_1" {
  type        = "ingress"
  protocol    = "6"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.web.id
}
`

const testAccSecurityGroupRuleIssue5310 = `
resource "aws_security_group" "issue_5310" {
  name        = "terraform-test-issue_5310"
  description = "SG for test of issue 5310"
}

resource "aws_security_group_rule" "issue_5310" {
  type              = "ingress"
  from_port         = 0
  to_port           = 65535
  protocol          = "tcp"
  security_group_id = aws_security_group.issue_5310.id
  self              = true
}
`

func testAccSecurityGroupRuleIngressClassicConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name        = "terraform_test_%d"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "ingress_1" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.web.id
}
`, rInt)
}

func testAccSecurityGroupRuleEgressConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name        = "terraform_test_%d"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "egress_1" {
  type        = "egress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.web.id
}
`, rInt)
}

const testAccSecurityGroupRuleMultiIngressConfig = `
resource "aws_security_group" "web" {
  name        = "terraform_acceptance_test_example_2"
  description = "Used in the terraform acceptance tests"
}

resource "aws_security_group" "worker" {
  name        = "terraform_acceptance_test_example_worker"
  description = "Used in the terraform acceptance tests"
}

resource "aws_security_group_rule" "ingress_1" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 22
  to_port     = 22
  cidr_blocks = ["10.0.0.0/8"]

  security_group_id = aws_security_group.web.id
}

resource "aws_security_group_rule" "ingress_2" {
  type      = "ingress"
  protocol  = "tcp"
  from_port = 80
  to_port   = 8000
  self      = true

  security_group_id = aws_security_group.web.id
}
`

func testAccSecurityGroupRuleMultiDescription(rInt int, rType string) string {
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
const testAccSecurityGroupRuleSelfReferenceConfig = `
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-self-ref"
  }
}

resource "aws_security_group" "web" {
  name   = "main"
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "sg-self-test"
  }
}

resource "aws_security_group_rule" "self" {
  type              = "ingress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  self              = true
  security_group_id = aws_security_group.web.id
}
`

func testAccSecurityGroupRulePartialMatchingConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-partial-match"
  }
}

resource "aws_security_group" "web" {
  name   = "tf-other-%d"
  vpc_id = aws_vpc.default.id

  tags = {
    Name = "tf-other-sg"
  }
}

resource "aws_security_group" "nat" {
  name   = "tf-nat-%d"
  vpc_id = aws_vpc.default.id

  tags = {
    Name = "tf-nat-sg"
  }
}

resource "aws_security_group_rule" "ingress" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24", "10.0.4.0/24"]

  security_group_id = aws_security_group.web.id
}

resource "aws_security_group_rule" "other" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.5.0/24"]

  security_group_id = aws_security_group.web.id
}

# same a above, but different group, to guard against bad hashing
resource "aws_security_group_rule" "nat_ingress" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24", "10.0.4.0/24"]

  security_group_id = aws_security_group.nat.id
}
`, rInt, rInt)
}

func testAccSecurityGroupRulePartialMatching_SourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-partial-match"
  }
}

resource "aws_security_group" "web" {
  name   = "tf-other-%d"
  vpc_id = aws_vpc.default.id

  tags = {
    Name = "tf-other-sg"
  }
}

resource "aws_security_group" "nat" {
  name   = "tf-nat-%d"
  vpc_id = aws_vpc.default.id

  tags = {
    Name = "tf-nat-sg"
  }
}

resource "aws_security_group_rule" "source_ingress" {
  type      = "ingress"
  from_port = 80
  to_port   = 80
  protocol  = "tcp"

  source_security_group_id = aws_security_group.nat.id
  security_group_id        = aws_security_group.web.id
}

resource "aws_security_group_rule" "other_ingress" {
  type        = "ingress"
  from_port   = 80
  to_port     = 80
  protocol    = "tcp"
  cidr_blocks = ["10.0.2.0/24", "10.0.3.0/24", "10.0.4.0/24"]

  security_group_id = aws_security_group.web.id
}
`, rInt, rInt)
}

const testAccSecurityGroupRulePrefixListEgressConfig = `
resource "aws_vpc" "tf_sg_prefix_list_egress_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-prefix-list-egress"
  }
}

resource "aws_route_table" "default" {
  vpc_id = aws_vpc.tf_sg_prefix_list_egress_test.id
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "s3_endpoint" {
  vpc_id          = aws_vpc.tf_sg_prefix_list_egress_test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.default.id]

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

resource "aws_security_group" "egress" {
  name        = "terraform_acceptance_test_prefix_list_egress"
  description = "Used in the terraform acceptance tests"
  vpc_id      = aws_vpc.tf_sg_prefix_list_egress_test.id
}

resource "aws_security_group_rule" "egress_1" {
  type              = "egress"
  protocol          = "-1"
  from_port         = 0
  to_port           = 0
  prefix_list_ids   = [aws_vpc_endpoint.s3_endpoint.prefix_list_id]
  security_group_id = aws_security_group.egress.id
}
`

func testAccSecurityGroupRuleIngressDescriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name        = "terraform_test_%d"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "ingress_1" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test ingress rule"

  security_group_id = aws_security_group.web.id
}
`, rInt)
}

func testAccSecurityGroupRuleIngress_updateDescriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name        = "terraform_test_%d"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "ingress_1" {
  type        = "ingress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test ingress rule updated"

  security_group_id = aws_security_group.web.id
}
`, rInt)
}

func testAccSecurityGroupRuleEgressDescriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name        = "terraform_test_%d"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "egress_1" {
  type        = "egress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test egress rule"

  security_group_id = aws_security_group.web.id
}
`, rInt)
}

func testAccSecurityGroupRuleEgress_updateDescriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "web" {
  name        = "terraform_test_%d"
  description = "Used in the terraform acceptance tests"

  tags = {
    Name = "tf-acc-test"
  }
}

resource "aws_security_group_rule" "egress_1" {
  type        = "egress"
  protocol    = "tcp"
  from_port   = 80
  to_port     = 8000
  cidr_blocks = ["10.0.0.0/8"]
  description = "TF acceptance test egress rule updated"

  security_group_id = aws_security_group.web.id
}
`, rInt)
}

func testAccSecurityGroupRuleDescriptionAllPortsConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %q

  tags = {
    Name = "tf-acc-test-ec2-security-group-rule"
  }
}

resource "aws_security_group_rule" "test" {
  cidr_blocks       = ["0.0.0.0/0"]
  description       = %q
  from_port         = 0
  protocol          = -1
  security_group_id = aws_security_group.test.id
  to_port           = 0
  type              = "ingress"
}
`, rName, description)
}

func testAccSecurityGroupRuleDescriptionAllPortsNonZeroPortsConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %q

  tags = {
    Name = "tf-acc-test-ec2-security-group-rule"
  }
}

resource "aws_security_group_rule" "test" {
  cidr_blocks       = ["0.0.0.0/0"]
  description       = %q
  from_port         = -1
  protocol          = -1
  security_group_id = aws_security_group.test.id
  to_port           = -1
  type              = "ingress"
}
`, rName, description)
}

func testAccSecurityGroupRuleMultipleRuleSearchingAllProtocolCrashConfig(rName string) string {
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

var testAccSecurityGroupRuleRace = func() string {
	var b bytes.Buffer
	iterations := 50
	b.WriteString(fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-race"
  }
}

resource "aws_security_group" "race" {
  name   = "tf-sg-rule-race-group-%d"
  vpc_id = aws_vpc.default.id
}
`, sdkacctest.RandInt()))
	for i := 1; i < iterations; i++ {
		b.WriteString(fmt.Sprintf(`
resource "aws_security_group_rule" "ingress%d" {
  security_group_id = aws_security_group.race.id
  type              = "ingress"
  from_port         = %d
  to_port           = %d
  protocol          = "tcp"
  cidr_blocks       = ["10.0.0.%d/32"]
}

resource "aws_security_group_rule" "egress%d" {
  security_group_id = aws_security_group.race.id
  type              = "egress"
  from_port         = %d
  to_port           = %d
  protocol          = "tcp"
  cidr_blocks       = ["10.0.0.%d/32"]
}
`, i, i, i, i, i, i, i, i))
	}
	return b.String()
}()

func testAccSecurityGroupRuleSelfInSource(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-self-ingress"
  }
}

resource "aws_security_group" "web" {
  name        = "allow_all-%d"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "allow_self" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.web.id
  source_security_group_id = aws_security_group.web.id
}
`, rInt)
}

func testAccSecurityGroupRule_Ingress_Source_with_AccountID(rInt int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-self-ingress"
  }
}

resource "aws_security_group" "web" {
  name        = "allow_all-%d"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "allow_self" {
  type                     = "ingress"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  description              = "some description"
  security_group_id        = aws_security_group.web.id
  source_security_group_id = "${data.aws_caller_identity.current.account_id}/${aws_security_group.web.id}"
}
`, rInt)
}

func testAccSecurityGroupRuleExpectInvalidType(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-rule-invalid-type"
  }
}

resource "aws_security_group" "web" {
  name        = "allow_all-%d"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.foo.id
}

resource "aws_security_group_rule" "allow_self" {
  type                     = "foobar"
  from_port                = 0
  to_port                  = 0
  protocol                 = "-1"
  security_group_id        = aws_security_group.web.id
  source_security_group_id = aws_security_group.web.id
}
`, rInt)
}

func testAccSecurityGroupRuleInvalidIPv4CIDR(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "foo" {
  name = "testing-failure-%d"
}

resource "aws_security_group_rule" "ing" {
  type              = "ingress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["1.2.3.4/33"]
  security_group_id = aws_security_group.foo.id
}
`, rInt)
}

func testAccSecurityGroupRuleInvalidIPv6CIDR(rInt int) string {
	return fmt.Sprintf(`
resource "aws_security_group" "foo" {
  name = "testing-failure-%d"
}

resource "aws_security_group_rule" "ing" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  ipv6_cidr_blocks  = ["::/244"]
  security_group_id = aws_security_group.foo.id
}
`, rInt)
}
