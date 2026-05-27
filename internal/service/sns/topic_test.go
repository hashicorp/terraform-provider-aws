// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/YakDriver/regexache"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/envvar"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGenerated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "application_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "application_success_feedback_sample_rate", "0"),
					resource.TestCheckResourceAttr(resourceName, "archive_policy", ""),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sns", regexache.MustCompile(`terraform-.+$`)),
					resource.TestCheckResourceAttr(resourceName, "beginning_archive_time", ""),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, ""),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_scope", ""),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtFalse),
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
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "terraform-"),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwner),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttr(resourceName, "signature_version", "0"),
					resource.TestCheckResourceAttr(resourceName, "sqs_failure_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "sqs_success_feedback_sample_rate", "0"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGenerated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsns.ResourceTopic(), resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sns", rName),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sns", regexache.MustCompile(fmt.Sprintf(`%s.+$`, rName))),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	expectedPolicy := fmt.Sprintf(`{"Statement":[{"Sid":"Stmt1445931846145","Effect":"Allow","Principal":{"AWS":"*"},"Action":"sns:Publish","Resource":"arn:%s:sns:%s::example"}],"Version":"2012-10-17","Id":"Policy1445931846145"}`, acctest.Partition(), acctest.Region())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					testAccCheckTopicHasPolicy(ctx, t, resourceName, expectedPolicy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_iamRole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	expectedPolicy := `{"http":{"defaultHealthyRetryPolicy": {"minDelayTarget": 20,"maxDelayTarget": 20,"numMaxDelayRetries": 0,"numRetries": 3,"numNoDelayRetries": 0,"numMinDelayRetries": 0,"backoffFunction": "linear"},"disableSubscriptionOverrides": false}}`

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_deliveryPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					testAccCheckTopicHasDeliveryPolicy(ctx, t, resourceName, expectedPolicy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_deliveryStatus(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameGeneratedFIFO,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix) + tfsns.FIFOTopicNameSuffix

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_nameFIFO(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_namePrefixFIFO(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_fifoContentBasedDeduplication(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
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
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "content_based_deduplication", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSNSTopic_fifoExpectContentBasedDeduplicationError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_fifoArchivePolicy(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
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
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestMatchResourceAttr(resourceName, "archive_policy", regexache.MustCompile(`"MessageRetentionPeriod":\s*"45"`)),
					resource.TestCheckResourceAttrSet(resourceName, "beginning_archive_time"),
				),
			},
			// "Invalid state: Cannot delete a topic with an ArchivePolicy".
			{
				Config: testAccTopicConfig_fifoArchivePolicy(rName, "{}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "archive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "beginning_archive_time", ""),
				),
			},
		},
	})
}

func TestAccSNSTopic_fifoExpectArchivePolicyError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccTopicConfig_expectArchivePolicyError(rName),
				ExpectError: regexache.MustCompile(`message archive policy can only be set for FIFO topics`),
			},
		},
	})
}

func TestAccSNSTopic_fifoWithHighThroughput(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_fifoThroughputScope(rName, "Topic"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_scope", "Topic"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicConfig_fifoThroughputScope(rName, "MessageGroup"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "fifo_topic", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fifo_throughput_scope", "MessageGroup"),
				),
			},
		},
	})
}

func TestAccSNSTopic_encryption(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", "alias/aws/sns"),
					resource.TestCheckResourceAttr(resourceName, "signature_version", "2"),
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
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "kms_master_key_id", ""),
				),
			},
		},
	})
}

/*
Create an IAM role to be assumed via:

data "aws_caller_identity" "current" {}

	resource "aws_iam_role" "terraform_execution_role" {
	  name = "sns-topic-test-terraform-execution-role"

	  assume_role_policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [
	      {
	        Effect = "Allow"
	        Principal = {
	          AWS = [
	            # Prefer the converted role ARN (works for callers using assumed-role sessions)
	            data.aws_caller_identity.current.arn,
	            # optionally keep account-root if you need cross-account/account-wide allow
	            "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
	          ]
	        }
	        # Allow both AssumeRole and SetSourceIdentity so callers may pass SourceIdentity
	        Action = [
	          "sts:AssumeRole",
	          "sts:SetSourceIdentity"
	        ]
	      }
	    ]
	  })

	  tags = {
	    Name        = "sns-topic-test-terraform-execution-role"
	    Purpose     = "Terraform execution role that supports SourceIdentity"
	    Environment = "demo"
	  }
	}

# Policy that allows many AWS operations but specifically excludes iam:PassRole

	resource "aws_iam_policy" "terraform_execution_policy" {
	  name = "sns-topic-test-terraform-execution-policy"

	  policy = jsonencode({
	    Version = "2012-10-17"
	    Statement = [

	      {
	        Sid    = "AllowEC2Operations"
	        Effect = "Allow"
	        Action = [
	          "ec2:*"
	        ]
	        Resource = "*"
	      },
	      {
	        Sid    = "AllowSNSOperations"
	        Effect = "Allow"
	        Action = [
	          "sns:*"
	        ]
	        Resource = "*"
	      },
	      {
	        Sid    = "AllowS3Operations"
	        Effect = "Allow"
	        Action = [
	          "s3:*"
	        ]
	        Resource = "*"
	      },
	      {
	        Sid    = "AllowIAMReadOperations"
	        Effect = "Allow"
	        Action = [
	          "iam:GetRole",
	          "iam:GetRolePolicy",
	          "iam:ListRolePolicies",
	          "iam:CreatePolicy",
	          "iam:ListAttachedRolePolicies",
	          "iam:GetPolicy",
	          "iam:GetPolicyVersion",
	          "iam:ListPolicyVersions",
	          "iam:GetInstanceProfile",
	          "iam:ListInstanceProfiles",
	          "iam:CreateRole",
	          "iam:DeleteRole",
	          "iam:CreateInstanceProfile",
	          "iam:DeleteInstanceProfile",
	          "iam:AddRoleToInstanceProfile",
	          "iam:RemoveRoleFromInstanceProfile",
	          "iam:AttachRolePolicy",
	          "iam:ListInstanceProfilesForRole",
	          "iam:DetachRolePolicy",
	          "iam:PutRolePolicy",
	          "iam:DeleteRolePolicy",
	          "iam:TagRole",
	          "iam:UntagRole",
	          "iam:DeletePolicy"
	        ]
	        Resource = "*"
	      }
	    ]
	  })
	}

# Attach the policy to the role

	resource "aws_iam_role_policy_attachment" "terraform_execution_policy_attachment" {
	  role       = aws_iam_role.terraform_execution_role.name
	  policy_arn = aws_iam_policy.terraform_execution_policy.arn
	}

	output "execution_role_arn" {
	  value = aws_iam_role.terraform_execution_role.arn
	}

Then run this test with TF_ACC_ASSUME_ROLE_ARN=<output value>.
*/
func TestAccSNSTopic_iamEventualConsistency(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAssumeRoleARN(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicConfig_iamEventualConsistency(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, resourceName, &attributes),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckTopicHasPolicy(ctx context.Context, t *testing.T, n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).SNSClient(ctx)

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
			return fmt.Errorf("testing policy equivalence: %w", err)
		}

		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s",
				expectedPolicyText, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckTopicHasDeliveryPolicy(ctx context.Context, t *testing.T, n string, expectedPolicyText string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).SNSClient(ctx)

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

func testAccCheckTopicDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SNSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sns_topic" {
				continue
			}

			_, err := tfsns.FindTopicAttributesByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckTopicExists(ctx context.Context, t *testing.T, n string, v *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).SNSClient(ctx)

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
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.region}::example"
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
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.region}::example"
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
        "AWS": "arn:${data.aws_partition.current.partition}:iam::123456789012:role/wooo"
      },
      "Action": "sns:Publish",
      "Resource": "arn:${data.aws_partition.current.partition}:sns:${data.aws_region.current.region}::example"
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

func testAccTopicConfig_fifoThroughputScope(rName string, throughputScope string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name                  = "%[1]s.fifo"
  fifo_topic            = true
  fifo_throughput_scope = "%[2]s"
}
`, rName, throughputScope)
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

func testAccTopicConfig_iamEventualConsistency(rName string) string {
	//lintignore:AT004
	return fmt.Sprintf(`
provider "aws" {
  assume_role {
    role_arn     = %[1]q
    session_name = "TerraformSNSTopicPassRoleTest"
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "sns_feedback_role" {
  name = "%[2]s-sns-cloudwatch-logs-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.${data.aws_partition.current.dns_suffix}"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# Policy for SNS to write to CloudWatch Logs.
resource "aws_iam_role_policy" "sns_feedback_policy" {
  name = "%[2]s-sns-cloudwatch-logs-policy"
  role = aws_iam_role.sns_feedback_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:PutMetricFilter",
          "logs:PutRetentionPolicy"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_sns_topic" "test" {
  name = %[2]q

  lambda_failure_feedback_role_arn = aws_iam_role.sns_feedback_role.arn
}

# Grant PassRole permission to the execution role.
resource "aws_iam_policy" "terraform_passrole_policy" {
  name = "%[2]s-terraform_passrole_policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "ExplicitlyAllowPassRole"
        Effect = "Allow"
        Action = [
          "iam:PassRole"
        ]
        Resource = "*"
      }
    ]
  })
}

# Attach the policy to the execution role.
resource "aws_iam_role_policy_attachment" "terraform_execution_policy_attachment" {
  # Extract the role name from the ARN.
  role       = trimprefix(provider::aws::arn_parse(%[1]q).resource, "role/")
  policy_arn = aws_iam_policy.terraform_passrole_policy.arn
}
`, os.Getenv(envvar.AccAssumeRoleARN), rName)
}
