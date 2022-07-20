package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccVPCDefaultSecurityGroup_VPC_basic(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "default"),
					resource.TestCheckResourceAttr(resourceName, "description", "default VPC security group"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "egress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					testAccCheckDefaultSecurityGroupARN(resourceName, &group),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", acctest.ResourcePrefix),
				),
			},
			{
				Config:   testAccVPCDefaultSecurityGroupConfig_basic,
				PlanOnly: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_VPC_empty(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultSecurityGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_Classic_basic(t *testing.T) {
	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupConfig_classic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultSecurityGroupClassicExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "name", "default"),
					resource.TestCheckResourceAttr(resourceName, "description", "default group"),
					resource.TestCheckResourceAttr(resourceName, "vpc_id", ""),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ingress.*", map[string]string{
						"protocol":      "tcp",
						"from_port":     "80",
						"to_port":       "8000",
						"cidr_blocks.#": "1",
						"cidr_blocks.0": "10.0.0.0/8",
					}),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
					testAccCheckDefaultSecurityGroupARNClassic(resourceName, &group),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", acctest.ResourcePrefix),
				),
			},
			{
				Config:   testAccVPCDefaultSecurityGroupConfig_classic(),
				PlanOnly: true,
			},
			{
				Config:                  testAccVPCDefaultSecurityGroupConfig_classic(),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"revoke_rules_on_delete"},
			},
		},
	})
}

func TestAccVPCDefaultSecurityGroup_Classic_empty(t *testing.T) {

	acctest.Skip(t, "This resource does not currently clear tags when adopting the resource")
	// Additional references:
	//  * https://github.com/hashicorp/terraform-provider-aws/issues/14631

	var group ec2.SecurityGroup
	resourceName := "aws_default_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckEC2Classic(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDefaultSecurityGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultSecurityGroupConfig_classicEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultSecurityGroupClassicExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ingress.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "egress.#", "0"),
				),
			},
		},
	})
}

func testAccCheckDefaultSecurityGroupDestroy(s *terraform.State) error {
	// We expect Security Group to still exist
	return nil
}

func testAccCheckDefaultSecurityGroupExists(n string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Default Security Group ID is set")
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

func testAccCheckDefaultSecurityGroupClassicExists(n string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Default Security Group ID is set")
		}

		conn := acctest.ProviderEC2Classic.Meta().(*conns.AWSClient).EC2Conn

		sg, err := tfec2.FindSecurityGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*group = *sg

		return nil
	}
}

func testAccCheckDefaultSecurityGroupARN(resourceName string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "ec2", fmt.Sprintf("security-group/%s", aws.StringValue(group.GroupId)))(s)
	}
}

func testAccCheckDefaultSecurityGroupARNClassic(resourceName string, group *ec2.SecurityGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return acctest.CheckResourceAttrRegionalARNEC2Classic(resourceName, "arn", "ec2", fmt.Sprintf("security-group/%s", aws.StringValue(group.GroupId)))(s)
	}
}

const testAccVPCDefaultSecurityGroupConfig_basic = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-security-group"
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  egress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`

const testAccVPCDefaultSecurityGroupConfig_empty = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-security-group"
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id
}
`

func testAccVPCDefaultSecurityGroupConfig_classic() string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		`
resource "aws_default_security_group" "test" {
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["10.0.0.0/8"]
  }

  tags = {
    Name = "tf-acc-test"
  }
}
`)
}

func testAccVPCDefaultSecurityGroupConfig_classicEmpty() string {
	return acctest.ConfigCompose(
		acctest.ConfigEC2ClassicRegionProvider(),
		`
resource "aws_default_security_group" "test" {
  # No attributes set.
}
`)
}

func TestDefaultSecurityGroupMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0": {
			StateVersion: 0,
			Attributes: map[string]string{
				"name": "test",
			},
			Expected: map[string]string{
				"name":                   "test",
				"revoke_rules_on_delete": "false",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := tfec2.DefaultSecurityGroupMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}

func TestDefaultSecurityGroupMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := tfec2.DefaultSecurityGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = tfec2.DefaultSecurityGroupMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
