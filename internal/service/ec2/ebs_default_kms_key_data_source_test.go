package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2EBSDefaultKMSKeyDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSDefaultKMSKeyDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEBSDefaultKMSKey("data.aws_ebs_default_kms_key.current"),
				),
			},
		},
	})
}

func testAccCheckEBSDefaultKMSKey(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		actual, err := conn.GetEbsDefaultKmsKeyId(&ec2.GetEbsDefaultKmsKeyIdInput{})
		if err != nil {
			return fmt.Errorf("Error reading EBS default KMS key: %q", err)
		}

		attr := rs.Primary.Attributes["key_arn"]

		if attr != aws.StringValue(actual.KmsKeyId) {
			return fmt.Errorf("EBS default KMS key is not the expected value (%s)", aws.StringValue(actual.KmsKeyId))
		}

		return nil
	}
}

const testAccEBSDefaultKMSKeyDataSourceConfig = `
data "aws_ebs_default_kms_key" "current" {}
`
