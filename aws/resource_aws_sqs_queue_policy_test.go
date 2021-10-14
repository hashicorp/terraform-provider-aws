package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSSQSQueuePolicy_basic(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sqs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccAWSSQSQueuePolicyConfig(rName),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "policy", queueResourceName, "policy"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueuePolicy_disappears(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sqs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceQueuePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSQSQueuePolicy_disappears_queue(t *testing.T) {
	var queueAttributes map[string]string
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sqs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceQueue(), queueResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSQSQueuePolicy_Update(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sqs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSQSPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func testAccAWSSQSQueuePolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": "sqs:*",
    "Resource": "${aws_sqs_queue.test.arn}",
    "Condition": {
      "ArnEquals": {
        "aws:SourceArn": "${aws_sqs_queue.test.arn}"
      }
    }
  }]
}
POLICY
}
`, rName)
}

func testAccAWSSQSPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": [
      "sqs:SendMessage",
      "sqs:ReceiveMessage"
    ],
    "Resource": "${aws_sqs_queue.test.arn}"
  }]
}
POLICY
}
`, rName)
}
