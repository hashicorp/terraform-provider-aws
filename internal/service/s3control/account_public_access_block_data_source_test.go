package s3control_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccS3ControlAccountPublicAccessBlockDataSource_basic(t *testing.T) {
	var conf s3control.PublicAccessBlockConfiguration
	resourceName := "aws_s3_account_public_access_block.test"
	dataSourceName := "data.aws_s3_account_public_access_block.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, s3control.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccS3ControlAccountPublicAccessBlockDataSource_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountPublicAccessBlockExistsDataSource(resourceName, &conf),
					testAccCheckAccountPublicAccessBlockMatch(dataSourceName, "block_public_acls", resourceName),
					testAccCheckAccountPublicAccessBlockMatch(dataSourceName, "block_public_policy", resourceName),
					testAccCheckAccountPublicAccessBlockMatch(dataSourceName, "ignore_public_acls", resourceName),
					testAccCheckAccountPublicAccessBlockMatch(dataSourceName, "restrict_public_buckets", resourceName),
				),
			},
		},
	})
}

func testAccCheckAccountPublicAccessBlockMatch(datasource, attribute, resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[datasource]
		if !ok {
			return fmt.Errorf("not found: %s", datasource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		attrDataSourceValue, ok := rs.Primary.Attributes[attribute]
		if !ok {
			return fmt.Errorf("attribute %q not found for %q", attribute, datasource)
		}

		rs, ok = s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("mo ID is set")
		}
		attrResourceValue, ok := rs.Primary.Attributes[attribute]
		if !ok {
			return fmt.Errorf("attribute %q not found for %q", attribute, resource)
		}
		if attrDataSourceValue != attrResourceValue {
			return fmt.Errorf("Account public access block policies differ for %s attribute.\aDataSourceValue: %s\aResourceValue: %s", attribute, attrDataSourceValue, attrResourceValue)
		}

		return nil
	}
}

func testAccCheckAccountPublicAccessBlockExistsDataSource(n string, configuration *s3control.PublicAccessBlockConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no account public access block is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlConn

		input := &s3control.GetPublicAccessBlockInput{
			AccountId: aws.String(rs.Primary.ID),
		}
		output, err := conn.GetPublicAccessBlock(input)
		if err != nil {
			return err
		}

		if output == nil || output.PublicAccessBlockConfiguration == nil {
			return fmt.Errorf("S3 Account Public Access Block not found")
		}

		*configuration = *output.PublicAccessBlockConfiguration

		return nil
	}
}

func testAccDataSourceAccountPublicAccessBlockBaseConfig() string {
	return fmt.Sprintf(`
resource "aws_s3_account_public_access_block" "test" {
  block_public_acls = true
  block_public_policy = true
  ignore_public_acls = true
  restrict_public_buckets = true
}
`)
}

func testAccS3ControlAccountPublicAccessBlockDataSource_basic() string {
	return acctest.ConfigCompose(testAccDataSourceAccountPublicAccessBlockBaseConfig(), `
data "aws_s3_account_public_access_block" "test" {
  depends_on = [aws_s3_account_public_access_block.test]
}
`)
}
