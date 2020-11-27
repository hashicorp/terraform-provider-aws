package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCredentials_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCredentialsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCredentials("data.aws_credentials.current"),
				),
			},
		},
	})
}

func testAccCheckAwsCredentials(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find credentials resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("credentials resource ID not set.")
		}

		expected, err := testAccProvider.Meta().(*AWSClient).credentials.Get()
		if err != nil {
			return fmt.Errorf("Error getting test Provider Credentials: %v", err)
		}

		if rs.Primary.Attributes["access_key"] != expected.AccessKeyID {
			return fmt.Errorf("Incorrect access_key: expected %q, got %q", expected.AccessKeyID, rs.Primary.Attributes["access_key"])
		}

		if rs.Primary.Attributes["secret_key"] != expected.SecretAccessKey {
			return fmt.Errorf("Incorrect secret_key: expected %q, got %q", expected.SecretAccessKey, rs.Primary.Attributes["secret_key"])
		}

		if rs.Primary.Attributes["token"] != expected.SessionToken {
			return fmt.Errorf("Incorrect token: expected %q, got %q", expected.SessionToken, rs.Primary.Attributes["token"])
		}

		return nil
	}
}

const testAccCheckAwsCredentialsConfig_basic = `
data "aws_credentials" "current" {}
`
