package sqs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsqs "github.com/hashicorp/terraform-provider-aws/internal/service/sqs"
)

func TestAccSQSQueuePolicy_basic(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sqs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccQueuePolicyConfig(rName),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "policy", queueResourceName, "policy"),
				),
			},
		},
	})
}

func TestAccSQSQueuePolicy_disappears(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sqs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(acctest.Provider, tfsqs.ResourceQueuePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSQSQueuePolicy_Disappears_queue(t *testing.T) {
	var queueAttributes map[string]string
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sqs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(queueResourceName, &queueAttributes),
					acctest.CheckResourceDisappears(acctest.Provider, tfsqs.ResourceQueue(), queueResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSQSQueuePolicy_update(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sqs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueuePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueueExists(queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func testAccQueuePolicyConfig(rName string) string {
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

func testAccPolicyUpdatedConfig(rName string) string {
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
