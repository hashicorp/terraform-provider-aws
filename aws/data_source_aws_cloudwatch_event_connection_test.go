package aws

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSDataSourceCloudwatch_Event_Connection_basic(t *testing.T) {
	dataSourceName := "data.aws_cloudwatch_event_connection.test"
	resourceName := "aws_cloudwatch_event_connection.api_key"

	name := sdkacctest.RandomWithPrefix("tf-acc-test")
	authorizationType := "API_KEY"
	description := sdkacctest.RandomWithPrefix("tf-acc-test")
	key := sdkacctest.RandomWithPrefix("tf-acc-test")
	value := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatch_Event_ConnectionDataConfig(
					name,
					description,
					authorizationType,
					key,
					value,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "secret_arn", resourceName, "secret_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorization_type", resourceName, "authorization_type"),
				),
			},
		},
	})
}

func testAccAWSCloudwatch_Event_ConnectionDataConfig(name, description, authorizationType, key, value string) string {
	return acctest.ConfigCompose(
		testAccAWSCloudWatchEventConnectionConfig_apiKey(name, description, authorizationType, key, value),
		`
data "aws_cloudwatch_event_connection" "test" {
  name = aws_cloudwatch_event_connection.api_key.name
}
`)
}
