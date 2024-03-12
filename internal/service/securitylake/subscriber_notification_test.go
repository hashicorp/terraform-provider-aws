// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
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
		},
	})
}

func testAccSubscriberNotification_https(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_subscriber_notification.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberNotificationConfig_https(rName),
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
		},
	})
}

func testAccSubscriberNotification_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_securitylake_subscriber_notification.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SecurityLake)
			acctest.PreCheckOrganizationsAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityLakeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSubscriberNotificationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSubscriberNotificationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubscriberNotificationExists(ctx, resourceName),
					// acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsecuritylake.ResourceSubscriber, resourceName),
					// resource.TestCheckResourceAttr(resourceName, "subscriber_name", rName),
					// resource.TestCheckResourceAttr(resourceName, "access_type", "S3"),
					// resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_name", "ROUTE53"),
					// resource.TestCheckResourceAttr(resourceName, "source.0.aws_log_source_resource.0.source_version", "1.0"),
					// resource.TestCheckResourceAttr(resourceName, "subscriber_identity.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "subscriber_identity.0.external_id", "example"),
				),
				ExpectNonEmptyPlan: true,
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

			_, err := tfsecuritylake.FindSubscriberNotificationByEndPointID(ctx, conn, rs.Primary.Attributes["subscriber_id"], rs.Primary.Attributes["endpoint_id"])

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

		_, err := tfsecuritylake.FindSubscriberNotificationByEndPointID(ctx, conn, rs.Primary.Attributes["subscriber_id"], rs.Primary.Attributes["endpoint_id"])

		return err
	}
}

func testAccSubscriberNotification_config() string {
	return acctest.ConfigCompose(testAccDataLakeConfig_basic(), `
resource "aws_iam_role" "test" {
  name = "AmazonSecurityLakeCustomDataGlueCrawler-windows-sysmon"
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
  name = "AmazonSecurityLakeCustomDataGlueCrawler-windows-sysmon"
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

  depends_on = [aws_securitylake_data_lake.test]
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
"Statement": [{
	"Action": "sts:AssumeRole",
	"Principal": {
	"Service": "events.amazonaws.com"
	},
	"Effect": "Allow"
}]
}
POLICY
}

resource "aws_iam_role_policy" "event_bridge" {
	name = "AmazonSecurityLakeSubscriberEventBridgePolicy"
	role = aws_iam_role.event_bridge.name
  
	policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [{
		"Effect": "Allow",
		"Action": ["events:InvokeApiDestination"],
		"Resource": "*"
}]
}
  POLICY
  
	depends_on = [aws_securitylake_data_lake.test]
}

resource "aws_securitylake_custom_log_source" "test" {
  source_name    = "windows-sysmon"
  source_version = "1.0"
  event_classes  = ["FILE_ACTIVITY"]

  configuration {
    crawler_configuration {
      role_arn = aws_iam_role.test.arn
    }

    provider_identity {
      external_id = "windows-sysmon-test"
      principal   = data.aws_caller_identity.current.account_id
    }
  }

  depends_on = [aws_securitylake_data_lake.test, aws_iam_role.test]
}

`)
}

func testAccSubscriberNotificationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSubscriberNotification_config(), fmt.Sprintf(`

resource "aws_securitylake_subscriber" "test" {
	subscriber_name        = %[1]q
	subscriber_description = "Example"
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
	
	depends_on = [aws_securitylake_custom_log_source.test]
}

resource "aws_securitylake_subscriber_notification" "test" {
	subscriber_id = aws_securitylake_subscriber.test.id
	configuration {
		sqs_notification_configuration {}
	}

	depends_on = [aws_securitylake_subscriber.test]
}
`, rName))
}

func testAccSubscriberNotificationConfig_https(rName string) string {
	return acctest.ConfigCompose(testAccSubscriberNotification_config(), fmt.Sprintf(`

resource "aws_securitylake_subscriber" "test" {
	subscriber_name        = %[1]q
	subscriber_description = "Example"
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
	
	depends_on = [aws_securitylake_custom_log_source.test]
}


resource "aws_securitylake_subscriber_notification" "test" {
	subscriber_id = aws_securitylake_subscriber.test.id
	configuration {
		https_notification_configuration {
			endpoint = "https://rqc0a4dz9k.execute-api.eu-west-1.amazonaws.com"
			target_role_arn = aws_iam_role.event_bridge.arn
		}
	}

	depends_on = [aws_securitylake_subscriber.test]
}
`, rName))
}
