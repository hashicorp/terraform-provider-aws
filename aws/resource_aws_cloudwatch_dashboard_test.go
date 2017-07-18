package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudWatchDashboard_basic(t *testing.T) {
	var dashboard cloudwatch.GetDashboardOutput
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchDashboardConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchDashboardExists("aws_cloudwatch_dashboard.foobar", &dashboard),
				),
			},
		},
	})
}

func testAccCheckCloudWatchDashboardExists(n string, dashboard *cloudwatch.GetDashboardOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn
		params := cloudwatch.GetDashboardInput{
			DashboardName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetDashboard(&params)
		if err != nil {
			return err
		}
		if len(*resp.DashboardName) == 0 {
			return fmt.Errorf("Dashboard not found")
		}
		*dashboard = *resp

		return nil
	}
}

func testAccCheckAWSCloudWatchDashboardDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_dashboard" {
			continue
		}

		params := cloudwatch.GetDashboardInput{
			DashboardName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetDashboard(&params)
		if err == nil {
			if len(*resp.DashboardArn) != 0 &&
				*resp.DashboardName == rs.Primary.ID {
				return fmt.Errorf("Dashboard still exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAWSCloudWatchDashboardConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_dashboard" "foobar" {
  dashboard_name = "terraform-test-dashboard-%d"
  dashboard_body = <<EOF
  {
    "widgets": [{
      "type": "text",
      "x": 0,
      "y": 0,
      "width": 6,
      "height": 6,
      "properties": {
        "markdown": "Hi there from CloudFormation"
      }
    }]
  }
  EOF
}`, rInt)
}
