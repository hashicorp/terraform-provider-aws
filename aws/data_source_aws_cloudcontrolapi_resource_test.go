package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudcontrolapi"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAwsCloudControlApiResourceDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_cloudcontrolapi_resource.test"
	resourceName := "aws_cloudcontrolapi_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudcontrolapi.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsCloudControlApiResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudControlApiResourceDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "properties", resourceName, "properties"),
					resource.TestCheckResourceAttrPair(dataSourceName, "type_name", resourceName, "type_name"),
				),
			},
		},
	})
}

func testAccAwsCloudControlApiResourceDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudcontrolapi_resource" "test" {
  type_name = "AWS::Logs::LogGroup"

  desired_state = jsonencode({
    LogGroupName = %[1]q
  })
}

data "aws_cloudcontrolapi_resource" "test" {
  identifier = aws_cloudcontrolapi_resource.test.id
  type_name  = aws_cloudcontrolapi_resource.test.type_name
}
`, rName)
}
