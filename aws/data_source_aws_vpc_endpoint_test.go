package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsVpcEndpoint_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcEndpointCheckExists("data.aws_vpc_endpoint.s3"),
					resource.TestCheckResourceAttrSet("data.aws_vpc_endpoint.s3", "prefix_list_id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_byId(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfigById,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcEndpointCheckExists("data.aws_vpc_endpoint.by_id"),
					resource.TestCheckResourceAttrSet("data.aws_vpc_endpoint.by_id", "prefix_list_id"),
					resource.TestCheckResourceAttrSet("data.aws_vpc_endpoint.by_id", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_withRouteTable(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointWithRouteTableConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcEndpointCheckExists("data.aws_vpc_endpoint.s3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint.s3", "route_table_ids.#", "1"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSourceAwsVpcEndpointCheckExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		vpceRs, ok := s.RootModule().Resources["aws_vpc_endpoint.s3"]
		if !ok {
			return fmt.Errorf("can't find aws_vpc_endpoint.s3 in state")
		}

		attr := rs.Primary.Attributes

		if attr["id"] != vpceRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"id is %s; want %s",
				attr["id"],
				vpceRs.Primary.Attributes["id"],
			)
		}

		return nil
	}
}

const testAccDataSourceAwsVpcEndpointConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags {
	  Name = "terraform-testacc-vpc-endpoint-data-source-foo"
  }
}

resource "aws_vpc_endpoint" "s3" {
    vpc_id = "${aws_vpc.foo.id}"
    service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
  state = "available"

  depends_on = ["aws_vpc_endpoint.s3"]
}
`

const testAccDataSourceAwsVpcEndpointConfigById = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags {
	  Name = "terraform-testacc-vpc-endpoint-data-source-foo"
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_vpc_endpoint" "by_id" {
  id = "${aws_vpc_endpoint.s3.id}"
}
`

const testAccDataSourceAwsVpcEndpointWithRouteTableConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags {
	  Name = "terraform-testacc-vpc-endpoint-data-source-foo"
  }
}

resource "aws_route_table" "rt" {
    vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "s3" {
    vpc_id = "${aws_vpc.foo.id}"
    service_name = "com.amazonaws.us-west-2.s3"
	route_table_ids = ["${aws_route_table.rt.id}"]
}

data "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
  state = "available"

  depends_on = ["aws_vpc_endpoint.s3"]
}
`
