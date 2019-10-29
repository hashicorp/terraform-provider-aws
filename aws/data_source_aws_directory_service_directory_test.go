package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsDirectoryServiceDefinition(t *testing.T) {
	rInt := acctest.RandInt()
	simpleDirectoryResourceName := "aws_directory_service_directory.simple"
	microsoftDirectoryResourceName := "aws_directory_service_directory.microsoft"
	simpleDirectoryName := "simple.corp.test"
	microsoftDirectoryName := "microsoft.corp.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsDirectoryServiceDefinitionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsDirectoryServiceDefinitionCheck(
						simpleDirectoryResourceName, "data.aws_directory_service_directory.simple_by_name", simpleDirectoryName),
					testAccDataSourceAwsDirectoryServiceDefinitionCheck(
						microsoftDirectoryResourceName, "data.aws_directory_service_directory.microsoft_by_name", microsoftDirectoryName),
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

	resource "aws_directory_service_directory" "simple" {
		name       = "simple.corp.test"
		short_name = "SIMPLE"
		password   = "SuperSecretPassw0rd"
		size       = "Small"

		vpc_settings {
			vpc_id     = "${aws_vpc.test.id}"
			subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
		}

		tags = {
			Environment = "dev-%d"
		}
	}

	resource "aws_directory_service_directory" "microsoft" {
		name       = "microsoft.corp.test"
		short_name = "MICROSOFT"
		password   = "SuperSecretPassw0rd"
		type       = "MicrosoftAD"

		vpc_settings {
			vpc_id     = "${aws_vpc.test.id}"
			subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
		}

		tags = {
			Environment = "dev-%d"
		}
	}

	data "aws_directory_service_directory" "simple_by_name" {
		name = "${aws_directory_service_directory.simple.name}"
	}
	data "aws_directory_service_directory" "microsoft_by_name" {
		name = "${aws_directory_service_directory.microsoft.name}"
	}
`, rInt, rInt, rInt, rInt)
}
