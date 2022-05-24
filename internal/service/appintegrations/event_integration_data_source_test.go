package appintegrations_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/appintegrationsservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEventIntegrationDataSource_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_appintegrations_event_integration.test"
	datasourceName := "data.aws_appintegrations_event_integration.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	sourceName := os.Getenv(key)
	if sourceName == "" {
		sourceName = "aws.partner/examplepartner.com"
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(appintegrationsservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, appintegrationsservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEventIntegrationDataSourceConfig_Name(rName, sourceName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "eventbridge_bus", resourceName, "eventbridge_bus"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "event_filter.#", resourceName, "event_filter.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "event_filter.0.source", resourceName, "event_filter.0.source"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1"),
				),
			},
		},
	})
}

func testAccEventIntegrationBaseDataSourceConfig(rName, sourceName string) string {
	return fmt.Sprintf(`
resource "aws_appintegrations_event_integration" "test" {
  name            = %[1]q
  description     = "example description"
  eventbridge_bus = "default"

  event_filter {
    source = %[2]q
  }

  tags = {
    "Key1" = "Value1"
  }
}
`, rName, sourceName)
}

func testAccEventIntegrationDataSourceConfig_Name(rName, sourceName string) string {
	return acctest.ConfigCompose(
		testAccEventIntegrationBaseDataSourceConfig(rName, sourceName),
		`
data "aws_appintegrations_event_integration" "test" {
  name = aws_appintegrations_event_integration.test.name
}
`)
}
