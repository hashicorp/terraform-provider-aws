package kafkaconnect_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKafkaConnectConnectorDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_mskconnect_connector.test"
	dataSourceName := "data.aws_mskconnect_connector.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(kafkaconnect.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, kafkaconnect.EndpointsID),
		CheckDestroy:      nil,
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccConnectorDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceName, "version"),
				),
			},
		},
	})
}

func testAccConnectorDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccConnectorConfig_basic(rName), `
data "aws_mskconnect_connector" "test" {
  name = aws_mskconnect_connector.test.name
}
`)
}
