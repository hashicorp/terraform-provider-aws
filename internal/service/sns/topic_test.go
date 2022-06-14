package sns_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(sns.EndpointsID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Invalid protocol type: firehose",
		"Unknown attribute FifoTopic",
	)
}

func TestAccSNSTopic_basic(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGenerated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "application_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_sample_rate", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "sns", regexp.MustCompile(`terraform-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "display_name", ""),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", "false"),
					resource.TestCheckResourceAttr(resourceName, "firehose_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "firehose_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "firehose_success_feedback_sample_rate", "0"),
					resource.TestCheckResourceAttr(resourceName, "http_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "http_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "http_success_feedback_sample_rate", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "lambda_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "lambda_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "lambda_success_feedback_sample_rate", "0"),
					create.TestCheckResourceAttrNameGenerated(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttr(resourceName, "sqs_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_sample_rate", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccSNSTopic_name(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", "false"),
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

func TestAccSNSTopic_namePrefix(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := "tf-acc-test-prefix-"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", rName),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", "false"),
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

func TestAccSNSTopic_policy(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandString(10)
	expectedPolicy := fmt.Sprintf(`{"Statement":[{"Sid":"Stmt1445931846145","Effect":"Allow","Principal":{"AWS":"*"},"Action":"sns:Publish","Resource":"arn:%s:sns:%s::example"}],"Version":"2012-10-17","Id":"Policy1445931846145"}`, acctest.Partition(), acctest.Region())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					testAccCheckNSTopicHasPolicy(resourceName, expectedPolicy),
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

func TestAccSNSTopic_withIAMRole(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
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

func TestAccSNSTopic_withFakeIAMRole(t *testing.T) {
	rName := sdkacctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTopicConfig_fakeIAMRole(rName),
				ExpectError: regexp.MustCompile(`PrincipalNotFound`),
			},
		},
	})
}

func TestAccSNSTopic_withDeliveryPolicy(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandString(10)
	expectedPolicy := `{"http":{"defaultHealthyRetryPolicy": {"minDelayTarget": 20,"maxDelayTarget": 20,"numMaxDelayRetries": 0,"numRetries": 3,"numNoDelayRetries": 0,"numMinDelayRetries": 0,"backoffFunction": "linear"},"disableSubscriptionOverrides": false}}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_deliveryPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					testAccCheckNSTopicHasDeliveryPolicy(resourceName, expectedPolicy),
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

func TestAccSNSTopic_deliveryStatus(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	iamRoleResourceName := "aws_iam_role.example"

	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_deliveryStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttrPair(resourceName, "application_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_sample_rate", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "application_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "lambda_success_feedback_sample_rate", "90"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "http_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "http_success_feedback_sample_rate", "80"),
					resource.TestCheckResourceAttrPair(resourceName, "http_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "sqs_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_sample_rate", "70"),
					resource.TestCheckResourceAttrPair(resourceName, "sqs_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "firehose_failure_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "firehose_success_feedback_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "firehose_success_feedback_sample_rate", "60"),
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

func TestAccSNSTopic_NameGenerated_fifoTopic(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGeneratedFIFO,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					create.TestCheckResourceAttrNameWithSuffixGenerated(resourceName, "name", tfsns.FIFOTopicNameSuffix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", "true"),
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

func TestAccSNSTopic_Name_fifoTopic(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) + tfsns.FIFOTopicNameSuffix

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameFIFO(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", "true"),
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

func TestAccSNSTopic_NamePrefix_fifoTopic(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := "tf-acc-test-prefix-"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_namePrefixFIFO(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					create.TestCheckResourceAttrNameWithSuffixFromPrefix(resourceName, "name", rName, tfsns.FIFOTopicNameSuffix),
					resource.TestCheckResourceAttr(resourceName, "name_prefix", rName),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", "true"),
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

func TestAccSNSTopic_fifoWithContentBasedDeduplication(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_fifoContentBasedDeduplication(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", "true"),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test attribute update
			{
				Config: testAccTopicConfig_fifoContentBasedDeduplication(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", "false"),
				),
			},
		},
	})
}

func TestAccSNSTopic_fifoExpectContentBasedDeduplicationError(t *testing.T) {
	rName := sdkacctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTopicConfig_expectContentBasedDeduplicationError(rName),
				ExpectError: regexp.MustCompile(`content-based deduplication can only be set for FIFO topics`),
			},
		},
	})
}

func TestAccSNSTopic_encryption(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sns"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
				),
			},
		},
	})
}

func TestAccSNSTopic_tags(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTopicConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSNSTopic_disappears(t *testing.T) {
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sns.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGenerated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(resourceName, &attributes),
					acctest.CheckResourceDisappears(acctest.Provider, tfsns.ResourceTopic(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckNSTopicHasPolicy(n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Queue URL specified")
		}

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS topic with that ARN exists")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTopicAttributes(params)
		if err != nil {
			return err
		}

		var actualPolicyText string
		for k, v := range resp.Attributes {
			if k == "Policy" {
				actualPolicyText = aws.StringValue(v)
				break
			}
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckNSTopicHasDeliveryPolicy(n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Queue URL specified")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

		params := &sns.GetTopicAttributesInput{
			TopicArn: aws.String(rs.Primary.ID),
		}
		resp, err := conn.GetTopicAttributes(params)
		if err != nil {
			return err
		}

		var actualPolicyText string
		for k, v := range resp.Attributes {
			if k == "DeliveryPolicy" {
				actualPolicyText = aws.StringValue(v)
				break
			}
		}

		equivalent := verify.SuppressEquivalentJSONDiffs("", actualPolicyText, expectedPolicyText, nil)

		if !equivalent {
			return fmt.Errorf("Non-equivalent delivery policy error:\n\nexpected: %s\n\n     got: %s",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckTopicDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sns_topic" {
			continue
		}

		_, err := tfsns.FindTopicAttributesByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SNS Topic %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckTopicExists(n string, v *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSConn

		output, err := tfsns.FindTopicAttributesByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

const testAccTopicConfig_nameGenerated = `
resource "aws_sns_topic" "test" {}
`

const testAccTopicConfig_nameGeneratedFIFO = `
resource "aws_sns_topic" "test" {
  fifo_topic = true
}
`

func testAccTopicConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccTopicConfig_nameFIFO(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name       = %[1]q
  fifo_topic = true
}
`, rName)
}

func testAccTopicConfig_namePrefix(prefix string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name_prefix = %[1]q
}
`, prefix)
}

func testAccTopicConfig_namePrefixFIFO(prefix string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name_prefix = %[1]q
  fifo_topic  = true
}
`, prefix)
}

func testAccTopicConfig_policy(r string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = "example-%s"

  policy = <<EOF
{
  "Statement": [
    {
      "Sid": "Stmt1445931846145",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "sns:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}::example"
    }
  ],
  "Version": "2012-10-17",
  "Id": "Policy1445931846145"
}
EOF
}
`, r)
}

// Test for https://github.com/hashicorp/terraform/issues/3660
func testAccTopicConfig_iamRole(r string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = "tf_acc_test_%[1]s"
  path = "/test/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = "tf-acc-test-with-iam-role-%[1]s"

  policy = <<EOF
{
  "Statement": [
    {
      "Sid": "Stmt1445931846145",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${aws_iam_role.example.arn}"
      },
      "Action": "sns:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}::example"
    }
  ],
  "Version": "2012-10-17",
  "Id": "Policy1445931846145"
}
EOF
}
`, r)
}

// Test for https://github.com/hashicorp/terraform/issues/14024
func testAccTopicConfig_deliveryPolicy(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "tf_acc_test_delivery_policy_%s"

  delivery_policy = <<EOF
{
  "http": {
    "defaultHealthyRetryPolicy": {
      "minDelayTarget": 20,
      "maxDelayTarget": 20,
      "numRetries": 3,
      "numMaxDelayRetries": 0,
      "numNoDelayRetries": 0,
      "numMinDelayRetries": 0,
      "backoffFunction": "linear"
    },
    "disableSubscriptionOverrides": false
  }
}
EOF
}
`, r)
}

// Test for https://github.com/hashicorp/terraform/issues/3660
func testAccTopicConfig_fakeIAMRole(r string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = "tf_acc_test_fake_iam_role_%s"

  policy = <<EOF
{
  "Statement": [
    {
      "Sid": "Stmt1445931846145",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::012345678901:role/wooo"
      },
      "Action": "sns:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.name}::example"
    }
  ],
  "Version": "2012-10-17",
  "Id": "Policy1445931846145"
}
EOF
}
`, r)
}

func testAccTopicConfig_deliveryStatus(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  depends_on                               = [aws_iam_role_policy.example]
  name                                     = "sns-delivery-status-topic-%[1]s"
  application_success_feedback_role_arn    = aws_iam_role.example.arn
  application_success_feedback_sample_rate = 100
  application_failure_feedback_role_arn    = aws_iam_role.example.arn
  lambda_success_feedback_role_arn         = aws_iam_role.example.arn
  lambda_success_feedback_sample_rate      = 90
  lambda_failure_feedback_role_arn         = aws_iam_role.example.arn
  http_success_feedback_role_arn           = aws_iam_role.example.arn
  http_success_feedback_sample_rate        = 80
  http_failure_feedback_role_arn           = aws_iam_role.example.arn
  sqs_success_feedback_role_arn            = aws_iam_role.example.arn
  sqs_success_feedback_sample_rate         = 70
  sqs_failure_feedback_role_arn            = aws_iam_role.example.arn
  firehose_success_feedback_sample_rate    = 60
  firehose_failure_feedback_role_arn       = aws_iam_role.example.arn
  firehose_success_feedback_role_arn       = aws_iam_role.example.arn
}

data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = "sns-delivery-status-role-%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "sns.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "example" {
  name = "sns-delivery-status-role-policy-%[1]s"
  role = aws_iam_role.example.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:PutMetricFilter",
        "logs:PutRetentionPolicy"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}
`, r)
}

func testAccTopicConfig_encryption(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name              = %[1]q
  kms_master_key_id = "alias/aws/sns"
}
`, rName)
}

func testAccTopicConfig_fifoContentBasedDeduplication(r string, cbd bool) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name                        = "terraform-test-topic-%s.fifo"
  fifo_topic                  = true
  content_based_deduplication = %t
}
`, r, cbd)
}

func testAccTopicConfig_expectContentBasedDeduplicationError(r string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name                        = "terraform-test-topic-%s"
  content_based_deduplication = true
}
`, r)
}

func testAccTopicConfig_tags1(r, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "terraform-test-topic-%s"

  tags = {
    %q = %q
  }
}
`, r, tag1Key, tag1Value)
}

func testAccTopicConfig_tags2(r, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = "terraform-test-topic-%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, r, tag1Key, tag1Value, tag2Key, tag2Value)
}
