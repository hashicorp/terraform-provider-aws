package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/neptune"
)

func TestAccAWSNeptuneSubnetGroup_basic(t *testing.T) {
	var v neptune.DBSubnetGroup

	testCheck := func(*terraform.State) error {
		return nil
	}

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccNeptuneSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "description", "Managed by Terraform"),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSNeptuneSubnetGroup_namePrefix(t *testing.T) {
	var v neptune.DBSubnetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccNeptuneSubnetGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_subnet_group.test", "name", regexp.MustCompile("^tf_test-")),
				),
			},
		},
	})
}

func TestAccAWSNeptuneSubnetGroup_generatedName(t *testing.T) {
	var v neptune.DBSubnetGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccNeptuneSubnetGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.test", &v),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/2603 and
// https://github.com/hashicorp/terraform/issues/2664
func TestAccAWSNeptuneSubnetGroup_withUndocumentedCharacters(t *testing.T) {
	var v neptune.DBSubnetGroup

	testCheck := func(*terraform.State) error {
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccNeptuneSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.underscores", &v),
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.periods", &v),
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.spaces", &v),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSNeptuneSubnetGroup_updateDescription(t *testing.T) {
	var v neptune.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccNeptuneSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},

			resource.TestStep{
				Config: testAccNeptuneSubnetGroupConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "description", "foo description updated"),
				),
			},
		},
	})
}

func testAccCheckNeptuneSubnetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_subnet_group" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeDBSubnetGroups(
			&neptune.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
		if err == nil {
			if len(resp.DBSubnetGroups) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		neptuneerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if neptuneerr.Code() != "DBSubnetGroupNotFoundFault" {
			return err
		}
	}

	return nil
}

func testAccCheckNeptuneSubnetGroupExists(n string, v *neptune.DBSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn
		resp, err := conn.DescribeDBSubnetGroups(
			&neptune.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
		if err != nil {
			return err
		}
		if len(resp.DBSubnetGroups) == 0 {
			return fmt.Errorf("DbSubnetGroup not found")
		}

		*v = *resp.DBSubnetGroups[0]

		return nil
	}
}

func testAccNeptuneSubnetGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "terraform-testacc-neptune-subnet-group"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags {
		Name = "tf-acc-neptune-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags {
		Name = "tf-acc-neptune-subnet-group-2"
	}
}

resource "aws_neptune_subnet_group" "foo" {
	name = "%s"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags {
		Name = "tf-neptunesubnet-group-test"
	}
}`, rName)
}

func testAccNeptuneSubnetGroupConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "terraform-testacc-neptune-subnet-group-updated-description"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags {
		Name = "tf-acc-neptune-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags {
		Name = "tf-acc-neptune-subnet-group-2"
	}
}

resource "aws_neptune_subnet_group" "foo" {
	name = "%s"
	description = "foo description updated"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags {
		Name = "tf-neptunesubnet-group-test"
	}
}`, rName)
}

const testAccNeptuneSubnetGroupConfig_namePrefix = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "terraform-testacc-neptune-subnet-group-name-prefix"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags {
		Name = "tf-acc-neptune-subnet-group-name-prefix-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags {
		Name = "tf-acc-neptune-subnet-group-name-prefix-b"
	}
}

resource "aws_neptune_subnet_group" "test" {
	name_prefix = "tf_test-"
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`

const testAccNeptuneSubnetGroupConfig_generatedName = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "terraform-testacc-neptune-subnet-group-generated-name"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags {
		Name = "tf-acc-neptune-subnet-group-generated-name-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags {
		Name = "tf-acc-neptune-subnet-group-generated-name-a"
	}
}

resource "aws_neptune_subnet_group" "test" {
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`

const testAccNeptuneSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces = `
resource "aws_vpc" "main" {
    cidr_block = "192.168.0.0/16"
		tags {
			Name = "terraform-testacc-neptune-subnet-group-w-underscores-etc"
		}
}

resource "aws_subnet" "frontend" {
    vpc_id = "${aws_vpc.main.id}"
    availability_zone = "us-west-2b"
    cidr_block = "192.168.1.0/24"
    tags {
        Name = "tf-acc-neptune-subnet-group-w-underscores-etc-front"
    }
}

resource "aws_subnet" "backend" {
    vpc_id = "${aws_vpc.main.id}"
    availability_zone = "us-west-2c"
    cidr_block = "192.168.2.0/24"
    tags {
        Name = "tf-acc-neptune-subnet-group-w-underscores-etc-back"
    }
}

resource "aws_neptune_subnet_group" "underscores" {
    name = "with_underscores"
    description = "Our main group of subnets"
    subnet_ids = ["${aws_subnet.frontend.id}", "${aws_subnet.backend.id}"]
}

resource "aws_neptune_subnet_group" "periods" {
    name = "with.periods"
    description = "Our main group of subnets"
    subnet_ids = ["${aws_subnet.frontend.id}", "${aws_subnet.backend.id}"]
}

resource "aws_neptune_subnet_group" "spaces" {
    name = "with spaces"
    description = "Our main group of subnets"
    subnet_ids = ["${aws_subnet.frontend.id}", "${aws_subnet.backend.id}"]
}
`
