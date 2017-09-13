package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDaxSubnetGroup_basic(t *testing.T) {
	var sg dax.SubnetGroup
	config := fmt.Sprintf(testAccAwsDaxSubnetGroupConfig, acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDaxSubnetGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxSubnetGroupExists("aws_dax_subnet_group.bar", &sg),
					resource.TestCheckResourceAttr(
						"aws_dax_subnet_group.bar", "description", "Managed by Terraform"),
				),
			},
		},
	})
}

func TestAccAwsDaxSubnetGroup_update(t *testing.T) {
	var sg dax.SubnetGroup
	rn := "aws_dax_subnet_group.bar"
	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAwsDaxSubnetGroupUpdateConfigPre, ri)
	postConfig := fmt.Sprintf(testAccAwsDaxSubnetGroupUpdateConfigPost, ri)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDaxSubnetGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxSubnetGroupExists(rn, &sg),
					testAccCheckAwsDaxSubnetGroupAttrs(&sg, rn, 1),
				),
			},

			resource.TestStep{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxSubnetGroupExists(rn, &sg),
					testAccCheckAwsDaxSubnetGroupAttrs(&sg, rn, 2),
				),
			},
		},
	})
}

func testAccCheckAwsDaxSubnetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).daxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dax_subnet_group" {
			continue
		}
		res, err := conn.DescribeSubnetGroups(&dax.DescribeSubnetGroupsInput{
			SubnetGroupNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			// Verify the error is what we want
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "SubnetGroupNotFoundFault" {
				continue
			}
			return err
		}
		if len(res.SubnetGroups) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccCheckAwsDaxSubnetGroupExists(n string, sg *dax.SubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No subnet group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).daxconn
		resp, err := conn.DescribeSubnetGroups(&dax.DescribeSubnetGroupsInput{
			SubnetGroupNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return fmt.Errorf("SubnetGroup error: %v", err)
		}

		for _, c := range resp.SubnetGroups {
			if rs.Primary.ID == *c.SubnetGroupName {
				*sg = *c
			}
		}

		if sg == nil {
			return fmt.Errorf("subnet group not found")
		}
		return nil
	}
}

func testAccCheckAwsDaxSubnetGroupAttrs(sg *dax.SubnetGroup, n string, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if len(sg.Subnets) != count {
			return fmt.Errorf("Bad subnet count, expected: %d, got: %d", count, len(sg.Subnets))
		}

		if rs.Primary.Attributes["description"] != *sg.Description {
			return fmt.Errorf("Bad subnet description, expected: %s, got: %s", rs.Primary.Attributes["description"], *sg.Description)
		}

		return nil
	}
}

var testAccAwsDaxSubnetGroupConfig = `
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
    tags {
            Name = "tf-test"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
            Name = "tf-test"
    }
}

resource "aws_dax_subnet_group" "bar" {
    // Including uppercase letters in this name to ensure
    // that we correctly handle the fact that the API
    // normalizes names to lowercase.
    name = "tf-TEST-subnet-%03d"
    subnet_ids = ["${aws_subnet.foo.id}"]
}
`
var testAccAwsDaxSubnetGroupUpdateConfigPre = `
resource "aws_vpc" "foo" {
    cidr_block = "10.0.0.0/16"
    tags {
            Name = "tf-elc-sub-test"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "10.0.1.0/24"
    availability_zone = "us-west-2a"
    tags {
            Name = "tf-test"
    }
}

resource "aws_dax_subnet_group" "bar" {
    name = "tf-test-subnet-%03d"
    description = "tf-test-subnet-group-descr"
    subnet_ids = ["${aws_subnet.foo.id}"]
}
`

var testAccAwsDaxSubnetGroupUpdateConfigPost = `
resource "aws_vpc" "foo" {
    cidr_block = "10.0.0.0/16"
    tags {
            Name = "tf-elc-sub-test"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "10.0.1.0/24"
    availability_zone = "us-west-2a"
    tags {
            Name = "tf-test"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "10.0.2.0/24"
    availability_zone = "us-west-2a"
    tags {
            Name = "tf-test-foo-update"
    }
}

resource "aws_dax_subnet_group" "bar" {
    name = "tf-test-subnet-%03d"
    description = "tf-test-subnet-group-descr-edited"
    subnet_ids = [
			"${aws_subnet.foo.id}",
			"${aws_subnet.bar.id}",
		]
}
`
