package ec2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccEC2SpotDatafeedSubscription_serial(t *testing.T) {
	cases := map[string]func(t *testing.T){
		"basic":      testAccSpotDatafeedSubscription_basic,
		"disappears": testAccSpotDatafeedSubscription_disappears,
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccSpotDatafeedSubscription_basic(t *testing.T) {
	var subscription ec2.SpotDatafeedSubscription
	resourceName := "aws_spot_datafeed_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckSpotDatafeedSubscription(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSpotDatafeedSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotDatafeedSubscription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotDatafeedSubscriptionExists(resourceName, &subscription),
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

func testAccCheckSpotDatafeedSubscriptionDisappears(subscription *ec2.SpotDatafeedSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err := conn.DeleteSpotDatafeedSubscription(&ec2.DeleteSpotDatafeedSubscriptionInput{})
		if err != nil {
			return err
		}

		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			_, err := conn.DescribeSpotDatafeedSubscription(&ec2.DescribeSpotDatafeedSubscriptionInput{})
			if err != nil {
				cgw, ok := err.(awserr.Error)
				if ok && cgw.Code() == "InvalidSpotDatafeed.NotFound" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving Spot Datafeed Subscription: %s", err))
			}
			return resource.RetryableError(fmt.Errorf("Waiting for Spot Datafeed Subscription"))
		})
	}
}

func testAccSpotDatafeedSubscription_disappears(t *testing.T) {
	var subscription ec2.SpotDatafeedSubscription
	resourceName := "aws_spot_datafeed_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckSpotDatafeedSubscription(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSpotDatafeedSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotDatafeedSubscription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotDatafeedSubscriptionExists(resourceName, &subscription),
					testAccCheckSpotDatafeedSubscriptionDisappears(&subscription),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSpotDatafeedSubscriptionExists(n string, subscription *ec2.SpotDatafeedSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.DescribeSpotDatafeedSubscription(&ec2.DescribeSpotDatafeedSubscriptionInput{})
		if err != nil {
			return err
		}

		*subscription = *resp.SpotDatafeedSubscription

		return nil
	}
}

func testAccCheckSpotDatafeedSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_spot_datafeed_subscription" {
			continue
		}

		_, err := conn.DescribeSpotDatafeedSubscription(&ec2.DescribeSpotDatafeedSubscriptionInput{})

		if tfawserr.ErrCodeEquals(err, "InvalidSpotDatafeed.NotFound") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error descripting EC2 Spot Datafeed Subscription: %w", err)
		}
	}

	return nil
}

func testAccPreCheckSpotDatafeedSubscription(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeSpotDatafeedSubscriptionInput{}

	_, err := conn.DescribeSpotDatafeedSubscription(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if tfawserr.ErrCodeEquals(err, "InvalidSpotDatafeed.NotFound") {
		return
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSpotDatafeedSubscription(rName string) string {
	return fmt.Sprintf(`
data "aws_canonical_user_id" "current" {}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  grant {
    id          = data.aws_canonical_user_id.current.id
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }

  grant {
    id          = "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0" # EC2 Account
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }
}

resource "aws_spot_datafeed_subscription" "test" {
  bucket = aws_s3_bucket.test.bucket
}
`, rName)
}
