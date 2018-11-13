package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsEip_Filter(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipConfigFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEip_Id(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipConfigId,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEip_PublicIP_EC2Classic(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	// Do not parallelize this test until the provider testing framework
	// has a stable us-east-1 alias
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipConfigPublicIpEc2Classic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEip_PublicIP_VPC(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipConfigPublicIpVpc,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEip_Tags(t *testing.T) {
	dataSourceName := "data.aws_eip.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ip", resourceName, "public_ip"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEipConfigFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags {
    Name = %q
  }
}

data "aws_eip" "test" {
  filter {
    name   = "tag:Name"
    values = ["${aws_eip.test.tags.Name}"]
  }
}
`, rName)
}

const testAccDataSourceAwsEipConfigId = `
resource "aws_eip" "test" {
  vpc = true
}

data "aws_eip" "test" {
  id = "${aws_eip.test.id}"
}
`

const testAccDataSourceAwsEipConfigPublicIpEc2Classic = `
provider "aws" {
  region = "us-east-1"
}

resource "aws_eip" "test" {}

data "aws_eip" "test" {
  public_ip = "${aws_eip.test.public_ip}"
}
`

const testAccDataSourceAwsEipConfigPublicIpVpc = `
resource "aws_eip" "test" {
  vpc = true
}

data "aws_eip" "test" {
  public_ip = "${aws_eip.test.public_ip}"
}
`

func testAccDataSourceAwsEipConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags {
    Name = %q
  }
}

data "aws_eip" "test" {
  tags {
    Name = "${aws_eip.test.tags["Name"]}"
  }
}
`, rName)
}
