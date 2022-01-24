package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccDataSourceAwsEips_Filter(t *testing.T) {
	var conf ec2.Address

	dataSourceName := "data.aws_eips.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipsConfigFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &conf),
					testAccCheckAWSEIPExists("aws_eip.test2", &conf),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "public_ips.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEipsClassic_Filter(t *testing.T) {
	var conf ec2.Address

	dataSourceName := "data.aws_eips.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipsClassicConfigFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &conf),
					testAccCheckAWSEIPExists("aws_eip.test2", &conf),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "public_ips.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEips_Tags(t *testing.T) {
	var conf ec2.Address

	dataSourceName := "data.aws_eips.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipsConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &conf),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ips.0", resourceName, "public_ip"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "public_ips.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEipsClassic_Tags(t *testing.T) {
	var conf ec2.Address

	dataSourceName := "data.aws_eips.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipsClassicConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &conf),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", "aws_eips.test", "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ips.0", "aws_eips.test", "public_ip"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "public_ips.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEipsConfigFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %q
  }
}
resource "aws_eip" "test2" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eips" "test" {
  filter {
    name   = "tag:Name"
    values = ["${aws_eip.test.tags.Name}"]
  }
}
`, rName, rName)
}

func testAccDataSourceAwsEipsConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eips" "test" {
  tags = {
    Name = "${aws_eip.test.tags["Name"]}"
  }
}
`, rName)
}

func testAccDataSourceAwsEipsClassicConfigFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = false

  tags = {
    Name = %q
  }
}
resource "aws_eip" "test2" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eips" "test" {
  filter {
    name   = "tag:Name"
    values = ["${aws_eip.test.tags.Name}"]
  }
}
`, rName, rName)
}

func testAccDataSourceAwsEipsClassicConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = false

  tags = {
    Name = %q
  }
}

data "aws_eips" "test" {
  tags = {
    Name = "${aws_eip.test.tags["Name"]}"
  }
}
`, rName)
}
