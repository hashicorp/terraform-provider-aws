package aws

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsVpc_basic(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	rInt := rand.Intn(16)
	cidr := fmt.Sprintf("172.%d.0.0/16", rInt)
	tag := fmt.Sprintf("terraform-testacc-vpc-data-source-basic-%d", rInt)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcConfig(cidr, tag),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcCheck("data.aws_vpc.by_id", cidr, tag),
					testAccDataSourceAwsVpcCheck("data.aws_vpc.by_cidr", cidr, tag),
					testAccDataSourceAwsVpcCheck("data.aws_vpc.by_tag", cidr, tag),
					testAccDataSourceAwsVpcCheck("data.aws_vpc.by_filter", cidr, tag),
					resource.TestCheckResourceAttr(
						"data.aws_vpc.by_id", "enable_dns_support", "true"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc.by_id", "enable_dns_hostnames", "false"),
					resource.TestCheckResourceAttrSet(
						"data.aws_vpc.by_id", "arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpc_ipv6Associated(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	rInt := rand.Intn(16)
	cidr := fmt.Sprintf("172.%d.0.0/16", rInt)
	tag := fmt.Sprintf("terraform-testacc-vpc-data-source-ipv6-associated-%d", rInt)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcConfigIpv6(cidr, tag),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcCheck("data.aws_vpc.by_id", cidr, tag),
					resource.TestCheckResourceAttrSet(
						"data.aws_vpc.by_id", "ipv6_association_id"),
					resource.TestCheckResourceAttrSet(
						"data.aws_vpc.by_id", "ipv6_cidr_block"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpc_multipleCidr(t *testing.T) {
	rInt := rand.Intn(16)
	rName := "data.aws_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcConfigMultipleCidr(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rName, "cidr_block_associations.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsVpcCheck(name, cidr, tag string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		vpcRs, ok := s.RootModule().Resources["aws_vpc.test"]
		if !ok {
			return fmt.Errorf("can't find aws_vpc.test in state")
		}

		attr := rs.Primary.Attributes

		if attr["id"] != vpcRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"id is %s; want %s",
				attr["id"],
				vpcRs.Primary.Attributes["id"],
			)
		}

		if attr["cidr_block"] != cidr {
			return fmt.Errorf("bad cidr_block %s, expected: %s", attr["cidr_block"], cidr)
		}
		if attr["tags.Name"] != tag {
			return fmt.Errorf("bad Name tag %s", attr["tags.Name"])
		}

		return nil
	}
}

func testAccDataSourceAwsVpcConfigIpv6(cidr, tag string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "test" {
  cidr_block = "%s"
  assign_generated_ipv6_cidr_block = true

  tags {
    Name = "%s"
  }
}

data "aws_vpc" "by_id" {
  id = "${aws_vpc.test.id}"
}`, cidr, tag)
}

func testAccDataSourceAwsVpcConfig(cidr, tag string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "test" {
  cidr_block = "%s"

  tags {
    Name = "%s"
  }
}

data "aws_vpc" "by_id" {
  id = "${aws_vpc.test.id}"
}

data "aws_vpc" "by_cidr" {
  cidr_block = "${aws_vpc.test.cidr_block}"
}

data "aws_vpc" "by_tag" {
  tags {
    Name = "${aws_vpc.test.tags["Name"]}"
  }
}

data "aws_vpc" "by_filter" {
  filter {
    name = "cidr"
    values = ["${aws_vpc.test.cidr_block}"]
  }
}`, cidr, tag)
}

func testAccDataSourceAwsVpcConfigMultipleCidr(octet int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.%d.0.0/16"
}

resource "aws_vpc_ipv4_cidr_block_association" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "172.%d.0.0/16"
}

data "aws_vpc" "test" {
  filter {
    name = "cidr-block-association.cidr-block"
    values = ["${aws_vpc_ipv4_cidr_block_association.test.cidr_block}"]
  }
}
`, octet, octet)
}
