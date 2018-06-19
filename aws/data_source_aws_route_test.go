package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsRoute_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRouteGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsRouteCheck("data.aws_route.by_destination_cidr_block"),
					testAccDataSourceAwsRouteCheck("data.aws_route.by_instance_id"),
					testAccDataSourceAwsRouteCheck("data.aws_route.by_peering_connection_id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataSourceAwsRouteCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		r, ok := s.RootModule().Resources["aws_route.test"]
		if !ok {
			return fmt.Errorf("can't find aws_route.test in state")
		}
		rts, ok := s.RootModule().Resources["aws_route_table.test"]
		if !ok {
			return fmt.Errorf("can't find aws_route_table.test in state")
		}

		attr := rs.Primary.Attributes

		if attr["route_table_id"] != r.Primary.Attributes["route_table_id"] {
			return fmt.Errorf(
				"route_table_id is %s; want %s",
				attr["route_table_id"],
				r.Primary.Attributes["route_table_id"],
			)
		}

		if attr["route_table_id"] != rts.Primary.Attributes["id"] {
			return fmt.Errorf(
				"route_table_id is %s; want %s",
				attr["route_table_id"],
				rts.Primary.Attributes["id"],
			)
		}

		return nil
	}
}

const testAccDataSourceAwsRouteGroupConfig = `
provider "aws" {
  region = "ap-southeast-2"
}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags {
    Name = "terraform-testacc-route-table-data-source"
  }
}

resource "aws_vpc" "dest" {
	cidr_block = "172.17.0.0/16"
  
	tags {
	  Name = "terraform-testacc-route-table-data-source"
	}
}

resource "aws_vpc_peering_connection" "test" {
	peer_vpc_id   = "${aws_vpc.dest.id}"
	vpc_id        = "${aws_vpc.test.id}"
	auto_accept   = true
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"
  tags {
    Name = "tf-acc-route-table-data-source"
  }
}

resource "aws_route_table" "test" {
	vpc_id = "${aws_vpc.test.id}"
	tags {
	  Name = "terraform-testacc-routetable-data-source"
	}
}

resource "aws_route" "pcx" {
	route_table_id  = "${aws_route_table.test.id}"
	vpc_peering_connection_id = "${aws_vpc_peering_connection.test.id}"
	destination_cidr_block    = "10.0.2.0/24"
}

resource "aws_route_table_association" "a" {
    subnet_id = "${aws_subnet.test.id}"
    route_table_id = "${aws_route_table.test.id}"
}

data "aws_ami" "ubuntu" {
	most_recent = true
  
	filter {
	  name   = "name"
	  values = ["ubuntu/images/hvm-ssd/ubuntu-trusty-14.04-amd64-server-*"]
	}
  
	filter {
	  name   = "virtualization-type"
	  values = ["hvm"]
	}
  
	owners = ["099720109477"] # Canonical
  }
  
  resource "aws_instance" "web" {
	ami           = "${data.aws_ami.ubuntu.id}"
	instance_type = "t2.micro"
	subnet_id = "${aws_subnet.test.id}"
	tags {
	  Name = "HelloWorld"
	}
  }


resource "aws_route" "test" {
	route_table_id            = "${aws_route_table.test.id}"
	destination_cidr_block    = "10.0.1.0/24"
	instance_id      = "${aws_instance.web.id}"
	timeouts {
		create ="5m"
	}
}

data "aws_route" "by_peering_connection_id"{
	route_table_id            = "${aws_route_table.test.id}"
	vpc_peering_connection_id = "${aws_route.pcx.vpc_peering_connection_id}"
}

data "aws_route" "by_destination_cidr_block"{
	route_table_id            = "${aws_route_table.test.id}"
	destination_cidr_block    = "10.0.1.0/24"
	depends_on                = ["aws_route.test"]
}

data "aws_route" "by_instance_id"{
	route_table_id            = "${aws_route_table.test.id}"
	instance_id               = "${aws_instance.web.id}"
	depends_on                = ["aws_route.test"]
}


`
