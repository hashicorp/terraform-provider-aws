package ec2_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2EBSEncryptionByDefaultDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSEncryptionByDefaultDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSEncryptionByDefaultDataSource("data.aws_ebs_encryption_by_default.current"),
				),
			},
		},
	})
}

func testAccCheckEBSEncryptionByDefaultDataSource(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

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

const testAccEBSEncryptionByDefaultDataSourceConfig = `
data "aws_ebs_encryption_by_default" "current" {}
`
