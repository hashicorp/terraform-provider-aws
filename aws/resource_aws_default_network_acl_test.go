package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSDefaultNetworkAcl_basic(t *testing.T) {
	rn := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultNetworkConfig_basic,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultNetworkAcl_basicIpv6Vpc(t *testing.T) {
	rn := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultNetworkConfig_basicIpv6Vpc,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultNetworkAcl_deny_ingress(t *testing.T) {
	// TestAccAWSDefaultNetworkAcl_deny_ingress will deny all Ingress rules, but
	// not Egress. We then expect there to be 3 rules, 2 AWS defaults and 1
	// additional Egress.
	rn := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultNetworkConfig_deny_ingress,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultNetworkAcl_withIpv6Ingress(t *testing.T) {
	rn := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultNetworkConfig_includingIpv6Rule,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultNetworkAcl_SubnetRemoval(t *testing.T) {
	rn := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultNetworkConfig_Subnets,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Here the Subnets have been removed from the Default Network ACL Config,
			// but have not been reassigned. The result is that the Subnets are still
			// there, and we have a non-empty plan
			{
				Config:             testAccAWSDefaultNetworkConfig_Subnets_remove,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultNetworkAcl_SubnetReassign(t *testing.T) {
	rn := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultNetworkAclDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultNetworkConfig_Subnets,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Here we've reassigned the subnets to a different ACL.
			// Without any otherwise association between the `aws_network_acl` and
			// `aws_default_network_acl` resources, we cannot guarantee that the
			// reassignment of the two subnets to the `aws_network_acl` will happen
			// before the update/read on the `aws_default_network_acl` resource.
			// Because of this, there could be a non-empty plan if a READ is done on
			// the default before the reassignment occurs on the other resource.
			//
			// For the sake of testing, here we introduce a depends_on attribute from
			// the default resource to the other acl resource, to ensure the latter's
			// update occurs first, and the former's READ will correctly read zero
			// subnets
			{
				Config: testAccAWSDefaultNetworkConfig_Subnets_move,
			},
			{
				ResourceName:      rn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSDefaultNetworkAclDestroy(s *terraform.State) error {
	// We can't destroy this resource; it comes and goes with the VPC itself.
	return nil
}

const testAccAWSDefaultNetworkConfig_basic = `
resource "aws_vpc" "tftestvpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-basic"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = "${aws_vpc.tftestvpc.default_network_acl_id}"

  tags = {
    Name = "tf-acc-default-acl-basic"
  }
}
`

const testAccAWSDefaultNetworkConfig_includingIpv6Rule = `
resource "aws_vpc" "tftestvpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-including-ipv6-rule"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = "${aws_vpc.tftestvpc.default_network_acl_id}"

  ingress {
    protocol   = -1
    rule_no    = 101
    action     = "allow"
    ipv6_cidr_block = "::/0"
    from_port  = 0
    to_port    = 0
  }

  tags = {
    Name = "tf-acc-default-acl-basic-including-ipv6-rule"
  }
}
`

const testAccAWSDefaultNetworkConfig_deny_ingress = `
resource "aws_vpc" "tftestvpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-deny-ingress"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = "${aws_vpc.tftestvpc.default_network_acl_id}"

  egress {
    protocol   = -1
    rule_no    = 100
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    to_port    = 0
  }

  tags = {
    Name = "tf-acc-default-acl-deny-ingress"
  }
}
`

const testAccAWSDefaultNetworkConfig_Subnets = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-subnets"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-network-acl-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-network-acl-two"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-acl-subnets"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = "${aws_vpc.foo.default_network_acl_id}"

  subnet_ids = ["${aws_subnet.one.id}", "${aws_subnet.two.id}"]

  tags = {
    Name = "tf-acc-default-acl-subnets"
  }
}
`

const testAccAWSDefaultNetworkConfig_Subnets_remove = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-subnets-remove"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-network-acl-subnets-remove-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-network-acl-subnets-remove-two"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-acl-subnets-remove"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = "${aws_vpc.foo.default_network_acl_id}"

  tags = {
    Name = "tf-acc-default-acl-subnets-remove"
  }
}
`

const testAccAWSDefaultNetworkConfig_Subnets_move = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-subnets-move"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-network-acl-subnets-move-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-acc-default-network-acl-subnets-move-two"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = "${aws_vpc.foo.id}"

  subnet_ids = ["${aws_subnet.one.id}", "${aws_subnet.two.id}"]

  tags = {
    Name = "tf-acc-default-acl-subnets-move"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = "${aws_vpc.foo.default_network_acl_id}"

  depends_on = ["aws_network_acl.bar"]

  tags = {
    Name = "tf-acc-default-acl-subnets-move"
  }
}
`

const testAccAWSDefaultNetworkConfig_basicIpv6Vpc = `
resource "aws_vpc" "tftestvpc" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true

	tags = {
		Name = "terraform-testacc-default-network-acl-basic-ipv6-vpc"
	}
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = "${aws_vpc.tftestvpc.default_network_acl_id}"

  tags = {
    Name = "tf-acc-default-acl-subnets-basic-ipv6-vpc"
  }
}
`
