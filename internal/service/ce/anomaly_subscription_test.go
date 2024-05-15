// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCEAnomalySubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalySubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "ce", regexache.MustCompile(`anomalysubscription/.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "frequency", "DAILY"),
					resource.TestCheckResourceAttrSet(resourceName, "monitor_arn_list.#"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.address", address),
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

func TestAccCEAnomalySubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalySubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfce.ResourceAnomalySubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCEAnomalySubscription_Frequency(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalySubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_frequency(rName, "DAILY", address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "frequency", "DAILY"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_frequency(rName, "WEEKLY", address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "frequency", "WEEKLY"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_MonitorARNList(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalySubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.0", "aws_ce_anomaly_monitor.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_monitorARNList(rName, rName2, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.0", "aws_ce_anomaly_monitor.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.1", "aws_ce_anomaly_monitor.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_Subscriber(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalySubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.address", address1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.address", address2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_subscriber2(rName, address1, address2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttrSet(resourceName, "subscriber.0.type"),
					resource.TestCheckResourceAttrSet(resourceName, "subscriber.0.address"),
					resource.TestCheckResourceAttrSet(resourceName, "subscriber.1.type"),
					resource.TestCheckResourceAttrSet(resourceName, "subscriber.1.address"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_subscriberSNS(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "SNS"),
					resource.TestCheckResourceAttrPair(resourceName, "subscriber.0.address", "aws_sns_topic.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	var subscription awstypes.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAnomalySubscriptionDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_tags1(rName, address, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_tags2(rName, address, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAnomalySubscriptionConfig_tags1(rName, address, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(ctx, resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckAnomalySubscriptionExists(ctx context.Context, n string, v *awstypes.AnomalySubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		output, err := tfce.FindAnomalySubscriptionByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAnomalySubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ce_anomaly_subscription" {
				continue
			}

			_, err := tfce.FindAnomalySubscriptionByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cost Explorer Anomaly Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAnomalySubscriptionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name         = %[1]q
  monitor_type = "CUSTOM"

  monitor_specification = <<JSON
{
	"And": null,
	"CostCategories": null,
	"Dimensions": null,
	"Not": null,
	"Or": null,
	"Tags": {
		"Key": "user:CostCenter",
		"MatchOptions": null,
		"Values": [
			"10000"
		]
	}
}
JSON
}
`, rName)
}

func testAccAnomalySubscriptionConfig_basic(rName, address string) string {
	return acctest.ConfigCompose(testAccAnomalySubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[2]q
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      values        = ["100.0"]
      match_options = ["GREATER_THAN_OR_EQUAL"]
    }
  }
}
`, rName, address))
}

func testAccAnomalySubscriptionConfig_monitorARNList(rName, rName2, address string) string {
	return acctest.ConfigCompose(testAccAnomalySubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test2" {
  name         = %[2]q
  monitor_type = "CUSTOM"

  monitor_specification = <<JSON
  {
	  "And": null,
	  "CostCategories": null,
	  "Dimensions": null,
	  "Not": null,
	  "Or": null,
	  "Tags": {
		  "Key": "user:CostCenter",
		  "MatchOptions": null,
		  "Values": [
			  "10000"
		  ]
	  }
  }
  JSON
}

resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  frequency = "WEEKLY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
    aws_ce_anomaly_monitor.test2.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[3]q
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      values        = ["100.0"]
      match_options = ["GREATER_THAN_OR_EQUAL"]
    }
  }
}
`, rName, rName2, address))
}

func testAccAnomalySubscriptionConfig_frequency(rName, rFrequency, address string) string {
	return acctest.ConfigCompose(testAccAnomalySubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  frequency = %[2]q

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[3]q
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      values        = ["100.0"]
      match_options = ["GREATER_THAN_OR_EQUAL"]
    }
  }
}
`, rName, rFrequency, address))
}

func testAccAnomalySubscriptionConfig_subscriber2(rName, address1, address2 string) string {
	return acctest.ConfigCompose(testAccAnomalySubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  frequency = "WEEKLY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[2]q
  }

  subscriber {
    type    = "EMAIL"
    address = %[3]q
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      values        = ["100.0"]
      match_options = ["GREATER_THAN_OR_EQUAL"]
    }
  }
}
`, rName, address1, address2))
}

func testAccAnomalySubscriptionConfig_subscriberSNS(rName string) string {
	return acctest.ConfigCompose(testAccAnomalySubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "test" {
  policy_id = %[1]q

  statement {
    sid = "AWSAnomalyDetectionSNSPublishingPermissions"

    actions = [
      "SNS:Publish",
    ]

    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["costalerts.amazonaws.com"]
    }

    resources = [
      aws_sns_topic.test.arn,
    ]
  }

  statement {
    sid = %[1]q

    actions = [
      "SNS:Subscribe",
      "SNS:SetTopicAttributes",
      "SNS:RemovePermission",
      "SNS:Receive",
      "SNS:Publish",
      "SNS:ListSubscriptionsByTopic",
      "SNS:GetTopicAttributes",
      "SNS:DeleteTopic",
      "SNS:AddPermission",
    ]

    condition {
      test     = "StringEquals"
      variable = "AWS:SourceOwner"

      values = [
        data.aws_caller_identity.current.account_id,
      ]
    }

    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    resources = [
      aws_sns_topic.test.arn,
    ]
  }
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn

  policy = data.aws_iam_policy_document.test.json
}

resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  frequency = "IMMEDIATE"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "SNS"
    address = aws_sns_topic.test.arn
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      values        = ["100.0"]
      match_options = ["GREATER_THAN_OR_EQUAL"]
    }
  }

  depends_on = [
    aws_sns_topic_policy.test,
  ]
}
`, rName))
}

func testAccAnomalySubscriptionConfig_tags1(rName, address, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAnomalySubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[4]q
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      values        = ["100.0"]
      match_options = ["GREATER_THAN_OR_EQUAL"]
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1, address))
}

func testAccAnomalySubscriptionConfig_tags2(rName, address, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAnomalySubscriptionConfig_base(rName), fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[6]q
  }

  threshold_expression {
    dimension {
      key           = "ANOMALY_TOTAL_IMPACT_ABSOLUTE"
      values        = ["100.0"]
      match_options = ["GREATER_THAN_OR_EQUAL"]
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2, address))
}
