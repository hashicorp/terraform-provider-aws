// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSuppressEquivalentTopicSubscriptionDeliveryPolicy(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		old        string
		new        string
		equivalent bool
	}{
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"maxDelayTarget":20,"minDelayTarget":5,"numRetries":5}}`,
			equivalent: true,
		},
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5}}`,
			equivalent: true,
		},
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"throttlePolicy":{}}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"throttlePolicy":{}}`,
			equivalent: true,
		},
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":6}}`,
			equivalent: false,
		},
		{
			old:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries":5,"numMaxDelayRetries":null,"numNoDelayRetries":null,"numMinDelayRetries":null,"backoffFunction":null},"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":false}`,
			new:        `{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20}}`,
			equivalent: false,
		},
		{
			old:        `{"healthyRetryPolicy":null,"sicklyRetryPolicy":null,"throttlePolicy":null,"guaranteed":true}`,
			new:        `{"guaranteed":true}`,
			equivalent: true,
		},
	}

	for i, tc := range testCases {
		actual := tfsns.SuppressEquivalentTopicSubscriptionDeliveryPolicy("", tc.old, tc.new, nil)
		if actual != tc.equivalent {
			t.Fatalf("Test Case %d: Got: %t Expected: %t", i, actual, tc.equivalent)
		}
	}
}

func TestAccSNSTopicSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sns", regexache.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "confirmation_timeout_in_minutes", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "confirmation_was_authenticated", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEndpoint, "aws_sqs_queue.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "endpoint_auto_confirms", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", ""),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "pending_confirmation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "sqs"),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "replay_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "subscription_role_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTopicARN, "aws_sns_topic.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
		},
	})
}

func TestAccSNSTopicSubscription_filterPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	filterPolicy1 := `{"key1":["val1"],"key2":["val2"]}`
	filterPolicy2 := `{"key3":["val3"],"key4":["val4"]}`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_filterPolicy(rName, strconv.Quote(filterPolicy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", filterPolicy1),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageAttributes"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute update
			{
				Config: testAccTopicSubscriptionConfig_filterPolicy(rName, strconv.Quote(filterPolicy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", filterPolicy2),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageAttributes"),
				),
			},
			// Test attribute removal
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", ""),
				),
			},
		},
	})
}

func TestAccSNSTopicSubscription_filterPolicyScope(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_filterPolicyScope(rName, strconv.Quote("MessageBody")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageBody"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			{
				Config: testAccTopicSubscriptionConfig_filterPolicyScope(rName, strconv.Quote("MessageAttributes")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageAttributes"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			{
				Config: testAccTopicSubscriptionConfig_filterPolicyScope(rName, strconv.Quote("MessageBody")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageBody"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			{
				Config: testAccTopicSubscriptionConfig_filterPolicyScope(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageAttributes"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			{
				Config: testAccTopicSubscriptionConfig_filterPolicyScope(rName, strconv.Quote("MessageBody")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageBody"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test transition from MessageAttributes to nested MessageBody ...
			{
				Config: testAccTopicSubscriptionConfig_filterPolicyScope(rName, strconv.Quote("MessageAttributes")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageAttributes"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			{
				Config: testAccTopicSubscriptionConfig_nestedFilterPolicyScope(rName, strconv.Quote("MessageBody"), true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageBody"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// ... and transition from nested MessageBody back to flat MessageAttributes
			{
				Config: testAccTopicSubscriptionConfig_filterPolicyScope(rName, strconv.Quote("MessageAttributes")),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "filter_policy_scope", "MessageAttributes"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
		},
	})
}

func TestAccSNSTopicSubscription_filterPolicyScope_policyNotSet(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccTopicSubscriptionConfig_filterPolicyScope_policyNotSet(rName),
				ExpectError: regexache.MustCompile(`filter_policy is required when filter_policy_scope is set`),
			},
		},
	})
}

func TestAccSNSTopicSubscription_deliveryPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_deliveryPolicy(rName, strconv.Quote(`{"healthyRetryPolicy":{"minDelayTarget":5,"maxDelayTarget":20,"numRetries": 5}}`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					testAccCheckTopicSubscriptionDeliveryPolicyAttribute(&attributes, &tfsns.TopicSubscriptionDeliveryPolicy{
						HealthyRetryPolicy: &tfsns.TopicSubscriptionDeliveryPolicyHealthyRetryPolicy{
							MaxDelayTarget: 20,
							MinDelayTarget: 5,
							NumRetries:     5,
						},
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute update
			{
				Config: testAccTopicSubscriptionConfig_deliveryPolicy(rName, strconv.Quote(`{"healthyRetryPolicy":{"minDelayTarget":3,"maxDelayTarget":78,"numRetries": 11}}`)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					testAccCheckTopicSubscriptionDeliveryPolicyAttribute(&attributes, &tfsns.TopicSubscriptionDeliveryPolicy{
						HealthyRetryPolicy: &tfsns.TopicSubscriptionDeliveryPolicyHealthyRetryPolicy{
							MaxDelayTarget: 78,
							MinDelayTarget: 3,
							NumRetries:     11,
						},
					}),
				),
			},
			// Test attribute removal
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
				),
			},
		},
	})
}

func TestAccSNSTopicSubscription_redrivePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	dlqName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	updatedDlqName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_redrivePolicy(rName, dlqName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					testAccCheckTopicSubscriptionRedrivePolicyAttribute(&attributes, dlqName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute update
			{
				Config: testAccTopicSubscriptionConfig_redrivePolicy(rName, updatedDlqName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					testAccCheckTopicSubscriptionRedrivePolicyAttribute(&attributes, updatedDlqName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute removal
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "redrive_policy", ""),
				),
			},
		},
	})
}

func TestAccSNSTopicSubscription_rawMessageDelivery(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_rawMessageDelivery(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
			// Test attribute update
			{
				Config: testAccTopicSubscriptionConfig_rawMessageDelivery(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", acctest.CtFalse),
				),
			},
			// Test attribute removal
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccSNSTopicSubscription_autoConfirmingEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_autoConfirmingEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
		},
	})
}

func TestAccSNSTopicSubscription_autoConfirmingSecuredEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_autoConfirmingSecuredEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"confirmation_timeout_in_minutes",
					"endpoint_auto_confirms",
				},
			},
		},
	})
}

func TestAccSNSTopicSubscription_email(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_email(rName, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sns", regexache.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "confirmation_was_authenticated", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEndpoint, acctest.DefaultEmailAddress),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "pending_confirmation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, names.AttrEmail),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTopicARN, "aws_sns_topic.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccSNSTopicSubscription_firehose(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_firehose(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "sns", regexache.MustCompile(fmt.Sprintf("%s:.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "delivery_policy", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEndpoint, "aws_kinesis_firehose_delivery_stream.test_stream", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "filter_policy", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrProtocol, "firehose"),
					resource.TestCheckResourceAttr(resourceName, "raw_message_delivery", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTopicARN, "aws_sns_topic.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "subscription_role_arn", "aws_iam_role.firehose_role", names.AttrARN),
				),
			},
		},
	})
}

func TestAccSNSTopicSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsns.ResourceTopicSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSNSTopicSubscription_Disappears_topic(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicSubscriptionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicSubscriptionExists(ctx, resourceName, &attributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsns.ResourceTopic(), "aws_sns_topic.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTopicSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sns_topic_subscription" {
				continue
			}

			output, err := tfsns.FindSubscriptionAttributesByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if output["Protocol"] == names.AttrEmail {
				continue
			}

			return fmt.Errorf("SNS Topic Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTopicSubscriptionExists(ctx context.Context, n string, v *map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		output, err := tfsns.FindSubscriptionAttributesByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func testAccCheckTopicSubscriptionDeliveryPolicyAttribute(attributes *map[string]string, expectedDeliveryPolicy *tfsns.TopicSubscriptionDeliveryPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiDeliveryPolicyJSONString, ok := (*attributes)["DeliveryPolicy"]

		if !ok {
			return fmt.Errorf("DeliveryPolicy attribute not found in attributes: %s", attributes)
		}

		var apiDeliveryPolicy tfsns.TopicSubscriptionDeliveryPolicy
		if err := json.Unmarshal([]byte(apiDeliveryPolicyJSONString), &apiDeliveryPolicy); err != nil {
			return fmt.Errorf("unable to unmarshal SNS Topic Subscription delivery policy JSON (%s): %s", apiDeliveryPolicyJSONString, err)
		}

		if reflect.DeepEqual(apiDeliveryPolicy, *expectedDeliveryPolicy) {
			return nil
		}

		return fmt.Errorf("SNS Topic Subscription delivery policy did not match:\n\nReceived\n\n%s\n\nExpected\n\n%s\n\n", apiDeliveryPolicy, *expectedDeliveryPolicy)
	}
}

func testAccCheckTopicSubscriptionRedrivePolicyAttribute(attributes *map[string]string, expectedRedrivePolicyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiRedrivePolicyJSONString, ok := (*attributes)["RedrivePolicy"]

		if !ok {
			return fmt.Errorf("RedrivePolicy attribute not found in attributes: %s", attributes)
		}

		var apiRedrivePolicy tfsns.TopicSubscriptionRedrivePolicy
		if err := json.Unmarshal([]byte(apiRedrivePolicyJSONString), &apiRedrivePolicy); err != nil {
			return fmt.Errorf("unable to unmarshal SNS Topic Subscription redrive policy JSON (%s): %s", apiRedrivePolicyJSONString, err)
		}

		expectedRedrivePolicy := tfsns.TopicSubscriptionRedrivePolicy{
			DeadLetterTargetArn: arn.ARN{
				AccountID: acctest.AccountID(),
				Partition: acctest.Partition(),
				Region:    acctest.Region(),
				Resource:  expectedRedrivePolicyResource,
				Service:   "sqs",
			}.String(),
		}

		if reflect.DeepEqual(apiRedrivePolicy, expectedRedrivePolicy) {
			return nil
		}

		return fmt.Errorf("SNS Topic Subscription redrive policy did not match:\n\nReceived\n\n%s\n\nExpected\n\n%s\n\n", apiRedrivePolicy, expectedRedrivePolicy)
	}
}

func testAccTopicSubscriptionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.test.arn
}
`, rName)
}

func testAccTopicSubscriptionConfig_filterPolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn     = aws_sns_topic.test.arn
  protocol      = "sqs"
  endpoint      = aws_sqs_queue.test.arn
  filter_policy = %[2]s
}
`, rName, policy)
}

func testAccTopicSubscriptionConfig_nestedFilterPolicyScope(rName, scope string, nested bool) string {
	filterPolicy := `jsonencode({"key1"=["value1"]})`
	if nested {
		filterPolicy = `jsonencode({"key2"={"key1"=["value1"]}})`
	}
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn           = aws_sns_topic.test.arn
  protocol            = "sqs"
  endpoint            = aws_sqs_queue.test.arn
  filter_policy       = %[2]s
  filter_policy_scope = %[3]s
}
`, rName, filterPolicy, scope)
}
func testAccTopicSubscriptionConfig_filterPolicyScope(rName, scope string) string {
	return testAccTopicSubscriptionConfig_nestedFilterPolicyScope(rName, scope, false)
}

func testAccTopicSubscriptionConfig_filterPolicyScope_policyNotSet(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn           = aws_sns_topic.test.arn
  protocol            = "sqs"
  endpoint            = aws_sqs_queue.test.arn
  filter_policy_scope = "MessageBody"
}
`, rName)
}

func testAccTopicSubscriptionConfig_deliveryPolicy(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  delivery_policy = %[2]s
  endpoint        = aws_sqs_queue.test.arn
  protocol        = "sqs"
  topic_arn       = aws_sns_topic.test.arn
}
`, rName, policy)
}

func testAccTopicSubscriptionConfig_redrivePolicy(rName, dlqName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sqs_queue" "test_dlq" {
  name = %[2]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  redrive_policy = jsonencode({ deadLetterTargetArn : aws_sqs_queue.test_dlq.arn })
  endpoint       = aws_sqs_queue.test.arn
  protocol       = "sqs"
  topic_arn      = aws_sns_topic.test.arn
}
`, rName, dlqName)
}

func testAccTopicSubscriptionConfig_rawMessageDelivery(rName string, rawMessageDelivery bool) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sqs_queue" "test" {
  name = %[1]q

  sqs_managed_sse_enabled = true
}

resource "aws_sns_topic_subscription" "test" {
  endpoint             = aws_sqs_queue.test.arn
  protocol             = "sqs"
  raw_message_delivery = %[2]t
  topic_arn            = aws_sns_topic.test.arn
}
`, rName, rawMessageDelivery)
}

func testAccTopicSubscriptionConfig_autoConfirmingEndpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api" "test" {
  name        = %[1]q
  description = "Terraform Acceptance test for SNS subscription"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_rest_api.test.root_resource_id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id             = aws_api_gateway_rest_api.test.id
  resource_id             = aws_api_gateway_rest_api.test.root_resource_id
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = aws_lambda_function.lambda.invoke_arn
}

resource "aws_api_gateway_integration_response" "test" {
  depends_on  = [aws_api_gateway_integration.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "iam_for_lambda" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "policy" {
  name = %[1]q
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.arn
  principal     = "apigateway.${data.aws_partition.current.dns_suffix}"
  source_arn    = "${aws_api_gateway_deployment.test.execution_arn}/*"
}

resource "aws_lambda_function" "lambda" {
  filename         = "test-fixtures/lambda_confirm_sns.zip"
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "main.confirm_subscription"
  source_code_hash = filebase64sha256("test-fixtures/lambda_confirm_sns.zip")
  runtime          = "python3.7"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on  = [aws_api_gateway_integration_response.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "acctest"
}

resource "aws_sns_topic_subscription" "test" {
  depends_on             = [aws_lambda_permission.apigw_lambda]
  topic_arn              = aws_sns_topic.test.arn
  protocol               = "https"
  endpoint               = aws_api_gateway_deployment.test.invoke_url
  endpoint_auto_confirms = true
}
`, rName)
}

func testAccTopicSubscriptionConfig_autoConfirmingSecuredEndpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_api_gateway_rest_api" "test" {
  name        = %[1]q
  description = "Terraform Acceptance test for SNS subscription"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_rest_api.test.root_resource_id
  http_method   = "POST"
  authorization = "CUSTOM"
  authorizer_id = aws_api_gateway_authorizer.test.id
}

resource "aws_api_gateway_method_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "200"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = true
  }
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id             = aws_api_gateway_rest_api.test.id
  resource_id             = aws_api_gateway_rest_api.test.root_resource_id
  http_method             = aws_api_gateway_method.test.http_method
  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = aws_lambda_function.lambda.invoke_arn
}

resource "aws_api_gateway_integration_response" "test" {
  depends_on  = [aws_api_gateway_integration.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_rest_api.test.root_resource_id
  http_method = aws_api_gateway_method.test.http_method
  status_code = aws_api_gateway_method_response.test.status_code

  response_parameters = {
    "method.response.header.Access-Control-Allow-Origin" = "'*'"
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "iam_for_lambda" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "policy" {
  name = %[1]q
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "logs:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.arn
  principal     = "apigateway.${data.aws_partition.current.dns_suffix}"
  source_arn    = "${aws_api_gateway_deployment.test.execution_arn}/*"
}

resource "aws_lambda_function" "lambda" {
  filename         = "test-fixtures/lambda_confirm_sns.zip"
  function_name    = %[1]q
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "main.confirm_subscription"
  source_code_hash = filebase64sha256("test-fixtures/lambda_confirm_sns.zip")
  runtime          = "python3.7"
}

resource "aws_api_gateway_deployment" "test" {
  depends_on  = [aws_api_gateway_integration_response.test]
  rest_api_id = aws_api_gateway_rest_api.test.id
  stage_name  = "acctest"
}

resource "aws_iam_role" "invocation_role" {
  name = "%[1]s-2"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = "%[1]s-2"
  role = aws_iam_role.invocation_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.authorizer.arn}"
    }
  ]
}
EOF
}

resource "aws_api_gateway_authorizer" "test" {
  name                   = %[1]q
  rest_api_id            = aws_api_gateway_rest_api.test.id
  authorizer_uri         = aws_lambda_function.authorizer.invoke_arn
  authorizer_credentials = aws_iam_role.invocation_role.arn
}

resource "aws_lambda_function" "authorizer" {
  filename         = "test-fixtures/lambda_basic_authorizer.zip"
  source_code_hash = filebase64sha256("test-fixtures/lambda_basic_authorizer.zip")
  function_name    = "%[1]s-2"
  role             = aws_iam_role.iam_for_lambda.arn
  handler          = "main.authenticate"
  runtime          = "nodejs16.x"

  environment {
    variables = {
      AUTH_USER = "davematthews"
      AUTH_PASS = "granny"
    }
  }
}

resource "aws_api_gateway_gateway_response" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  status_code   = "401"
  response_type = "UNAUTHORIZED"

  response_templates = {
    "application/json" = "{'message':$context.error.messageString}"
  }

  response_parameters = {
    "gatewayresponse.header.WWW-Authenticate" = "'Basic'"
  }
}

resource "aws_sns_topic_subscription" "test" {
  depends_on             = [aws_lambda_permission.apigw_lambda]
  topic_arn              = aws_sns_topic.test.arn
  protocol               = "https"
  endpoint               = replace(aws_api_gateway_deployment.test.invoke_url, "https://", "https://davematthews:granny@")
  endpoint_auto_confirms = true

  confirmation_timeout_in_minutes = 3
}
`, rName)
}

func testAccTopicSubscriptionConfig_email(rName, email string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  topic_arn = aws_sns_topic.test.arn
  protocol  = "email"
  endpoint  = %[2]q
}
`, rName, email)
}

func testAccTopicSubscriptionConfig_firehose(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_subscription" "test" {
  endpoint              = aws_kinesis_firehose_delivery_stream.test_stream.arn
  protocol              = "firehose"
  topic_arn             = aws_sns_topic.test.arn
  subscription_role_arn = aws_iam_role.firehose_role.arn
}
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

data "aws_partition" "current" {}

resource "aws_iam_role" "firehose_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
  {
	"Action": "sts:AssumeRole",
	"Principal": {
	  "Service": ["sns.${data.aws_partition.current.dns_suffix}","firehose.${data.aws_partition.current.dns_suffix}"]
	},
	"Effect": "Allow",
	"Sid": ""
  }
]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test_stream" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose_role.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`, rName)
}
