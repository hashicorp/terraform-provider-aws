package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSSQSQueuePolicy_basic(t *testing.T) {
	var queueAttributes map[string]*string
	resourceName := "aws_sqs_queue_policy.test"
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSPolicyConfigBasic(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists("aws_sqs_queue.test", &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					resource.TestMatchResourceAttr("aws_sqs_queue_policy.test", "policy",
						regexp.MustCompile("^{\"Version\":\"2012-10-17\".+")),
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

func TestAccAWSSQSQueuePolicy_disappears_queue(t *testing.T) {
	var queueAttributes map[string]*string
	resourceName := "aws_sqs_queue_policy.test"
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSPolicyConfigBasic(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists("aws_sqs_queue.test", &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSqsQueue(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSQSQueuePolicy_disappears(t *testing.T) {
	var queueAttributes map[string]*string
	resourceName := "aws_sqs_queue_policy.test"
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSPolicyConfigBasic(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists("aws_sqs_queue.test", &queueAttributes),
					testAccCheckAWSSQSQueueDefaultAttributes(&queueAttributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSqsQueuePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSSQSPolicyConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = "${aws_sqs_queue.test.id}"
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"sqspolicy",
  "Statement":[
    {
      "Sid":"First",
      "Effect":"Allow",
      "Principal":"*",
      "Action":"sqs:SendMessage",
      "Resource":"${aws_sqs_queue.test.arn}",
      "Condition":{
        "ArnEquals":{
          "aws:SourceArn":"${aws_sqs_queue.test.arn}"
        }
      }
    }
  ]
}
POLICY
}
`, rName)
}
