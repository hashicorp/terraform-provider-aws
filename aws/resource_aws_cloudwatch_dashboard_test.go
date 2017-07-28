package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudWatchDashboard_basic(t *testing.T) {
	var db cloudwatch.GetDashboardOutput

	emptyBody := `{"widgets": []}`
	noRealChange := `{"widgets"  : [  ]}` //only whitespace change
	changedBody := `{
	"widgets": [{
		"type": "metric",
		"properties": {
        	"view": "singleValue",
        	"metrics": [
          	[ "AWS/DynamoDB", "ProvisionedWriteCapacityUnits", "TableName", "jobs" ]
        	],
        "region": "us-east-1"
      }}]}`

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchDashboardConfig("foodash", emptyBody),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchDashboardExists("aws_cloudwatch_dashboard.foodash", &db),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "name", "foodash"),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "body", emptyBody+"\n"),
				),
			},
			{
				Config: testAccAWSCloudWatchDashboardConfig("foodash", noRealChange),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchDashboardExists("aws_cloudwatch_dashboard.foodash", &db),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "name", "foodash"),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "body", emptyBody+"\n"),
				),
			},
			{
				Config: testAccAWSCloudWatchDashboardConfig("foodash-2", emptyBody),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchDashboardExists("aws_cloudwatch_dashboard.foodash", &db),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "name", "foodash-2"),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "body", emptyBody+"\n"),
				),
			},
			{
				Config: testAccAWSCloudWatchDashboardConfig("foodash-2", changedBody),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudWatchDashboardExists("aws_cloudwatch_dashboard.foodash", &db),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "name", "foodash-2"),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foodash", "body", changedBody+"\n"),
				),
			},
		},
	})
}

func testAccAWSCloudWatchDashboardConfig(name, body string) string {
	return fmt.Sprintf(
		`resource "aws_cloudwatch_dashboard" "foodash" {
					name = "%s",
					body = <<EOF
%s
EOF
}`, name, body)
}

func testAccCheckAWSCloudWatchDashboardExists(n string, dashboard *cloudwatch.GetDashboardOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found in tfstate: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn
		db, err := lookupCloudWatchDashboard(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*dashboard = *db

		return nil
	}
}

func testAccCheckAWSCloudWatchDashboardDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_dashboard" {
			continue
		}

		_, err := lookupCloudWatchDashboard(conn, rs.Primary.ID)
		if cloudwatch.ErrCodeDashboardNotFoundError != err.(awserr.Error).Code() {
			return fmt.Errorf("Unexpected error or dashboard %s still exists", rs.Primary.ID)
		}

	}

	return nil
}

func lookupCloudWatchDashboard(conn *cloudwatch.CloudWatch, name string) (*cloudwatch.GetDashboardOutput, error) {
	input := cloudwatch.GetDashboardInput{
		DashboardName: aws.String(name),
	}

	return conn.GetDashboard(&input)
}
