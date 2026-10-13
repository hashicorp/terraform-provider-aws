// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eventbridgev2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridgev2"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfeventbridgev2 "github.com/hashicorp/terraform-provider-aws/internal/service/eventbridgev2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEventBridgeV2ResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_resource_policy.test"
	busResourceName := "aws_eventbridgev2_event_bus.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName, "events:PutEvents"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrResourceARN), busResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPolicy), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("revision_id"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrResourceARN),
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				// The service reorders the policy document's keys.
				ImportStateVerifyIgnore: []string{
					names.AttrPolicy,
				},
			},
		},
	})
}

func TestAccEventBridgeV2ResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName, "events:PutEvents"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfeventbridgev2.ResourceResourcePolicy, resourceName),
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

func TestAccEventBridgeV2ResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_resource_policy.test"
	revisionID := statecheck.CompareValue(compare.ValuesDiffer())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName, "events:PutEvents"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					revisionID.AddStateValue(resourceName, tfjsonpath.New("revision_id")),
				},
			},
			{
				Config: testAccResourcePolicyConfig_basic(rName, "events:PutRawEvents"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					revisionID.AddStateValue(resourceName, tfjsonpath.New("revision_id")),
				},
			},
		},
	})
}

func TestAccEventBridgeV2ResourcePolicy_crossAccountSubscriber(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_resource_policy.test"
	busResourceName := "aws_eventbridgev2_event_bus.test"
	subscriberResourceName := "aws_eventbridgev2_subscriber.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_crossAccountSubscriber(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(subscriberResourceName, tfjsonpath.New("event_bus_arn"), busResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(subscriberResourceName, tfjsonpath.New(names.AttrState), knownvalue.StringExact("RUNNING")),
				},
			},
		},
	})
}

func TestAccEventBridgeV2ResourcePolicy_ramSharedEventBus(t *testing.T) {
	ctx := acctest.Context(t)
	var v eventbridgev2.GetResourcePolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_eventbridgev2_resource_policy.test"
	busResourceName := "aws_eventbridgev2_event_bus.test"
	subscriberResourceName := "aws_eventbridgev2_subscriber.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EventBridgeV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_ramSharedEventBus(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, t, resourceName, &v),
					testAccCheckResourcePolicyRAMExists(ctx, t, busResourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(subscriberResourceName, tfjsonpath.New("event_bus_arn"), busResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(subscriberResourceName, tfjsonpath.New(names.AttrState), knownvalue.StringExact("RUNNING")),
				},
			},
		},
	})
}

// testAccCheckResourcePolicyRAMExists verifies that the AWS RAM managed policy is attached alongside the customer-managed policy.
func testAccCheckResourcePolicyRAMExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

		input := eventbridgev2.GetResourcePolicyInput{
			PolicyName:  aws.String("AWS_RAM"),
			ResourceArn: aws.String(rs.Primary.Attributes[names.AttrARN]),
		}
		_, err := conn.GetResourcePolicy(ctx, &input)

		return err
	}
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eventbridgev2_resource_policy" {
				continue
			}

			ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

			_, err := tfeventbridgev2.FindResourcePolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EventBridge V2 Resource Policy %s still exists", rs.Primary.Attributes[names.AttrResourceARN])
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, t *testing.T, n string, v *eventbridgev2.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := conns.NewResourceContext(ctx, "", "", "", rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).EventBridgeV2Client(ctx)

		output, err := tfeventbridgev2.FindResourcePolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrResourceARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName, action string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q
}

resource "aws_eventbridgev2_resource_policy" "test" {
  resource_arn = aws_eventbridgev2_event_bus.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowAccount"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Action   = %[2]q
      Resource = aws_eventbridgev2_event_bus.test.arn
    }]
  })
}
`, rName, action)
}

// testAccResourcePolicyConfig_ramSharedEventBus shares an event bus with the alternate account through AWS RAM, adds a
// customer-managed policy with an explicit deny that no RAM managed permission can express, and creates a subscriber on
// the event bus from the alternate account. Both accounts must be in the same AWS Organization with RAM sharing with
// AWS Organizations enabled.
func testAccResourcePolicyConfig_ramSharedEventBus(rName string) string {
	return acctest.ConfigCompose(testAccResourcePolicyConfig_alternateSubscriber(rName), fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q
}

resource "aws_ram_resource_share" "test" {
  name                      = %[1]q
  allow_external_principals = false
  permission_arns           = ["arn:${data.aws_partition.current.partition}:ram::aws:permission/AWSRAMEventBridgeEventBusV2FullAccess"]
}

resource "aws_ram_resource_association" "test" {
  resource_arn       = aws_eventbridgev2_event_bus.test.arn
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_ram_principal_association" "test" {
  principal          = data.aws_caller_identity.alternate.account_id
  resource_share_arn = aws_ram_resource_share.test.arn
}

resource "aws_eventbridgev2_resource_policy" "test" {
  resource_arn = aws_eventbridgev2_event_bus.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "DenyAccountToPublish"
      Effect = "Deny"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.alternate.account_id}:root"
      }
      Action = [
        "events:PutEvents",
        "events:PutRawEvents",
      ]
      Resource = aws_eventbridgev2_event_bus.test.arn
    }]
  })

  depends_on = [aws_ram_resource_association.test, aws_ram_principal_association.test]
}
`, rName))
}

// testAccResourcePolicyConfig_crossAccountSubscriber grants the alternate account access to an event bus with a
// customer-managed policy only, without AWS RAM, and creates a subscriber on the event bus from the alternate account.
func testAccResourcePolicyConfig_crossAccountSubscriber(rName string) string {
	return acctest.ConfigCompose(testAccResourcePolicyConfig_alternateSubscriber(rName), fmt.Sprintf(`
resource "aws_eventbridgev2_event_bus" "test" {
  name = %[1]q
}

resource "aws_eventbridgev2_resource_policy" "test" {
  resource_arn = aws_eventbridgev2_event_bus.test.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowAccountToSubscribe"
      Effect = "Allow"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.alternate.account_id}:root"
      }
      Action = [
        "events:CreateSubscriber",
        "events:DescribeEventBus",
      ]
      Resource = aws_eventbridgev2_event_bus.test.arn
    }]
  })
}
`, rName))
}

// testAccResourcePolicyConfig_alternateSubscriber creates a subscriber from the alternate account on the event bus
// "aws_eventbridgev2_event_bus.test", once "aws_eventbridgev2_resource_policy.test" grants access to it.
func testAccResourcePolicyConfig_alternateSubscriber(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

data "aws_partition" "current" {}

resource "aws_sqs_queue" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_iam_role" "test" {
  provider = "awsalternate"

  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "events.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  provider = "awsalternate"

  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "sqs:SendMessage"
      Effect   = "Allow"
      Resource = aws_sqs_queue.test.arn
    }]
  })
}

resource "aws_eventbridgev2_subscriber" "test" {
  provider = "awsalternate"

  name          = %[1]q
  event_bus_arn = aws_eventbridgev2_event_bus.test.arn

  filter_configuration {
    filter {
      scope   = "METADATA"
      pattern = jsonencode({ source = ["example"] })
    }
  }

  invoke_configuration {
    role_arn   = aws_iam_role.test.arn
    target_arn = aws_sqs_queue.test.arn
  }

  depends_on = [aws_iam_role_policy.test, aws_eventbridgev2_resource_policy.test]
}
`, rName))
}
