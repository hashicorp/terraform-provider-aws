package aws

import (
	"fmt"
	"testing"
	"time"

	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/jen20/awspolicyequivalence"
)

func TestAccAWSSQSQueue_basic(t *testing.T) {
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
				),
			},
			{
				Config: testAccAWSSQSConfigWithOverrides(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithOverrides("aws_sqs_queue.queue"),
				),
			},
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_tags(t *testing.T) {
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithTags(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "tags.Usage", "original"),
				),
			},
			{
				Config: testAccAWSSQSConfigWithTagsChanged(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
					resource.TestCheckNoResourceAttr("aws_sqs_queue.queue", "tags"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_namePrefix(t *testing.T) {
	prefix := "acctest-sqs-queue"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithNamePrefix(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
					testAccCheckAWSSQSGeneratedNamePrefix("aws_sqs_queue.queue", "acctest-sqs-queue", false),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_namePrefix_fifo(t *testing.T) {
	prefix := "acctest-sqs-queue"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSFifoConfigWithNamePrefix(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
					testAccCheckAWSSQSGeneratedNamePrefix("aws_sqs_queue.queue", "acctest-sqs-queue", true),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_policy(t *testing.T) {
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	topicName := fmt.Sprintf("sns-topic-%s", acctest.RandString(10))

	expectedPolicyText := fmt.Sprintf(
		`{"Version": "2012-10-17","Id": "sqspolicy","Statement":[{"Sid": "Stmt1451501026839","Effect": "Allow","Principal":"*","Action":"sqs:SendMessage","Resource":"arn:aws:sqs:us-west-2:470663696735:%s","Condition":{"ArnEquals":{"aws:SourceArn":"arn:aws:sns:us-west-2:470663696735:%s"}}}]}`,
		topicName, queueName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfig_PolicyFormat(topicName, queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQSHasPolicy("aws_sqs_queue.test-email-events", expectedPolicyText),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_queueDeletedRecently(t *testing.T) {
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
				),
			},
			{
				Config: "# delete queue to quickly recreate",
				Check:  testAccCheckAWSSQSQueueDestroy,
			},
			{
				Config: testAccAWSSQSConfigWithDefaults(queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.queue"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_redrivePolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithRedrive(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithDefaults("aws_sqs_queue.my_dead_letter_queue"),
				),
			},
		},
	})
}

// Tests formatting and compacting of Policy, Redrive json
func TestAccAWSSQSQueue_Policybasic(t *testing.T) {
	queueName := fmt.Sprintf("sqs-queue-%s", acctest.RandString(10))
	topicName := fmt.Sprintf("sns-topic-%s", acctest.RandString(10))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfig_PolicyFormat(topicName, queueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExistsWithOverrides("aws_sqs_queue.test-email-events"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_FIFO(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithFIFO(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExists("aws_sqs_queue.queue"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "fifo_queue", "true"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_FIFOExpectNameError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSQSConfigWithFIFOExpectError(acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Error validating the FIFO queue name`),
			},
		},
	})
}

func TestAccAWSSQSQueue_FIFOWithContentBasedDeduplication(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithFIFOContentBasedDeduplication(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExists("aws_sqs_queue.queue"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "fifo_queue", "true"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "content_based_deduplication", "true"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueue_ExpectContentBasedDeduplicationError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccExpectContentBasedDeduplicationError(acctest.RandString(10)),
				ExpectError: regexp.MustCompile(`Content based deduplication can only be set with FIFO queues`),
			},
		},
	})
}

func testAccCheckAWSSQSQueueDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sqsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sqs_queue" {
			continue
		}

		// Check if queue exists by checking for its attributes
		params := &sqs.GetQueueAttributesInput{
			QueueUrl: aws.String(rs.Primary.ID),
		}
		err := resource.Retry(15*time.Second, func() *resource.RetryError {
			_, err := conn.GetQueueAttributes(params)
			if err != nil {
				if isAWSErr(err, sqs.ErrCodeQueueDoesNotExist, "") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			return resource.RetryableError(fmt.Errorf("Queue %s still exists. Failing!", rs.Primary.ID))
		})
		if err != nil {
			return err
		}
	}

	return nil
}
func testAccCheckAWSQSHasPolicy(n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).sqsconn

		params := &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(rs.Primary.ID),
			AttributeNames: []*string{aws.String("Policy")},
		}
		resp, err := conn.GetQueueAttributes(params)
		if err != nil {
			return err
		}

		var actualPolicyText string
		for k, v := range resp.Attributes {
			if k == "Policy" {
				actualPolicyText = *v
				break
			}
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckAWSSQSExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		return nil
	}
}

func testAccCheckAWSSQSExistsWithDefaults(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).sqsconn

		params := &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(rs.Primary.ID),
			AttributeNames: []*string{aws.String("All")},
		}
		resp, err := conn.GetQueueAttributes(params)

		if err != nil {
			return err
		}

		// checking if attributes are defaults
		for k, v := range resp.Attributes {
			if k == "VisibilityTimeout" && *v != "30" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 30", *v)
			}

			if k == "MessageRetentionPeriod" && *v != "345600" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 345600", *v)
			}

			if k == "MaximumMessageSize" && *v != "262144" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 262144", *v)
			}

			if k == "DelaySeconds" && *v != "0" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 0", *v)
			}

			if k == "ReceiveMessageWaitTimeSeconds" && *v != "0" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 0", *v)
			}
		}

		return nil
	}
}

func testAccCheckAWSSQSGeneratedNamePrefix(resource, prefix string, isFifo bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		if isFifo && !strings.HasSuffix(name, ".fifo") {
			return fmt.Errorf("Name: %q, does not have suffix: %q", name, ".fifo")
		}
		return nil
	}
}

func testAccCheckAWSSQSExistsWithOverrides(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Queue URL specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).sqsconn

		params := &sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(rs.Primary.ID),
			AttributeNames: []*string{aws.String("All")},
		}
		resp, err := conn.GetQueueAttributes(params)

		if err != nil {
			return err
		}

		// checking if attributes match our overrides
		for k, v := range resp.Attributes {
			if k == "VisibilityTimeout" && *v != "60" {
				return fmt.Errorf("VisibilityTimeout (%s) was not set to 60", *v)
			}

			if k == "MessageRetentionPeriod" && *v != "86400" {
				return fmt.Errorf("MessageRetentionPeriod (%s) was not set to 86400", *v)
			}

			if k == "MaximumMessageSize" && *v != "2048" {
				return fmt.Errorf("MaximumMessageSize (%s) was not set to 2048", *v)
			}

			if k == "DelaySeconds" && *v != "90" {
				return fmt.Errorf("DelaySeconds (%s) was not set to 90", *v)
			}

			if k == "ReceiveMessageWaitTimeSeconds" && *v != "10" {
				return fmt.Errorf("ReceiveMessageWaitTimeSeconds (%s) was not set to 10", *v)
			}
		}

		return nil
	}
}

func TestAccAWSSQSQueue_Encryption(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSConfigWithEncryption(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSExists("aws_sqs_queue.queue"),
					resource.TestCheckResourceAttr("aws_sqs_queue.queue", "kms_master_key_id", "alias/aws/sqs"),
				),
			},
		},
	})
}

func testAccAWSSQSConfigWithDefaults(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name = "%s"
}
`, r)
}

func testAccAWSSQSConfigWithNamePrefix(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name_prefix = "%s"
}
`, r)
}

func testAccAWSSQSFifoConfigWithNamePrefix(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name_prefix = "%s"
  fifo_queue  = true
}
`, r)
}

func testAccAWSSQSFifoConfigWithDefaults(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name = "%s.fifo"
  fifo_queue = true
}
`, r)
}

func testAccAWSSQSConfigWithOverrides(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                       = "%s"
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60
}`, r)
}

func testAccAWSSQSConfigWithRedrive(name string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "my_queue" {
  name                       = "tftestqueuq-%s"
  delay_seconds              = 0
  visibility_timeout_seconds = 300

  redrive_policy = <<EOF
{
  "maxReceiveCount": 3,
  "deadLetterTargetArn": "${aws_sqs_queue.my_dead_letter_queue.arn}"
}
EOF
}

resource "aws_sqs_queue" "my_dead_letter_queue" {
  name = "tfotherqueuq-%s"
}
`, name, name)
}

func testAccAWSSQSConfig_PolicyFormat(queue, topic string) string {
	return fmt.Sprintf(`
variable "sns_name" {
  default = "%s"
}

variable "sqs_name" {
  default = "%s"
}

resource "aws_sns_topic" "test_topic" {
  name = "${var.sns_name}"
}

resource "aws_sqs_queue" "test-email-events" {
  name                       = "${var.sqs_name}"
  depends_on                 = ["aws_sns_topic.test_topic"]
  delay_seconds              = 90
  max_message_size           = 2048
  message_retention_seconds  = 86400
  receive_wait_time_seconds  = 10
  visibility_timeout_seconds = 60

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "sqspolicy",
  "Statement": [
    {
      "Sid": "Stmt1451501026839",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sqs:SendMessage",
      "Resource": "arn:aws:sqs:us-west-2:470663696735:${var.sqs_name}",
      "Condition": {
        "ArnEquals": {
          "aws:SourceArn": "arn:aws:sns:us-west-2:470663696735:${var.sns_name}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_sns_topic_subscription" "test_queue_target" {
  topic_arn = "${aws_sns_topic.test_topic.arn}"
  protocol  = "sqs"
  endpoint  = "${aws_sqs_queue.test-email-events.arn}"
}
`, topic, queue)
}

func testAccAWSSQSConfigWithFIFO(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name       = "%s.fifo"
  fifo_queue = true
}
`, queue)
}

func testAccAWSSQSConfigWithFIFOContentBasedDeduplication(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                        = "%s.fifo"
  fifo_queue                  = true
  content_based_deduplication = true
}
`, queue)
}

func testAccAWSSQSConfigWithFIFOExpectError(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name       = "%s"
  fifo_queue = true
}
`, queue)
}

func testAccExpectContentBasedDeduplicationError(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                        = "%s"
  content_based_deduplication = true
}
`, queue)
}

func testAccAWSSQSConfigWithEncryption(queue string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name                              = "%s"
  kms_master_key_id                 = "alias/aws/sqs"
  kms_data_key_reuse_period_seconds = 300
}
`, queue)
}

func testAccAWSSQSConfigWithTags(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name = "%s"

  tags {
    Environment = "production"
    Usage = "original"
  }
}
`, r)
}

func testAccAWSSQSConfigWithTagsChanged(r string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "queue" {
  name = "%s"

  tags {
    Usage = "changed"
  }
}
`, r)
}
