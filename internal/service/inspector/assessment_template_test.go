// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inspector_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/inspector"
	awstypes "github.com/aws/aws-sdk-go-v2/service/inspector/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfinspector "github.com/hashicorp/terraform-provider-aws/internal/service/inspector"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccInspectorAssessmentTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "inspector", regexache.MustCompile(`target/.+/template/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDuration, "3600"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "rules_package_arns.#", "data.aws_inspector_rules_packages.available", "arns.#"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &v),
					testAccCheckTemplateDisappears(ctx, &v),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &v),
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
				Config: testAccAssessmentTemplateConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAssessmentTemplateConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAssessmentTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccInspectorAssessmentTemplate_eventSubscription(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AssessmentTemplate
	resourceName := "aws_inspector_assessment_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	event1 := "ASSESSMENT_RUN_STARTED"
	event1Updated := "ASSESSMENT_RUN_COMPLETED"
	event2 := "ASSESSMENT_RUN_STATE_CHANGED"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.InspectorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssessmentTemplateConfig_eventSubscription(rName, event1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.#", acctest.Ct1),
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
					testAccCheckTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.#", acctest.Ct1),
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
					testAccCheckTemplateExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "event_subscription.#", acctest.Ct2),
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

func testAccCheckTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_inspector_assessment_template" {
				continue
			}

			_, err := tfinspector.FindAssessmentTemplateByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*retry.NotFoundError](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Inspector, create.ErrActionCheckingDestroyed, tfinspector.ResNameAssessmentTemplate, rs.Primary.ID, err)
			}

			return create.Error(names.Inspector, create.ErrActionCheckingDestroyed, tfinspector.ResNameAssessmentTemplate, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTemplateDisappears(ctx context.Context, v *awstypes.AssessmentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		_, err := conn.DeleteAssessmentTemplate(ctx, &inspector.DeleteAssessmentTemplateInput{
			AssessmentTemplateArn: v.Arn,
		})

		return err
	}
}

func testAccCheckTemplateExists(ctx context.Context, name string, v *awstypes.AssessmentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Inspector Classic Assessment template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).InspectorClient(ctx)

		resp, err := tfinspector.FindAssessmentTemplateByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Inspector, create.ErrActionCheckingExistence, tfinspector.ResNameAssessmentTemplate, rs.Primary.ID, err)
		}

		*v = *resp

		return nil
	}
}

func testAccTemplateAssessmentBase(rName string) string {
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
	return testAccTemplateAssessmentBase(rName) + fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns
}
`, rName)
}

func testAccAssessmentTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccTemplateAssessmentBase(rName) + fmt.Sprintf(`
resource "aws_inspector_assessment_template" "test" {
  name       = %[1]q
  target_arn = aws_inspector_assessment_target.test.arn
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.available.arns

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAssessmentTemplateConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccTemplateAssessmentBase(rName) + fmt.Sprintf(`
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
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAssessmentTemplateBase_eventSubscription(rName string) string {
	return acctest.ConfigCompose(
		testAccTemplateAssessmentBase(rName),
		fmt.Sprintf(`
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
		`, rName),
	)
}

func testAccAssessmentTemplateConfig_eventSubscription(rName, event string) string {
	return acctest.ConfigCompose(
		testAccAssessmentTemplateBase_eventSubscription(rName),
		fmt.Sprintf(`
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
		`, rName, event),
	)
}

func testAccAssessmentTemplateConfig_eventSubscriptionMultiple(rName, event1, event2 string) string {
	return acctest.ConfigCompose(
		testAccAssessmentTemplateBase_eventSubscription(rName),
		fmt.Sprintf(`
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
		`, rName, event1, event2),
	)
}
