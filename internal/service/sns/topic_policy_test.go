// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSNSTopicPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, "aws_sns_topic.test", &attributes),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_sns_topic.test", names.AttrARN),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwner),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
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

func TestAccSNSTopicPolicy_updated(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, "aws_sns_topic.test", &attributes),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTopicPolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, "aws_sns_topic.test", &attributes),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile("SNS:DeleteTopic")),
				),
			},
		},
	})
}

func TestAccSNSTopicPolicy_Disappears_topic(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	topicResourceName := "aws_sns_topic.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, topicResourceName, &attributes),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsns.ResourceTopic(), topicResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSNSTopicPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, "aws_sns_topic.test", &attributes),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsns.ResourceTopicPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSNSTopicPolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicPolicyConfig_equivalent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, t, "aws_sns_topic.test", &attributes),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_sns_topic.test", names.AttrARN),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwner),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccTopicPolicyConfig_equivalent2(rName),
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

func testAccCheckTopicPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SNSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sns_topic_policy" {
				continue
			}

			_, err := tfsns.FindTopicAttributesByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SNS Topic Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTopicPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"default",
  "Statement":[
    {
      "Sid":"%[1]s",
      "Effect":"Allow",
      "Principal":{
        "AWS":"*"
      },
      "Action":[
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission"
      ],
      "Resource":"${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccTopicPolicyConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn    = aws_sns_topic.test.arn
  policy = <<POLICY
{
  "Version":"2012-10-17",
  "Id":"default",
  "Statement":[
    {
      "Sid":"%[1]s",
      "Effect":"Allow",
      "Principal":{
        "AWS":"*"
      },
      "Action":[
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic"
      ],
      "Resource":"${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccTopicPolicyConfig_equivalent(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action = [
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic",
      ]
      Resource = aws_sns_topic.test.arn
    }]
  })
}
`, rName)
}

func testAccTopicPolicyConfig_equivalent2(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_policy" "test" {
  arn = aws_sns_topic.test.arn
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "default"
    Statement = [{
      Sid    = %[1]q
      Effect = "Allow"
      Principal = {
        AWS = ["*"]
      }
      Action = [
        "SNS:SetTopicAttributes",
        "SNS:RemovePermission",
        "SNS:DeleteTopic",
        "SNS:AddPermission",
        "SNS:GetTopicAttributes",
      ]
      Resource = [aws_sns_topic.test.arn]
    }]
  })
}
`, rName)
}

func testAccCheckTopicPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SNS Topic ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).SNSClient(ctx)
		output, err := tfsns.FindTopicAttributesByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output[tfsns.TopicAttributeNamePolicy] == "" {
			return fmt.Errorf("Topic policy not found")
		}

		return nil
	}
}
