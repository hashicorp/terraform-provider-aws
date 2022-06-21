package ce_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/costexplorer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCEAnomalySubscription_basic(t *testing.T) {
	var subscription costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "ce", regexp.MustCompile(`anomalysubscription/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "frequency", "DAILY"),
					resource.TestCheckResourceAttrSet(resourceName, "monitor_arn_list.#"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.address", address),
					resource.TestCheckResourceAttr(resourceName, "threshold", "100"),
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
	var subscription costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					acctest.CheckResourceDisappears(acctest.Provider, tfce.ResourceAnomalySubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCEAnomalySubscription_Frequency(t *testing.T) {
	var subscription costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_frequency(rName, "DAILY", address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
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
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "frequency", "WEEKLY"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_MonitorARNList(t *testing.T) {
	var subscription costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.0", "aws_ce_anomaly_monitor.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_monitorARNList(rName2, rName3, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.0", "aws_ce_anomaly_monitor.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.1", "aws_ce_anomaly_monitor.test2", "arn"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_Subscriber(t *testing.T) {
	var subscription costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address1 := acctest.RandomEmailAddress(domain)
	address2 := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_basic(rName, address1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
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
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
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
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
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
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "SNS"),
					resource.TestCheckResourceAttrPair(resourceName, "subscriber.0.address", "aws_sns_topic.test", "arn"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_Threshold(t *testing.T) {
	var subscription costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_threshold(rName, 100, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "threshold", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_threshold(rName, 200, address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "threshold", "200"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_Tags(t *testing.T) {
	var subscription costexplorer.AnomalySubscription
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	address := acctest.RandomEmailAddress(domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_tags1(rName, "key1", "value1", address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
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
				Config: testAccAnomalySubscriptionConfig_tags2(rName, "key1", "value1updated", "key2", "value2", address),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAnomalySubscriptionConfig_tags1(rName, "key2", "value2", address),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName, &subscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAnomalySubscriptionExists(n string, anomalySubscription *costexplorer.AnomalySubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cost Explorer Anomaly Subscription is set")
		}

		resp, err := tfce.FindAnomalySubscriptionByARN(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if resp == nil {
			return fmt.Errorf("Cost Explorer %q does not exist", rs.Primary.ID)
		}

		*anomalySubscription = *resp

		return nil
	}
}

func testAccCheckAnomalySubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_anomaly_subscription" {
			continue
		}

		_, err := tfce.FindAnomalySubscriptionByARN(context.Background(), conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResAnomalySubscription, rs.Primary.ID, errors.New("still exists"))

	}
	return nil
}

func testAccAnomalySubscriptionConfigBase(rName string) string {
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
		"Key": "CostCenter",
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

func testAccAnomalySubscriptionConfig_basic(rName string, address string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  threshold = 100
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[2]q
  }
}
`, rName, address))
}

func testAccAnomalySubscriptionConfig_monitorARNList(rName string, rName2 string, address string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
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
		  "Key": "CostCenter",
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
  threshold = 100
  frequency = "WEEKLY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
    aws_ce_anomaly_monitor.test2.arn,
  ]
  subscriber {
    type    = "EMAIL"
    address = %[3]q
  }
}
`, rName, rName2, address))
}

func testAccAnomalySubscriptionConfig_frequency(rName string, rFrequency string, address string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  threshold = 100
  frequency = %[2]q

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[3]q
  }
}
`, rName, rFrequency, address))
}

func testAccAnomalySubscriptionConfig_subscriber2(rName string, address1 string, address2 string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  threshold = 100
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
}
`, rName, address1, address2))
}

func testAccAnomalySubscriptionConfig_subscriberSNS(rName string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
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
  threshold = 100000000
  frequency = "IMMEDIATE"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "SNS"
    address = aws_sns_topic.test.arn
  }

  depends_on = [
    aws_sns_topic_policy.test,
  ]
}
`, rName))
}

func testAccAnomalySubscriptionConfig_threshold(rName string, rThreshold int, address string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  threshold = %[2]d
  frequency = "WEEKLY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[3]q
  }
}
`, rName, rThreshold, address))
}

func testAccAnomalySubscriptionConfig_tags1(rName string, tagKey1, tagValue1 string, address string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  threshold = 100
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[4]q
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1, address))
}

func testAccAnomalySubscriptionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string, address string) string {
	return acctest.ConfigCompose(
		testAccAnomalySubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ce_anomaly_subscription" "test" {
  name      = %[1]q
  threshold = 100
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = %[6]q
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2, address))
}
