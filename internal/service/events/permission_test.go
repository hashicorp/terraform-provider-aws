// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfevents "github.com/hashicorp/terraform-provider-aws/internal/service/events"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventsPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	principal1 := "111111111111"
	principal2 := "*"
	statementID := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccPermissionConfig_basic("", statementID),
				ExpectError: regexache.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccPermissionConfig_basic(".", statementID),
				ExpectError: regexache.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccPermissionConfig_basic("12345678901", statementID),
				ExpectError: regexache.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccPermissionConfig_basic("abcdefghijkl", statementID),
				ExpectError: regexache.MustCompile(`must be \* or a 12 digit AWS account ID`),
			},
			{
				Config:      testAccPermissionConfig_basic(principal1, ""),
				ExpectError: regexache.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccPermissionConfig_basic(principal1, sdkacctest.RandString(65)),
				ExpectError: regexache.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccPermissionConfig_basic(principal1, " "),
				ExpectError: regexache.MustCompile(`must be one or more alphanumeric, hyphen, or underscore characters`),
			},
			{
				Config: testAccPermissionConfig_basic(principal1, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "events:PutEvents"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, principal1),
					resource.TestCheckResourceAttr(resourceName, "statement_id", statementID),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", tfevents.DefaultEventBusName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccPermissionConfig_basic(principal2, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, principal2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPermissionConfig_defaultBusName(principal2, statementID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccEventsPermission_eventBusName(t *testing.T) {
	ctx := acctest.Context(t)
	principal1 := "111111111111"
	statementID := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	busName := acctest.RandomWithPrefix(t, "tf-acc-test-bus")

	resourceName := "aws_cloudwatch_event_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_eventBusName(principal1, busName, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "events:PutEvents"),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPrincipal, principal1),
					resource.TestCheckResourceAttr(resourceName, "statement_id", statementID),
					resource.TestCheckResourceAttr(resourceName, "event_bus_name", busName),
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

func TestAccEventsPermission_action(t *testing.T) {
	ctx := acctest.Context(t)
	principal := "111111111111"
	statementID := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccPermissionConfig_action("", principal, statementID),
				ExpectError: regexache.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccPermissionConfig_action(sdkacctest.RandString(65), principal, statementID),
				ExpectError: regexache.MustCompile(`must be between 1 and 64 characters`),
			},
			{
				Config:      testAccPermissionConfig_action("events:", principal, statementID),
				ExpectError: regexache.MustCompile(`must be: events: followed by one or more alphabetic characters`),
			},
			{
				Config:      testAccPermissionConfig_action("events:1", principal, statementID),
				ExpectError: regexache.MustCompile(`must be: events: followed by one or more alphabetic characters`),
			},
			{
				Config: testAccPermissionConfig_action("events:PutEvents", principal, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction, "events:PutEvents"),
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

func TestAccEventsPermission_condition(t *testing.T) {
	ctx := acctest.Context(t)
	statementID := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_event_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_conditionOrganization(statementID, "o-1234567890"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.key", "aws:PrincipalOrgID"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.type", "StringEquals"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.value", "o-1234567890"),
				),
			},
			{
				Config: testAccPermissionConfig_conditionOrganization(statementID, "o-0123456789"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "condition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.key", "aws:PrincipalOrgID"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.type", "StringEquals"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.value", "o-0123456789"),
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

func TestAccEventsPermission_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	principal1 := "111111111111"
	principal2 := "222222222222"
	statementID1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	statementID2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_cloudwatch_event_permission.test"
	resourceName2 := "aws_cloudwatch_event_permission.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(principal1, statementID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName1),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPrincipal, principal1),
					resource.TestCheckResourceAttr(resourceName1, "statement_id", statementID1),
				),
			},
			{
				Config: testAccPermissionConfig_multiple(principal1, statementID1, principal2, statementID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName1),
					testAccCheckPermissionExists(ctx, t, resourceName2),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPrincipal, principal1),
					resource.TestCheckResourceAttr(resourceName1, "statement_id", statementID1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPrincipal, principal2),
					resource.TestCheckResourceAttr(resourceName2, "statement_id", statementID2),
				),
			},
		},
	})
}

func TestAccEventsPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_cloudwatch_event_permission.test"
	principal := "111111111111"
	statementID := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EventsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_basic(principal, statementID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfevents.ResourcePermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPermissionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		_, err := tfevents.FindPermissionByTwoPartKey(ctx, conn, rs.Primary.Attributes["event_bus_name"], rs.Primary.Attributes["statement_id"])

		return err
	}
}

func testAccCheckPermissionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).EventsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_event_permission" {
				continue
			}

			_, err := tfevents.FindPermissionByTwoPartKey(ctx, conn, rs.Primary.Attributes["event_bus_name"], rs.Primary.Attributes["statement_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge Permission %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPermissionConfig_basic(principal, statementID string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test" {
  principal    = "%[1]s"
  statement_id = "%[2]s"
}
`, principal, statementID)
}

func testAccPermissionConfig_defaultBusName(principal, statementID string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test" {
  principal      = %[1]q
  statement_id   = %[2]q
  event_bus_name = "default"
}
`, principal, statementID)
}

func testAccPermissionConfig_eventBusName(principal, busName, statementID string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test" {
  principal      = %[1]q
  statement_id   = %[2]q
  event_bus_name = aws_cloudwatch_event_bus.test.name
}

resource "aws_cloudwatch_event_bus" "test" {
  name = %[3]q
}
`, principal, statementID, busName)
}

func testAccPermissionConfig_action(action, principal, statementID string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test" {
  action       = "%[1]s"
  principal    = "%[2]s"
  statement_id = "%[3]s"
}
`, action, principal, statementID)
}

func testAccPermissionConfig_conditionOrganization(statementID, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test" {
  principal    = "*"
  statement_id = %q

  condition {
    key   = "aws:PrincipalOrgID"
    type  = "StringEquals"
    value = %q
  }
}
`, statementID, value)
}

func testAccPermissionConfig_multiple(principal1, statementID1, principal2, statementID2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_permission" "test" {
  principal    = "%[1]s"
  statement_id = "%[2]s"
}

resource "aws_cloudwatch_event_permission" "test2" {
  principal    = "%[3]s"
  statement_id = "%[4]s"
}
`, principal1, statementID1, principal2, statementID2)
}
