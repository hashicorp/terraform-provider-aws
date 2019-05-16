package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsDirectoryServiceDefinition(t *testing.T) {
	rInt := acctest.RandInt()
	directoryResourceName := "aws_directory_service_directory.test"
	directoryName := "corp.test"
	// directoryPass := "corp-pass"
	// directorySize := "small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsDirectoryServiceDefinitionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsDirectoryServiceDefinitionCheck(
						directoryResourceName, "data.aws_ds_directory.by_name", directoryName),
				),
			},
		},
	})
}

// rsName for the name of the created resource
// dsName for the name of the created data source
// zName for the name of the domain
func testAccDataSourceAwsDirectoryServiceDefinitionCheck(rsName, dsName, zName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rsName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", rsName)
		}

		if !ok {
			return fmt.Errorf("can't find directory %q in state", dsName)
		}

		attr := rs.Primary.Attributes

		if attr["name"] != zName {
			return fmt.Errorf("Directory Service definition name is %q; want %q", attr["name"], zName)
		}

		return nil
	}
}

func testAccDataSourceAwsDirectoryServiceDefinitionConfig(rInt int) string {
	return fmt.Sprintf(`
	provider "aws" {
		region = "us-east-1"
	}

	resource "aws_vpc" "test" {
		cidr_block = "172.16.0.0/16"
	tags = {
			Name = "terraform-testacc-ds-definition-data-source"
		}
	}

	resource "aws_subnet" "foo" {
		vpc_id            = "${aws_vpc.test.id}"
		availability_zone = "us-east-1a"
		cidr_block        = "172.16.50.0/24"

		tags = {
			Environment = "dev-%d"
		}
	}

	resource "aws_subnet" "bar" {
		vpc_id            = "${aws_vpc.test.id}"
		availability_zone = "us-east-1b"
		cidr_block        = "172.16.51.0/24"

		tags = {
			Environment = "dev-%d"
		}
	}

	resource "aws_directory_service_directory" "test" {
		name     = "corp.test"
		password = "SuperSecretPassw0rd"
		size     = "Small"

		vpc_settings {
			vpc_id     = "${aws_vpc.test.id}"
			subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
		}

		tags = {
			Environment = "dev-%d"
		}
	}

	data "aws_ds_directory" "by_name" {
		name = "${aws_directory_service_directory.test.name}"
	}`, rInt, rInt, rInt)
}
