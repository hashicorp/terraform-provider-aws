package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
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
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSpotDatafeedSubscription(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotDatafeedSubscriptionDestroy,
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

func testAccSpotDatafeedSubscription_disappears(t *testing.T) {
	var subscription ec2.SpotDatafeedSubscription
	resourceName := "aws_spot_datafeed_subscription.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckSpotDatafeedSubscription(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotDatafeedSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotDatafeedSubscription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSpotDatafeedSubscriptionExists(resourceName, &subscription),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceSpotDataFeedSubscription(), resourceName),
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

		if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidSpotDatafeedNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error describing EC2 Spot Datafeed Subscription: %w", err)
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

	if tfawserr.ErrCodeEquals(err, tfec2.ErrCodeInvalidSpotDatafeedNotFound) {
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
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    grant {
      grantee {
        id   = "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0" # EC2 Account
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }
}

resource "aws_spot_datafeed_subscription" "test" {
  # Must have bucket grants configured
  depends_on = [aws_s3_bucket_acl.test]

  bucket = aws_s3_bucket.test.bucket
}
`, rName)
}
