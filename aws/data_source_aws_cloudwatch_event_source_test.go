package aws

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchevents"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsCloudWatchEventSource(t *testing.T) {
	//resourceName := "aws_cloudwatch_event_bus.test"
	dataSourceName := "data.aws_cloudwatch_event_source.test"

	key := "EVENT_BRIDGE_PARTNER_EVENT_BUS_NAME"
	busName := os.Getenv(key)
	if busName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}
	parts := strings.Split(busName, "/")
	if len(parts) < 2 {
		t.Errorf("unable to parse partner event bus name %s", busName)
	}
	namePrefix := parts[0] + "/" + parts[1]
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, cloudwatchevents.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourcePartnerEventSourceConfig(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", busName),
					resource.TestCheckResourceAttr(dataSourceName, "created_by", namePrefix),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
				),
			},
		},
	})
}

func testAccAwsDataSourcePartnerEventSourceConfig(namePrefix string) string {
	return fmt.Sprintf(`
data "aws_cloudwatch_event_source" "test" {
  name_prefix = "%s"
}
`, namePrefix)
}
