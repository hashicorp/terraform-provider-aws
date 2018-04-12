package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCloudwatchLogGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudwatch_log_group.test", "name", "Test"),
					resource.TestCheckResourceAttrSet("data.aws_cloudwatch_log_group.test", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_cloudwatch_log_group.test", "creation_time"),
				),
			},
		},
	})
	return
}

var testAccCheckAWSCloudwatchLogGroupDataSourceConfig = fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "Test"
}

data aws_cloudwatch_log_group "test" {
  name = "${aws_cloudwatch_log_group.test.name}"
}
`)
