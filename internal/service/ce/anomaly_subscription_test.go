package ce_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfce "github.com/hashicorp/terraform-provider-aws/internal/service/ce"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCEAnomalySubscription_basic(t *testing.T) {
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "frequency", "DAILY"),
					resource.TestCheckResourceAttrSet(resourceName, "monitor_arn_list.#"),
					resource.TestCheckResourceAttrSet(resourceName, "subscriber.#"),
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

func TestAccCEAnomalySubscription_Frequency(t *testing.T) {
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_Frequency(rName, "DAILY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "DAILY"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_Frequency(rName, "WEEKLY"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "frequency", "WEEKLY"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_MonitorArnList(t *testing.T) {
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.0", "aws_ce_anomaly_monitor.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_MonitorArnList(rName2, rName3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.0", "aws_ce_anomaly_monitor.test1", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "monitor_arn_list.1", "aws_ce_anomaly_monitor.test2", "arn"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_Subscriber(t *testing.T) {
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.address", "abc@example.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_Subscriber1(rName, "abcd@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.address", "abcd@example.com"),
				),
			},
			{
				Config: testAccAnomalySubscriptionConfig_Subscriber2(rName, "cba@example.com", "dcbad@example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.0.address", "cba@example.com"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.1.type", "EMAIL"),
					resource.TestCheckResourceAttr(resourceName, "subscriber.1.address", "dcbad@example.com"),
				),
			},
		},
	})
}

func TestAccCEAnomalySubscription_Threshold(t *testing.T) {
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_Threshold(rName, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "threshold", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAnomalySubscriptionConfig_Threshold(rName, 200),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "threshold", "200"),
				),
			},
		},
	})
}
func TestAccCEAnomalySubscription_Tags(t *testing.T) {
	resourceName := "aws_ce_anomaly_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAnomalySubscriptionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, costexplorer.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAnomalySubscriptionConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
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
				Config: testAccAnomalySubscriptionConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAnomalySubscriptionConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAnomalySubscriptionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAnomalySubscriptionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cost Explorer Anomaly Subscription is set")
		}

		resp, err := conn.GetAnomalySubscriptions(&costexplorer.GetAnomalySubscriptionsInput{SubscriptionArnList: aws.StringSlice([]string{rs.Primary.ID})})

		if err != nil {
			return fmt.Errorf("Error describing Cost Explorer Anomaly Subscription: %s", err.Error())
		}

		if resp == nil || len(resp.AnomalySubscriptions) < 1 {
			return fmt.Errorf("Anomaly Subscription (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAnomalySubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CEConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ce_anomaly_subscription" {
			continue
		}

		resp, err := conn.GetAnomalySubscriptions(&costexplorer.GetAnomalySubscriptionsInput{SubscriptionArnList: aws.StringSlice([]string{rs.Primary.ID})})

		if err != nil {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResAnomalySubscription, rs.Primary.ID, err)
		}

		if resp != nil && len(resp.AnomalySubscriptions) > 0 {
			return names.Error(names.CE, names.ErrActionCheckingDestroyed, tfce.ResAnomalySubscription, rs.Primary.ID, errors.New("still exists"))
		}
	}

	return nil

}

func testAccAnomalySubscriptionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }
}
`, rName)
}

func testAccAnomalySubscriptionConfig_MonitorArnList(rName string, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test1" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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

resource "aws_ce_anomaly_monitor" "test2" {
  name = %[2]q
  type = "CUSTOM"

  specification = <<JSON
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
    aws_ce_anomaly_monitor.test1.arn,
    aws_ce_anomaly_monitor.test2.arn,
  ]
  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }
}
`, rName, rName2)
}

func testAccAnomalySubscriptionConfig_Frequency(rName string, rFrequency string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
  frequency = %[2]q

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }
}
`, rName, rFrequency)
}

func testAccAnomalySubscriptionConfig_Subscriber1(rName string, rAddress string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
  ]

  subscriber {
    type    = "EMAIL"
    address = %[2]q
  }
}
`, rName, rAddress)
}

func testAccAnomalySubscriptionConfig_Subscriber2(rName string, rAddress1 string, rAddress2 string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
`, rName, rAddress1, rAddress2)
}

func testAccAnomalySubscriptionConfig_Threshold(rName string, rThreshold int) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
  threshold = %[2]d
  frequency = "WEEKLY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }
}
`, rName, rThreshold)
}

func testAccAnomalySubscriptionConfig_Tags1(rName string, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAnomalySubscriptionConfig_Tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ce_anomaly_monitor" "test" {
  name = %[1]q
  type = "CUSTOM"

  specification = <<JSON
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
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_monitor.test.arn,
  ]

  subscriber {
    type    = "EMAIL"
    address = "abc@example.com"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
