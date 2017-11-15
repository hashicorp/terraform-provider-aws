package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsSqsQueue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceAwsSqsQueueConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsSqsQueueCheck("data.aws_sqs_queue.by_name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsSqsQueueCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		sqsQueueRs, ok := s.RootModule().Resources["aws_sqs_queue.tf_test"]
		if !ok {
			return fmt.Errorf("can't find aws_sqs_queue.tf_test in state")
		}

		attr := rs.Primary.Attributes

		if attr["name"] != sqsQueueRs.Primary.Attributes["name"] {
			return fmt.Errorf(
				"name is %s; want %s",
				attr["name"],
				sqsQueueRs.Primary.Attributes["name"],
			)
		}

		return nil
	}
}

const testAccDataSourceAwsSqsQueueConfig = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_sqs_queue" "tf_wrong1" {
  name = "wrong1"
}
resource "aws_sqs_queue" "tf_test" {
  name = "tf_test"
}
resource "aws_sqs_queue" "tf_wrong2" {
  name = "wrong2"
}

data "aws_sqs_queue" "by_name" {
  name = "${aws_sqs_queue.tf_test.name}"
}
`
