package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/atest"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func TestAccDataSourceAwsEBSEncryptionByDefault_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { atest.PreCheck(t) },
		ErrorCheck: atest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  atest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEBSEncryptionByDefaultConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceAwsEBSEncryptionByDefault("data.aws_ebs_encryption_by_default.current"),
				),
			},
		},
	})
}

func testAccCheckDataSourceAwsEBSEncryptionByDefault(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := atest.Provider.Meta().(*awsprovider.AWSClient).EC2Conn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		actual, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
		if err != nil {
			return fmt.Errorf("Error reading default EBS encryption toggle: %q", err)
		}

		attr, _ := strconv.ParseBool(rs.Primary.Attributes["enabled"])

		if attr != aws.BoolValue(actual.EbsEncryptionByDefault) {
			return fmt.Errorf("EBS encryption by default is not in expected state (%t)", aws.BoolValue(actual.EbsEncryptionByDefault))
		}

		return nil
	}
}

const testAccDataSourceAwsEBSEncryptionByDefaultConfig = `
data "aws_ebs_encryption_by_default" "current" {}
`
