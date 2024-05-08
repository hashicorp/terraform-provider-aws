// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecuritylake "github.com/hashicorp/terraform-provider-aws/internal/service/securitylake"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSubscriberNotification_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_subscriber_notification.test"
	rName := randomCustomLogSourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberNotificationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberNotificationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "subscriber_id", "aws_securitylake_subscriber.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.sqs_notification_configuration.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration"},
			},
		},
	})
}

func testAccSubscriberNotification_https(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_subscriber_notification.test"
	rName := randomCustomLogSourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberNotificationConfig_https(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberNotificationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "subscriber_id", "aws_securitylake_subscriber.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.https_notification_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.https_notification_configuration.0.endpoint", "aws_apigatewayv2_api.test", "api_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.https_notification_configuration.0.target_role_arn", "aws_iam_role.event_bridge", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration"},
			},
		},
	})
}

func testAccSubscriberNotification_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_subscriber_notification.test"
	rName := randomCustomLogSourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberNotificationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberNotificationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceSubscriberNotification, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSubscriberNotification_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_securitylake_subscriber_notification.test"
	rName := randomCustomLogSourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberNotificationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberNotificationExists(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration"},
			},
			{
				Config: testAccSubscriberNotificationConfig_https(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberNotificationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "subscriber_id", "aws_securitylake_subscriber.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.https_notification_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.https_notification_configuration.0.endpoint", "aws_apigatewayv2_api.test", "api_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.https_notification_configuration.0.target_role_arn", "aws_iam_role.event_bridge", names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration"},
			},
			{
				Config: testAccSubscriberNotificationConfig_https_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberNotificationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "subscriber_id", "aws_securitylake_subscriber.test", names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.https_notification_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.https_notification_configuration.0.endpoint", "aws_apigatewayv2_api.test", "api_endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration.0.https_notification_configuration.0.target_role_arn", "aws_iam_role.event_bridge", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.https_notification_configuration.0.http_method", "POST"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configuration"},
			},
		},
	})
}

func testAccCheckSubscriberNotificationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securitylake_subscriber_notification" {
				continue
			}

			_, _, err := tfsecuritylake.FindSubscriberNotificationByEndPointID(ctx, conn, rs.Primary.Attributes["subscriber_id"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Lake Subscriber Notification %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSubscriberNotificationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityLakeClient(ctx)

		_, _, err := tfsecuritylake.FindSubscriberNotificationByEndPointID(ctx, conn, rs.Primary.Attributes["subscriber_id"])

		return err
	}
}

func testAccSubscriberNotification_config(rName string) string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), fmt.Sprintf(`
resource "aws_securitylake_subscriber" "test" {
  subscriber_name = %[1]q

  source {
    custom_log_source_resource {
      source_name    = aws_securitylake_custom_log_source.test.source_name
      source_version = aws_securitylake_custom_log_source.test.source_version
    }
  }
  subscriber_identity {
    external_id = "example"
    principal   = data.aws_caller_identity.current.account_id
  }
}

resource "aws_securitylake_custom_log_source" "test" {
  source_name = %[1]q

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "%[1]s-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/service-role/"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "glue.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "s3:GetObject",
      "s3:PutObject"
    ],
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role" "event_bridge" {
  name = "AmazonSecurityLakeSubscriberEventBridge"
  path = "/service-role/"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "event_bridge" {
  name = "AmazonSecurityLakeSubscriberEventBridgePolicy"
  role = aws_iam_role.event_bridge.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "events:InvokeApiDestination"
      ],
      "Resource": "*"
    }
  ]
}
POLICY

  depends_on = [aws_securitylake_data_lake.test]
}
`, rName))
}

func testAccSubscriberNotificationConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSubscriberNotification_config(rName), `
resource "aws_securitylake_subscriber_notification" "test" {
  subscriber_id = aws_securitylake_subscriber.test.id
  configuration {
    sqs_notification_configuration {}
  }
}
`)
}

func testAccSubscriberNotificationConfig_https(rName string) string {
	return acctest.ConfigCompose(
		testAccSubscriberNotification_config(rName), fmt.Sprintf(`
resource "aws_securitylake_subscriber_notification" "test" {
  subscriber_id = aws_securitylake_subscriber.test.id
  configuration {
    https_notification_configuration {
      endpoint        = aws_apigatewayv2_api.test.api_endpoint
      target_role_arn = aws_iam_role.event_bridge.arn
    }
  }
}

resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName))
}

func testAccSubscriberNotificationConfig_https_update(rName string) string {
	return acctest.ConfigCompose(
		testAccSubscriberNotification_config(rName), fmt.Sprintf(`
resource "aws_securitylake_subscriber_notification" "test" {
  subscriber_id = aws_securitylake_subscriber.test.id
  configuration {
    https_notification_configuration {
      endpoint        = aws_apigatewayv2_api.test.api_endpoint
      target_role_arn = aws_iam_role.event_bridge.arn
      http_method     = "POST"
    }
  }
}

resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, rName))
}
