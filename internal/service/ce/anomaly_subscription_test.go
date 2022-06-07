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

func TestAccAnomalySubscription_basic(t *testing.T) {
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
			return fmt.Errorf("Anomaly Subscription (%s) not found", rs.Primary.Attributes["name"])
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
resource "aws_ce_anomaly_subscription" "test" {
  name      = "DailyAnomalySubscription"
  threshold = 100
  frequency = "DAILY"

  monitor_arn_list = [
    aws_ce_anomaly_subscription.anomaly_subscription.arn,
  ]

  subscribers = [
    {
      type    = "EMAIL"
      address = "abc@example.com"
    }
  ]
}
`, rName)
}
