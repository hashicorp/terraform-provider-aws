package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSProviderCredentials_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsProviderCredentialsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsProviderCredentialsAccountId("data.aws_provider_credentials.current"),
				),
			},
		},
	})
}

// Protects against a panic in the AWS Provider configuration.
// See https://github.com/terraform-providers/terraform-provider-aws/pull/1227
func TestAccAWSProviderCredentials_basic_panic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsProviderCredentialsConfig_basic_panic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsProviderCredentialsAccountId("data.aws_provider_credentials.current"),
				),
			},
		},
	})
}

func testAccCheckAwsProviderCredentialsAccountId(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AccountID resource: %s", n)
		}

		if rs.Primary.Attributes["access_key"] == "" {
			return fmt.Errorf("Access Key expected to not be nil")
		}

		if rs.Primary.Attributes["secret_key"] == "" {
			return fmt.Errorf("Secret Key expected to not be nil")
		}

		return nil
	}
}

const testAccCheckAwsProviderCredentialsConfig_basic = `
data "aws_provider_credentials" "current" { }
`

const testAccCheckAwsProviderCredentialsConfig_basic_panic = `
provider "aws" {
  assume_role {
  }
}

data "aws_provider_credentials" "current" {}
`
