package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsKmsKey(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsKmsKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsKmsKeyCheck("data.aws_kms_key.arbitrary"),
				),
			},
		},
	})
}

func testAccDataSourceAwsKmsKeyCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		kmsKeyRs, ok := s.RootModule().Resources["aws_kms_key.arbitrary"]
		if !ok {
			return fmt.Errorf("can't find aws_kms_key.arbitrary in state")
		}

		attr := rs.Primary.Attributes

		checkProperties := []string{"arn", "key_usage", "description"}

		for _, p := range checkProperties {
			if attr[p] != kmsKeyRs.Primary.Attributes[p] {
				return fmt.Errorf(
					"%s is %s; want %s",
					p,
					attr[p],
					kmsKeyRs.Primary.Attributes[p],
				)
			}
		}

		return nil
	}
}

const testAccDataSourceAwsKmsKeyConfig = `
resource "aws_kms_key" "arbitrary" {
    description = "Terraform acc test"
    deletion_window_in_days = 7
}

data "aws_kms_key" "arbitrary" {
  key_id = "${aws_kms_key.arbitrary.key_id}"
}`
