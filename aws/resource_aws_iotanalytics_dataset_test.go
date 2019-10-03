package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIoTAnalyticsDataset_basic(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iotanalytics_dataset.dataset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTAnalyticsDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTAnalyticsDataset_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIoTAnalyticsDatasetExists_basic("aws_iotanalytics_dataset.dataset"),
					resource.TestCheckResourceAttr("aws_iotanalytics_dataset.dataset", "name", fmt.Sprintf("test_dataset_%s", rString)),
					testAccCheckAWSIoTAnalyticsDataset_basic(rString),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSIoTAnalyticsDataset_basic(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iotanalytics_dataset" {
				continue
			}

			params := &iotanalytics.DescribeDatasetInput{
				DatasetName: aws.String(rs.Primary.ID),
			}
			out, err := conn.DescribeDataset(params)

			if err != nil {
				return err
			}

			dataset := out.Dataset

			action := dataset.Actions[0]
			expectedActionName := "test_action"

			if *action.ActionName != expectedActionName {
				return fmt.Errorf("Expected action.ActionName %s is not equal to %s", expectedActionName, *action.ActionName)
			}

			if action.QueryAction == nil {
				return fmt.Errorf("Expected action.QueryAction is not nil")
			}

			if action.ContainerAction != nil {
				return fmt.Errorf("Expected action.ContainerAction is nil")
			}

			queryAction := action.QueryAction
			expectedSQLQuery := fmt.Sprintf("select * from test_datastore_%s", rString)

			if *queryAction.SqlQuery != expectedSQLQuery {
				return fmt.Errorf("Expected queryAction.SqlQuery %s is not equal to %s", expectedSQLQuery, *queryAction.SqlQuery)
			}

			filters := queryAction.Filters
			if len(filters) != 1 {
				return fmt.Errorf("Expected queryAction.Filters len %d is not equal to %d", 1, len(filters))
			}

			queryFilter := filters[0]
			expectedOffset := int64(30)
			if *queryFilter.DeltaTime.OffsetSeconds != expectedOffset {
				return fmt.Errorf("Expected queryFilter.DeltaTime.OffsetSeconds %d is not equal to %d", expectedOffset, *queryFilter.DeltaTime.OffsetSeconds)
			}

			expectedTimeExpression := "date"
			if *queryFilter.DeltaTime.TimeExpression != expectedTimeExpression {
				return fmt.Errorf("Expected queryFilter.DeltaTime.TimeExpression %s is not equal to %s", expectedTimeExpression, *queryFilter.DeltaTime.TimeExpression)
			}
		}
		return nil
	}
}

func testAccCheckAWSIoTAnalyticsDatasetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_dataset.dataset" {
			continue
		}

		params := &iotanalytics.DescribeDatasetInput{
			DatasetName: aws.String(rs.Primary.ID),
		}
		_, err := conn.DescribeDataset(params)

		if err != nil {
			if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected IoTAnalytics Dataset to be destroyed, %s found", rs.Primary.ID)

	}

	return nil
}

func testAccCheckAWSIoTAnalyticsDatasetExists_basic(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccAWSIoTAnalyticsDatasetRole = `
resource "aws_iam_role" "iotanalytics_role" {
    name = "test_role_%[1]s"
    assume_role_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "iotanalytics.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
    }]
}
EOF
}

resource "aws_iam_policy" "policy" {
    name = "test_policy_%[1]s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
    }]
}
EOF
}

resource "aws_iam_policy_attachment" "attach_policy" {
    name = "test_policy_attachment_%[1]s"
    roles = ["${aws_iam_role.iotanalytics_role.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
}

resource "aws_iotanalytics_datastore" "datastore" {
	name = "test_datastore_%[1]s"
  
	storage {
		service_managed_s3 {}
	}
  
	retention_period {
		unlimited = true
	}
  }
`

func testAccAWSIoTAnalyticsDataset_basic(rString string) string {
	return fmt.Sprintf(testAccAWSIoTAnalyticsDatasetRole+`
resource "aws_iotanalytics_dataset" "dataset" {
  name = "test_dataset_%[1]s"

  action {
	  name = "test_action"

	  query_action {

		filter {
			delta_time {
				offset_seconds = 30
				time_expression = "date"
			}
		}

		  sql_query = "select * from ${aws_iotanalytics_datastore.datastore.name}"
	  }
  }

}
`, rString)
}
