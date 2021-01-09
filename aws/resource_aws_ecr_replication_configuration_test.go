package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcrReplicationConfiguration_basic(t *testing.T) {
	resourceName := "aws_ecr_replication_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrReplicationConfiguration(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrReplicationConfigurationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrReplicationConfigurationUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrReplicationConfigurationExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckAWSEcrReplicationConfigurationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSEcrReplicationConfiguration() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_regions" "test" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = tolist(data.aws_regions.test.names)[0]
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
`)
}

func testAccAWSEcrReplicationConfigurationUpdated() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_regions" "test" {}

resource "aws_ecr_replication_configuration" "test" {
  replication_configuration {
    rule {
      destination {
        region      = tolist(data.aws_regions.test.names)[1]
        registry_id = data.aws_caller_identity.current.account_id
      }
    }
  }
}
`)
}
