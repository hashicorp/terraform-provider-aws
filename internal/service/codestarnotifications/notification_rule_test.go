// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codestarnotifications_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codestarnotifications"
	"github.com/aws/aws-sdk-go-v2/service/codestarnotifications/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodestarnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/codestarnotifications"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeStarNotificationsNotificationRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarNotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codestar-notifications", regexache.MustCompile("notificationrule/.+")),
					resource.TestCheckResourceAttr(resourceName, "detail_type", string(types.DetailTypeBasic)),
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.NotificationRuleStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
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

func TestAccCodeStarNotificationsNotificationRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarNotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodestarnotifications.ResourceNotificationRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeStarNotificationsNotificationRule_status(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarNotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleConfig_status(rName, string(types.NotificationRuleStatusDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.NotificationRuleStatusDisabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleConfig_status(rName, string(types.NotificationRuleStatusEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.NotificationRuleStatusEnabled)),
				),
			},
			{
				Config: testAccNotificationRuleConfig_status(rName, string(types.NotificationRuleStatusDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.NotificationRuleStatusDisabled)),
				),
			},
		},
	})
}

func TestAccCodeStarNotificationsNotificationRule_targets(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarNotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleConfig_targets1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleConfig_targets2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct2),
				),
			},
			{
				Config: testAccNotificationRuleConfig_targets1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccCodeStarNotificationsNotificationRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarNotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccNotificationRuleConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccCodeStarNotificationsNotificationRule_eventTypeIDs(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeStarNotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNotificationRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationRuleConfig_eventTypeIDs1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNotificationRuleConfig_eventTypeIDs2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", acctest.Ct2),
				),
			},
			{
				Config: testAccNotificationRuleConfig_eventTypeIDs3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNotificationRuleExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckNotificationRuleExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarNotificationsClient(ctx)

		_, err := tfcodestarnotifications.FindNotificationRuleByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckNotificationRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarNotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codestarnotifications_notification_rule" {
				continue
			}

			_, err := tfcodestarnotifications.FindNotificationRuleByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeStar Notification Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeStarNotificationsClient(ctx)

	input := &codestarnotifications.ListTargetsInput{
		MaxResults: aws.Int32(1),
	}

	_, err := conn.ListTargets(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccNotificationRuleConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccNotificationRuleConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccNotificationRuleConfig_status(rName, status string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = %[2]q

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName, status))
}

func testAccNotificationRuleConfig_targets1(rName string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccNotificationRuleConfig_targets2(rName string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test2" {
  name = "%[1]s-2"
}

resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn

  target {
    address = aws_sns_topic.test.arn
  }

  target {
    address = aws_sns_topic.test2.arn
  }
}
`, rName))
}

func testAccNotificationRuleConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    %[2]q = %[3]q
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccNotificationRuleConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccNotificationRuleConfig_eventTypeIDs1(rName string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-comments-on-commits",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccNotificationRuleConfig_eventTypeIDs2(rName string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-comments-on-commits",
    "codecommit-repository-pull-request-created",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccNotificationRuleConfig_eventTypeIDs3(rName string) string {
	return acctest.ConfigCompose(testAccNotificationRuleConfig_base(rName), fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-pull-request-created",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName))
}
