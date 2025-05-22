// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/hashicorp/terraform-provider-aws/internal/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDynamoDBResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy dynamodb.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcepolicy),
					resource.TestMatchResourceAttr(resourceName, "revision_id", regexache.MustCompile(`\d+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"confirm_remove_self_resource_access", names.AttrPolicy},
			},
		},
	})
}

func TestAccDynamoDBResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var resourcepolicy dynamodb.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DynamoDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &resourcepolicy),
					resource.TestMatchResourceAttr(resourceName, "revision_id", regexache.MustCompile(`\d+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"confirm_remove_self_resource_access", names.AttrPolicy},
			},
			{
				Config: testAccResourcePolicyConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(`.*GetItem.*`)),
				),
			},
		},
	})
}

func TestAccDynamoDBResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var out dynamodb.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dynamodb_resource_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourcePolicyExists(ctx, resourceName, &out),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdynamodb.ResourceResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dynamodb_resource_policy" {
				continue
			}

			_, err := tfdynamodb.FindResourcePolicyByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Kinesis Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckResourcePolicyExists(ctx context.Context, n string, v *dynamodb.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBClient(ctx)

		output, err := tfdynamodb.FindResourcePolicyByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccResourcePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTableConfig_basic(rName), `
data "aws_caller_identity" "current" {}
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["dynamodb:*"]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    resources = [aws_dynamodb_table.test.arn, "${aws_dynamodb_table.test.arn}/*"]
  }
}

resource "aws_dynamodb_resource_policy" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  policy       = data.aws_iam_policy_document.test.json
}
`)
}

func testAccResourcePolicyConfig_update(rName string) string {
	return acctest.ConfigCompose(testAccTableConfig_basic(rName), `
data "aws_caller_identity" "current" {}
data "aws_iam_policy_document" "test" {
  statement {
    actions = ["dynamodb:GetItem"]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    resources = [aws_dynamodb_table.test.arn, "${aws_dynamodb_table.test.arn}/*"]
  }
}

resource "aws_dynamodb_resource_policy" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  policy       = data.aws_iam_policy_document.test.json
}
`)
}
