package events_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEventsRuleDataSource_basic(t *testing.T) {

	ruleName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	busName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_cloudwatch_event_rule.test"
	dataSourceName := "data.aws_cloudwatch_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleDataSourceConfig_basic(ruleName, busName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
				),
			},
		},
	})
}

func testAccRuleDataSourceConfig_basic(ruleName, busName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_event_rule" "test" {
  name = %[2]q
  event_pattern = <<EOF
  {
	"detail-type": [
	  "AWS Console Sign In via CloudTrail"
	]
  }
EOF
  event_bus_name = resource.aws_cloudwatch_event_bus.test.name
}

data "aws_cloudwatch_event_rule" "test" {
  name = %[2]q
  event_bus_name = resource.aws_cloudwatch_event_bus.test.name

  depends_on = [aws_cloudwatch_event_rule.test]
}
`, busName, ruleName)
}
