package kafkaconnect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKafkaConnectWorkerConfigurationDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	propertiesFileContent := "key.converter=hello\nvalue.converter=world"

	resourceName := "aws_mskconnect_worker_configuration.test"
	dataSourceName := "data.aws_mskconnect_worker_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy: nil,
		Providers:    acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkerConfigurationDataSourceConfigBasic(rName, propertiesFileContent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_revision", dataSourceName, "latest_revision"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "properties_file_content", dataSourceName, "properties_file_content"),
				),
			},
		},
	})
}

func testAccWorkerConfigurationDataSourceConfigBasic(name, content string) string {
	return fmt.Sprintf(`
resource "aws_mskconnect_worker_configuration" "test" {
  name                    = %[1]q
  properties_file_content = %[2]q
}

data "aws_mskconnect_worker_configuration" "test" {
  name = aws_mskconnect_worker_configuration.test.name
}
`, name, content)
}
