package configservice_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccDeliveryChannel_basic(t *testing.T) {
	var dc configservice.DeliveryChannel
	rInt := sdkacctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)
	expectedBucketName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeliveryChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryChannelConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryChannelExists("aws_config_delivery_channel.foo", &dc),
					testAccCheckDeliveryChannelName("aws_config_delivery_channel.foo", expectedName, &dc),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "name", expectedName),
					resource.TestCheckResourceAttr("aws_config_delivery_channel.foo", "s3_bucket_name", expectedBucketName),
				),
			},
		},
	})
}

func testAccDeliveryChannel_allParams(t *testing.T) {
	resourceName := "aws_config_delivery_channel.foo"
	var dc configservice.DeliveryChannel
	rInt := sdkacctest.RandInt()
	expectedName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)
	expectedBucketName := fmt.Sprintf("tf-acc-test-awsconfig-%d", rInt)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeliveryChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryChannelConfig_allParams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeliveryChannelExists(resourceName, &dc),
					testAccCheckDeliveryChannelName(resourceName, expectedName, &dc),
					resource.TestCheckResourceAttr(resourceName, "name", expectedName),
					resource.TestCheckResourceAttr(resourceName, "s3_bucket_name", expectedBucketName),
					resource.TestCheckResourceAttr(resourceName, "s3_key_prefix", "one/two/three"),
					resource.TestCheckResourceAttrPair(resourceName, "s3_kms_key_arn", "aws_kms_key.k", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", "aws_sns_topic.t", "arn"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_delivery_properties.0.delivery_frequency", "Six_Hours"),
				),
			},
		},
	})
}

func testAccDeliveryChannel_importBasic(t *testing.T) {
	resourceName := "aws_config_delivery_channel.foo"
	rInt := sdkacctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, configservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDeliveryChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryChannelConfig_basic(rInt),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDeliveryChannelName(n, desired string, obj *configservice.DeliveryChannel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if *obj.Name != desired {
			return fmt.Errorf("Expected name: %q, given: %q", desired, *obj.Name)
		}
		return nil
	}
}

func testAccCheckDeliveryChannelExists(n string, obj *configservice.DeliveryChannel) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No delivery channel ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn
		out, err := conn.DescribeDeliveryChannels(&configservice.DescribeDeliveryChannelsInput{
			DeliveryChannelNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe delivery channel: %s", err)
		}
		if len(out.DeliveryChannels) < 1 {
			return fmt.Errorf("No delivery channel found when describing %q", rs.Primary.Attributes["name"])
		}

		dc := out.DeliveryChannels[0]
		*obj = *dc

		return nil
	}
}

func testAccCheckDeliveryChannelDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigServiceConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_delivery_channel" {
			continue
		}

		resp, err := conn.DescribeDeliveryChannels(&configservice.DescribeDeliveryChannelsInput{
			DeliveryChannelNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.DeliveryChannels) != 0 &&
				*resp.DeliveryChannels[0].Name == rs.Primary.Attributes["name"] {
				return fmt.Errorf("Delivery Channel still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccDeliveryChannelConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "foo" {
  name     = "tf-acc-test-%d"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "r" {
  name = "tf-acc-test-awsconfig-%d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
  name = "tf-acc-test-awsconfig-%d"
  role = aws_iam_role.r.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "b" {
  bucket        = "tf-acc-test-awsconfig-%d"
  force_destroy = true
}

resource "aws_config_delivery_channel" "foo" {
  name           = "tf-acc-test-awsconfig-%d"
  s3_bucket_name = aws_s3_bucket.b.bucket
  depends_on     = [aws_config_configuration_recorder.foo]
}
`, randInt, randInt, randInt, randInt, randInt)
}

func testAccDeliveryChannelConfig_allParams(randInt int) string {
	return fmt.Sprintf(`
resource "aws_config_configuration_recorder" "foo" {
  name     = "tf-acc-test-%d"
  role_arn = aws_iam_role.r.arn
}

resource "aws_iam_role" "r" {
  name = "tf-acc-test-awsconfig-%d"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "config.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "p" {
  name = "tf-acc-test-awsconfig-%d"
  role = aws_iam_role.r.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.b.arn}",
        "${aws_s3_bucket.b.arn}/*"
      ]
    },
    {
        "Effect": "Allow",
        "Action": [
            "kms:Decrypt",
            "kms:GenerateDataKey"
            ],
            "Resource": "${aws_kms_key.k.arn}"
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "b" {
  bucket        = "tf-acc-test-awsconfig-%d"
  force_destroy = true
}

resource "aws_sns_topic" "t" {
  name = "tf-acc-test-%d"
}

resource "aws_kms_key" "k" {
  description             = "tf-acc-test-awsconfig-%d"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "tf-acc-test-awsconfig-%d",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_config_delivery_channel" "foo" {
  name           = "tf-acc-test-awsconfig-%d"
  s3_bucket_name = aws_s3_bucket.b.bucket
  s3_key_prefix  = "one/two/three"
  s3_kms_key_arn = aws_kms_key.k.arn
  sns_topic_arn  = aws_sns_topic.t.arn

  snapshot_delivery_properties {
    delivery_frequency = "Six_Hours"
  }

  depends_on = [aws_config_configuration_recorder.foo]
}
`, randInt, randInt, randInt, randInt, randInt, randInt, randInt, randInt)
}
