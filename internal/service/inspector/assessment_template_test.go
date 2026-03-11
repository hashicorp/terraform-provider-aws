// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package inspector_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfinspector "github.com/hashicorp/terraform-provider-aws/internal/service/inspector"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspectorAssessmentTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "inspector", regexache.MustCompile(`target/.+/template/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDuration, "3600"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "rules_package_arns.#", "data.aws_inspector_rules_packages.available", "arns.#"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTargetARN, "aws_inspector_assessment_target.test", names.AttrARN),
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

func TestAccInspectorAssessmentTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfinspector.ResourceAssessmentTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccInspectorAssessmentTemplate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssessmentTemplateConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAssessmentTemplateConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAssessmentTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccInspectorAssessmentTemplate_eventSubscription(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	event1 := "ASSESSMENT_RUN_STARTED"
	event1Updated := "ASSESSMENT_RUN_COMPLETED"
	event2 := "ASSESSMENT_RUN_STATE_CHANGED"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssessmentTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_eventSubscription(rName, event1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.0.event", event1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssessmentTemplateConfig_eventSubscription(rName, event1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.0.event", event1Updated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAssessmentTemplateConfig_eventSubscriptionMultiple(rName, event1, event2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssessmentTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.#", "2"),
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

func testAccCheckAssessmentTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).InspectorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector_assessment_template" {
				continue
			}

			_, err := tfinspector.FindAssessmentTemplateByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Inspector Classic Assessment Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAssessmentTemplateExists(ctx context.Context, t *testing.T, n string, v *awstypes.AssessmentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).InspectorClient(ctx)

		output, err := tfinspector.FindAssessmentTemplateByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAssessmentTemplateConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_inspector_rules_packages" "available" {}

resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_inspector_assessment_target" "test" {
  name               = %[1]q
  resource_group_arn = aws_inspector_resource_group.test.arn
}
`, rName)
}

func testAccAssessmentTemplateConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAssessmentTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns
}
`, rName))
}

func testAccAssessmentTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccAssessmentTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAssessmentTemplateConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAssessmentTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccAssessmentTemplateConfig_baseEventSubscription(rName string) string {
	return acctest.ConfigCompose(testAccAssessmentTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = ["*"]
    }

    actions = [
      "SNS:Publish"
    ]

    resources = [aws_sns_topic.test.arn]
  }
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = data.aws_iam_policy_document.test.json
}
`, rName))
}

func testAccAssessmentTemplateConfig_eventSubscription(rName, event string) string {
	return acctest.ConfigCompose(testAccAssessmentTemplateConfig_baseEventSubscription(rName), fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns

  event_subscription {
    event     = %[2]q
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName, event))
}

func testAccAssessmentTemplateConfig_eventSubscriptionMultiple(rName, event1, event2 string) string {
	return acctest.ConfigCompose(testAccAssessmentTemplateConfig_baseEventSubscription(rName), fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns

  event_subscription {
    event     = %[2]q
    topic_arn = aws_sns_topic.test.arn
  }

  event_subscription {
    event     = %[3]q
    topic_arn = aws_sns_topic.test.arn
  }
}
`, rName, event1, event2))
}
