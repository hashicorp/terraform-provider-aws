// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3tables_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3tables"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfs3tables "github.com/hashicorp/terraform-provider-aws/internal/service/s3tables"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3TablesTablePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var tablepolicy s3tables.GetTablePolicyOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTablePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTablePolicyConfig_basic(rName, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTablePolicyExists(ctx, resourceName, &tablepolicy),
					resource.TestCheckResourceAttrSet(resourceName, "resource_policy"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, "aws_s3tables_table.test", names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNamespace, "aws_s3tables_table.test", names.AttrNamespace),
					resource.TestCheckResourceAttrPair(resourceName, "table_bucket_arn", "aws_s3tables_table.test", "table_bucket_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccTablePolicyImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"resource_policy"},
			},
		},
	})
}

func TestAccS3TablesTablePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var tablepolicy s3tables.GetTablePolicyOutput
	bucketName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	namespace := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_s3tables_table_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3TablesServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTablePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTablePolicyConfig_basic(rName, namespace, bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTablePolicyExists(ctx, resourceName, &tablepolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3tables.ResourceTablePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTablePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3tables_table_policy" {
				continue
			}

			_, err := tfs3tables.FindTablePolicy(ctx, conn,
				rs.Primary.Attributes["table_bucket_arn"],
				rs.Primary.Attributes[names.AttrNamespace],
				rs.Primary.Attributes[names.AttrName],
			)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.S3Tables, create.ErrActionCheckingDestroyed, tfs3tables.ResNameTablePolicy, rs.Primary.ID, err)
			}

			return create.Error(names.S3Tables, create.ErrActionCheckingDestroyed, tfs3tables.ResNameTablePolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTablePolicyExists(ctx context.Context, name string, tablepolicy *s3tables.GetTablePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTablePolicy, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["table_bucket_arn"] == "" || rs.Primary.Attributes[names.AttrNamespace] == "" || rs.Primary.Attributes[names.AttrName] == "" {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTablePolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3TablesClient(ctx)

		resp, err := tfs3tables.FindTablePolicy(ctx, conn,
			rs.Primary.Attributes["table_bucket_arn"],
			rs.Primary.Attributes[names.AttrNamespace],
			rs.Primary.Attributes[names.AttrName],
		)
		if err != nil {
			return create.Error(names.S3Tables, create.ErrActionCheckingExistence, tfs3tables.ResNameTablePolicy, rs.Primary.ID, err)
		}

		*tablepolicy = *resp

		return nil
	}
}

func testAccTablePolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		identifier := tfs3tables.TableIdentifier{
			TableBucketARN: rs.Primary.Attributes["table_bucket_arn"],
			Namespace:      rs.Primary.Attributes[names.AttrNamespace],
			Name:           rs.Primary.Attributes[names.AttrName],
		}

		return identifier.String(), nil
	}
}

func testAccTablePolicyConfig_basic(rName, namespace, bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3tables_table_policy" "test" {
  resource_policy  = data.aws_iam_policy_document.test.json
  name             = aws_s3tables_table.test.name
  namespace        = aws_s3tables_table.test.namespace
  table_bucket_arn = aws_s3tables_table.test.table_bucket_arn
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["s3tables:*"]
    principals {
      type        = "AWS"
      identifiers = [data.aws_caller_identity.current.account_id]
    }
    resources = [aws_s3tables_table.test.arn]
  }
}

resource "aws_s3tables_table" "test" {
  name             = %[1]q
  namespace        = aws_s3tables_namespace.test.namespace
  table_bucket_arn = aws_s3tables_namespace.test.table_bucket_arn
  format           = "ICEBERG"
}

resource "aws_s3tables_namespace" "test" {
  namespace        = %[2]q
  table_bucket_arn = aws_s3tables_table_bucket.test.arn

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = %[3]q
}

data "aws_caller_identity" "current" {}
`, rName, namespace, bucketName)
}
