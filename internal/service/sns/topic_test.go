// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.SNSServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Invalid protocol type: firehose",
		"Unknown attribute FifoTopic",
	)
}

func TestAccSNSTopic_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGenerated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "application_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_sample_rate", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "archive_policy", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sns", regexache.MustCompile(`terraform-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "beginning_archive_time", ""),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, ""),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "firehose_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "firehose_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "firehose_success_feedback_sample_rate", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "http_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "http_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "http_success_feedback_sample_rate", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "lambda_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "lambda_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "lambda_success_feedback_sample_rate", acctest.Ct0),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwner),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttr(resourceName, "signature_version", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "sqs_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_sample_rate", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "tracing_config", ""),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSNSTopic_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGenerated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsns.ResourceTopic(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSNSTopic_name(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := "tf-acc-test-prefix-"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, rName),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtFalse),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	expectedPolicy := fmt.Sprintf(`{"Statement":[{"Sid":"Stmt1445931846145","Effect":"Allow","Principal":{"AWS":"*"},"Action":"sns:Publish","Resource":"arn:%s:sns:%s::example"}],"Version":"2012-10-17","Id":"Policy1445931846145"}`, acctest.Partition(), acctest.Region())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					testAccCheckTopicHasPolicy(ctx, resourceName, expectedPolicy),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_iamRole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					acctest.CheckResourceAttrJMESPair(resourceName, names.AttrPolicy, "Statement[0].Principal.AWS", "aws_iam_role.example", names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTopicConfig_fakeIAMRole(rName),
				ExpectError: regexache.MustCompile(`PrincipalNotFound`),
			},
		},
	})
}

func TestAccSNSTopic_withDeliveryPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	expectedPolicy := `{"http":{"defaultHealthyRetryPolicy": {"minDelayTarget": 20,"maxDelayTarget": 20,"numMaxDelayRetries": 0,"numRetries": 3,"numNoDelayRetries": 0,"numMinDelayRetries": 0,"backoffFunction": "linear"},"disableSubscriptionOverrides": false}}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_deliveryPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					testAccCheckTopicHasDeliveryPolicy(ctx, resourceName, expectedPolicy),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	iamRoleResourceName := "aws_iam_role.example"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_deliveryStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttrPair(resourceName, "application_success_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_sample_rate", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "application_failure_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_success_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lambda_success_feedback_sample_rate", "90"),
					resource.TestCheckResourceAttrPair(resourceName, "lambda_failure_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "http_success_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "http_success_feedback_sample_rate", "80"),
					resource.TestCheckResourceAttrPair(resourceName, "http_failure_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "sqs_success_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_sample_rate", "70"),
					resource.TestCheckResourceAttrPair(resourceName, "sqs_failure_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "firehose_failure_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "firehose_success_feedback_role_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "firehose_success_feedback_sample_rate", "60"),
					resource.TestCheckResourceAttr(resourceName, "tracing_config", "Active"),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGeneratedFIFO,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					acctest.CheckResourceAttrNameWithSuffixGenerated(resourceName, names.AttrName, tfsns.FIFOTopicNameSuffix),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix) + tfsns.FIFOTopicNameSuffix

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameFIFO(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "archive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "beginning_archive_time", ""),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := "tf-acc-test-prefix-"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_namePrefixFIFO(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					acctest.CheckResourceAttrNameWithSuffixFromPrefix(resourceName, names.AttrName, rName, tfsns.FIFOTopicNameSuffix),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, rName),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_fifoContentBasedDeduplication(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", acctest.CtTrue),
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
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSNSTopic_fifoExpectContentBasedDeduplicationError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTopicConfig_expectContentBasedDeduplicationError(rName),
				ExpectError: regexache.MustCompile(`content-based deduplication can only be set for FIFO topics`),
			},
		},
	})
}

func TestAccSNSTopic_fifoWithArchivePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policy1 := `
{
  "MessageRetentionPeriod": "30"
}
`
	policy2 := `
{
  "MessageRetentionPeriod": "45"
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_fifoArchivePolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, "archive_policy", regexache.MustCompile(`"MessageRetentionPeriod":\s*"30"`)),
					resource.TestCheckResourceAttrSet(resourceName, "beginning_archive_time"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicConfig_fifoArchivePolicy(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestMatchResourceAttr(resourceName, "archive_policy", regexache.MustCompile(`"MessageRetentionPeriod":\s*"45"`)),
					resource.TestCheckResourceAttrSet(resourceName, "beginning_archive_time"),
				),
			},
			// "Invalid state: Cannot delete a topic with an ArchivePolicy".
			{
				Config: testAccTopicConfig_fifoArchivePolicy(rName, "{}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "archive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "beginning_archive_time", ""),
				),
			},
		},
	})
}

func TestAccSNSTopic_fifoExpectArchivePolicyError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTopicConfig_expectArchivePolicyError(rName),
				ExpectError: regexache.MustCompile(`message archive policy can only be set for FIFO topics`),
			},
		},
	})
}

func TestAccSNSTopic_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sns"),
					resource.TestCheckResourceAttr(resourceName, "signature_version", acctest.Ct2),
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
					testAccCheckTopicExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
				),
			},
		},
	})
}

func testAccCheckTopicHasPolicy(ctx context.Context, n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		attributes, err := tfsns.FindTopicAttributesWithValidAWSPrincipalsByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		actualPolicyText := attributes[tfsns.TopicAttributeNamePolicy]

		if actualPolicyText == "" {
			return fmt.Errorf("SNS Topic Policy (%s) not found", rs.Primary.ID)
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicyText)

		if err != nil {
			return fmt.Errorf("testing policy equivalence: %s", err)
		}

		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckTopicHasDeliveryPolicy(ctx context.Context, n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		attributes, err := tfsns.FindTopicAttributesWithValidAWSPrincipalsByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		actualPolicyText := attributes[tfsns.TopicAttributeNameDeliveryPolicy]

		if actualPolicyText == "" {
			return fmt.Errorf("SNS Topic Delivery Policy (%s) not found", rs.Primary.ID)
		}

		equivalent := verify.SuppressEquivalentJSONDiffs("", actualPolicyText, expectedPolicyText, nil)

		if !equivalent {
			return fmt.Errorf("Non-equivalent delivery policy error:\n\nexpected: %s\n\n     got: %s",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckTopicDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sns_topic" {
				continue
			}

			_, err := tfsns.FindTopicAttributesByARN(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckTopicExists(ctx context.Context, n string, v *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		output, err := tfsns.FindTopicAttributesByARN(ctx, conn, rs.Primary.ID)

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

func testAccTopicConfig_policy(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = %[1]q

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
`, rName)
}

// Test for https://github.com/hashicorp/terraform/issues/3660
func testAccTopicConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_sns_topic" "test" {
  name = %[1]q

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

resource "aws_iam_role" "example" {
  name = %[1]q
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
`, rName)
}

// Test for https://github.com/hashicorp/terraform/issues/14024
func testAccTopicConfig_deliveryPolicy(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q

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
`, rName)
}

// Test for https://github.com/hashicorp/terraform/issues/3660
func testAccTopicConfig_fakeIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_sns_topic" "test" {
  name = %[1]q

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
`, rName)
}

func testAccTopicConfig_deliveryStatus(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  depends_on = [aws_iam_role_policy.example]

  name                                     = %[1]q
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

  tracing_config = "Active"
}

data "aws_partition" "current" {}

resource "aws_iam_role" "example" {
  name = %[1]q
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
  name = %[1]q
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
`, rName)
}

func testAccTopicConfig_encryption(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name              = %[1]q
  kms_master_key_id = "alias/aws/sns"
  signature_version = 2
}
`, rName)
}

func testAccTopicConfig_fifoContentBasedDeduplication(rName string, cbd bool) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name                        = "%[1]s.fifo"
  fifo_topic                  = true
  content_based_deduplication = %[2]t
}
`, rName, cbd)
}

func testAccTopicConfig_expectContentBasedDeduplicationError(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name                        = %[1]q
  content_based_deduplication = true
}
`, rName)
}

func testAccTopicConfig_fifoArchivePolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name           = "%[1]s.fifo"
  fifo_topic     = true
  archive_policy = %[2]q
}
`, rName, policy)
}

func testAccTopicConfig_expectArchivePolicyError(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name           = %[1]q
  archive_policy = <<EOF
{
  "MessageRetentionPeriod": "30"
}
EOF
}
`, rName)
}
