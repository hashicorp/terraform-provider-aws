// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccDSLogSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_log_subscription.subscription"
	logGroupName := "ad-service-log-subscription-test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			// test create
			{
				Config: testAccLogSubscriptionConfig_basic(rName, domainName, logGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogSubscriptionExists(ctx, resourceName,
						logGroupName,
					),
				),
			},
			// test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckLogSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_log_subscription" {
				continue
			}

			res, err := conn.ListLogSubscriptionsWithContext(ctx, &directoryservice.ListLogSubscriptionsInput{
				DirectoryId: aws.String(rs.Primary.ID),
			})

			if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
				continue
			}

			if err != nil {
				return err
			}

			if len(res.LogSubscriptions) > 0 {
				return fmt.Errorf("Expected AWS Directory Service Log Subscription to be gone, but was still found")
			}
		}

		return nil
	}
}

func testAccCheckLogSubscriptionExists(ctx context.Context, name string, logGroupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn(ctx)

		res, err := conn.ListLogSubscriptionsWithContext(ctx, &directoryservice.ListLogSubscriptionsInput{
			DirectoryId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(res.LogSubscriptions) == 0 {
			return fmt.Errorf("No Log subscription found")
		}

		if *(res.LogSubscriptions[0].LogGroupName) != logGroupName {
			return fmt.Errorf("Expected Log subscription not found")
		}

		return nil
	}
}

func testAccLogSubscriptionConfig_basic(rName, domain, logGroupName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_log_subscription" "subscription" {
  directory_id   = aws_directory_service_directory.test.id
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    Name = "terraform-testacc-directory-service-log-subscription"
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[2]q
  retention_in_days = 1
}

data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }

    resources = ["${aws_cloudwatch_log_group.test.arn}:*"]

    effect = "Allow"
  }
}

resource "aws_cloudwatch_log_resource_policy" "ad-log-policy" {
  policy_document = data.aws_iam_policy_document.ad-log-policy.json
  policy_name     = "ad-log-policy"
}
`, domain, logGroupName),
	)
}
