// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnotifications "github.com/hashicorp/terraform-provider-aws/internal/service/notifications"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNotificationsEventRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var eventrule notifications.GetEventRuleOutput
	rConfigName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEventPattern := "{\"detail\":{\"state\":{\"value\":[\"ALARM\"]}}}"
	rEventType := "CloudWatch Alarm State Change"
	rRegion1 := "us-east-1" //lintignore:AWSAT003
	rRegion2 := "us-west-2" //lintignore:AWSAT003
	rSource := "aws.cloudwatch"
	resourceName := "aws_notifications_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventRuleConfig_basic(rConfigName, rEventPattern, rEventType, rRegion1, rRegion2, rSource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventRuleExists(ctx, resourceName, &eventrule),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", rEventPattern),
					resource.TestCheckResourceAttr(resourceName, "event_type", rEventType),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", rRegion1),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", rRegion2),
					resource.TestCheckResourceAttr(resourceName, names.AttrSource, rSource),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "notifications", regexache.MustCompile(`configuration/.+/rule/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccNotificationsEventRule_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v notifications.GetEventRuleOutput
	rConfigName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEventPatternV1 := "{\"detail\":{\"state\":{\"value\":[\"ALARM\"]}}}"
	rEventPatternV2 := "{\"detail\":{\"state\":{\"value\":[\"OK\"]}}}"
	rEventType := "CloudWatch Alarm State Change"
	rRegion1 := "us-east-1" //lintignore:AWSAT003
	rRegion2 := "us-west-2" //lintignore:AWSAT003
	rRegion3 := "eu-west-1" //lintignore:AWSAT003
	rSource := "aws.cloudwatch"
	resourceName := "aws_notifications_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventRuleConfig_basic(rConfigName, rEventPatternV1, rEventType, rRegion1, rRegion2, rSource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventRuleExists(ctx, resourceName, &v),
				),
			},
			{
				Config: testAccEventRuleConfig_basic(rConfigName, rEventPatternV2, rEventType, rRegion1, rRegion3, rSource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", rEventPatternV2),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", rRegion1),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", rRegion3),
				),
			},
		},
	})
}

func TestAccNotificationsEventRule_recreate(t *testing.T) {
	ctx := acctest.Context(t)

	var v1, v2 notifications.GetEventRuleOutput
	rConfigName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEventPattern := "{\"detail\":{\"state\":{\"value\":[\"ALARM\"]}}}"
	rEventTypeV1 := "CloudWatch Alarm State Change"
	rEventTypeV2 := "CloudWatch Alarm Configuration Change"
	rRegion1 := "us-east-1" //lintignore:AWSAT003
	rRegion2 := "us-west-2" //lintignore:AWSAT003
	rSource := "aws.cloudwatch"
	resourceName := "aws_notifications_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventRuleConfig_basic(rConfigName, rEventPattern, rEventTypeV1, rRegion1, rRegion2, rSource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventRuleExists(ctx, resourceName, &v1),
				),
			},
			{
				Config: testAccEventRuleConfig_basic(rConfigName, rEventPattern, rEventTypeV2, rRegion1, rRegion2, rSource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventRuleExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "event_pattern", rEventPattern),
					resource.TestCheckResourceAttr(resourceName, "event_type", rEventTypeV2),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", rRegion1),
					resource.TestCheckTypeSetElemAttr(resourceName, "regions.*", rRegion2),
					resource.TestCheckResourceAttr(resourceName, names.AttrSource, rSource),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "notifications", regexache.MustCompile(`configuration/.+/rule/.+$`)),
				),
			},
		},
	})
}

func TestAccNotificationsEventRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var eventrule notifications.GetEventRuleOutput
	rConfigName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rEventPattern := "{\"detail\":{\"state\":{\"value\":[\"ALARM\"]}}}"
	rEventType := "CloudWatch Alarm State Change"
	rRegion1 := "us-east-1" //lintignore:AWSAT003
	rRegion2 := "us-west-2" //lintignore:AWSAT003
	rSource := "aws.cloudwatch"
	resourceName := "aws_notifications_event_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.NotificationsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NotificationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEventRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEventRuleConfig_basic(rConfigName, rEventPattern, rEventType, rRegion1, rRegion2, rSource),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEventRuleExists(ctx, resourceName, &eventrule),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfnotifications.ResourceEventRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckEventRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_notifications_event_rule" {
				continue
			}

			_, err := tfnotifications.FindEventRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("User Notifications Event Rule %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckEventRuleExists(ctx context.Context, n string, v *notifications.GetEventRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NotificationsClient(ctx)

		output, err := tfnotifications.FindEventRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEventRuleConfig_basic(rName, rEventPattern, rEventType, rRegion1, rRegion2, rSource string) string {
	return fmt.Sprintf(`
resource "aws_notifications_notification_configuration" "test" {
  name        = %[1]q
  description = "example"
}

resource "aws_notifications_event_rule" "test" {
  event_pattern                  = %[2]q
  event_type                     = %[3]q
  notification_configuration_arn = aws_notifications_notification_configuration.test.arn
  regions                        = [%[4]q, %[5]q]
  source                         = %[6]q
}
`, rName, rEventPattern, rEventType, rRegion1, rRegion2, rSource)
}
