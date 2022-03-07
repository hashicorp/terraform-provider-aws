package kafkaconnect_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKafkaConnectConnectorDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connect.test"
	dataSourceName := "data.aws_mskconnect_connect.test"

	propertiesFileContent := "key.converter=hello\nvalue.converter=world"
	bootstrapServers := fmt.Sprintf("%s:9094,%s:9094", acctest.RandomDomainName(), acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorDataSourceConfig(rName, bootstrapServers, propertiesFileContent),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "state", dataSourceName, "state"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceName, "version"),
				),
			},
		},
	})
}

func testAccConnectorDataSourceConfig(rName string, bootstrapServers string, workerConfigurationPropertiesFileContent string) string {
	return acctest.ConfigCompose(testAccConnectorConfigBasic(rName, bootstrapServers, workerConfigurationPropertiesFileContent), `
data "aws_mskconnect_connector" "test" {
  name = aws_mskconnect_connector.test.name
}
`)
}
