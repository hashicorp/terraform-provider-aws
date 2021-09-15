package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsCloudformationResourceDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_cloudformation_resource.test"
	resourceName := "aws_cloudformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, cloudformation.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudformationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudformationResourceDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "resource_model", resourceName, "resource_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type_name", resourceName, "type_name"),
				),
			},
		},
	})
}

func testAccAwsCloudformationResourceDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}

data "aws_cloudformation_resource" "test" {
  identifier = aws_cloudformation_resource.test.id
  type_name  = aws_cloudformation_resource.test.type_name
}
`, rName)
}
