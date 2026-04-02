// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSLogSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_log_subscription.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogSubscriptionConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogSubscriptionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrLogGroupName, rName),
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

func TestAccDSLogSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_log_subscription.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogSubscriptionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLogSubscriptionConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogSubscriptionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfds.ResourceLogSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLogSubscriptionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_log_subscription" {
				continue
			}

			_, err := tfds.FindLogSubscriptionByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service Log Subscription %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLogSubscriptionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		_, err := tfds.FindLogSubscriptionByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccLogSubscriptionConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_directory_service_log_subscription" "test" {
  directory_id   = aws_directory_service_directory.test.id
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_directory_service_directory" "test" {
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = 1
}

data "aws_iam_policy_document" "test" {
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

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_document = data.aws_iam_policy_document.test.json
  policy_name     = %[1]q
}
`, rName, domain))
}
