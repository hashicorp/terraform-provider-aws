// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfresiliencehub "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "resiliencehub", regexache.MustCompile(`app/.+`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccResilienceHubApp_terraformSource(t *testing.T) {
	ctx := acctest.Context(t)
	var v resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_terraformSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "resiliencehub", regexache.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "resource_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_mapping.0.mapping_type", "Terraform"),
				),
			},
		},
	})
}

func TestAccResilienceHubApp_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var v resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_complete(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Complete test app"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "resiliencehub", regexache.MustCompile(`app/.+`)),
					resource.TestCheckResourceAttr(resourceName, "app_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_template.0.version", "2.0"),
				),
			},
		},
	})
}

func TestAccResilienceHubApp_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v resiliencehub.DescribeAppOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_resiliencehub_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccAppConfig_updateTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "app_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "app_template.0.version", "2.0"),
				),
			},
			{
				Config: testAccAppConfig_updateResourceMapping(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "resource_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_mapping.0.mapping_type", "Terraform"),
				),
			},
		},
	})
}

func testAccCheckAppDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehub_app" {
				continue
			}

			_, err := tfresiliencehub.FindAppByARN(ctx, conn, rs.Primary.Attributes["arn"])
			if errs.IsA[*retry.NotFoundError](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("ResilienceHub App %s still exists", rs.Primary.Attributes["arn"])
		}

		return nil
	}
}

func testAccCheckAppExists(ctx context.Context, n string, v *resiliencehub.DescribeAppOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ResilienceHubClient(ctx)

		output, err := tfresiliencehub.FindAppByARN(ctx, conn, rs.Primary.Attributes["arn"])
		if err != nil {
			return err
		}

		*v = resiliencehub.DescribeAppOutput{App: output}

		return nil
	}
}

func testAccAppConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_app" "test" {
  name = %[1]q

  app_template {
    version = "2.0"

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }
  }
}
`, rName)
}

func testAccAppConfig_terraformSource(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "resiliencehub.amazonaws.com"
        }
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*"
        ]
      }
    ]
  })
}

resource "aws_s3_object" "tfstate" {
  bucket = aws_s3_bucket.test.bucket
  key    = "terraform.tfstate"
  content = jsonencode({
    version           = 4
    terraform_version = "1.0.0"
    serial            = 1
    lineage           = "test"
    outputs           = {}
    resources = [
      {
        mode     = "managed"
        type     = "aws_lambda_function"
        name     = "test"
        provider = "provider[\"registry.terraform.io/hashicorp/aws\"]"
        instances = [
          {
            schema_version = 0
            attributes = {
              function_name = "test-function"
              arn           = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:test-function"
            }
          }
        ]
      }
    ]
  })
}

resource "aws_resiliencehub_app" "test" {
  name                    = %[1]q
  description             = "Test app with S3 Terraform source"
  assessment_schedule = "Disabled"

  app_template {
    version = "2.0"

    resource {
      name = "lambda-function"
      type = "AWS::Lambda::Function"

      logical_resource_id {
        identifier            = "MyLambda"
        terraform_source_name = "my-terraform-source"
      }
    }

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }

    app_component {
      name           = "compute-tier"
      type           = "AWS::ResilienceHub::ComputeAppComponent"
      resource_names = ["lambda-function"]
    }
  }

  resource_mapping {
    mapping_type          = "Terraform"
    resource_name         = "lambda-function"
    terraform_source_name = "my-terraform-source"

    physical_resource_id {
      type       = "Native"
      identifier = "s3://${aws_s3_bucket.test.bucket}/terraform.tfstate"
    }
  }

  depends_on = [aws_s3_object.tfstate, aws_s3_bucket_policy.test]
}
`, rName)
}

func testAccAppConfig_complete(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_app" "test" {
  name                    = %[1]q
  description             = "Complete test app"
  assessment_schedule = "Daily"

  app_template {
    version = "2.0"

    resource {
      name = "lambda-function"
      type = "AWS::Lambda::Function"

      logical_resource_id {
        identifier         = "MyLambda"
        logical_stack_name = "my-stack"
      }
    }
    resource {
      name = "database"
      type = "AWS::RDS::DBInstance"

      logical_resource_id {
        identifier         = "MyDatabase"
        logical_stack_name = "my-stack"
      }
    }

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }
    app_component {
      name           = "compute-tier"
      type           = "AWS::ResilienceHub::ComputeAppComponent"
      resource_names = ["lambda-function"]
    }

    app_component {
      name           = "database-tier"
      type           = "AWS::ResilienceHub::DatabaseAppComponent"
      resource_names = ["database"]
    }
  }

  tags = {
    Environment = "test"
    Purpose     = "testing"
  }
}
`, rName)
}

func testAccAppConfig_updateTemplate(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_app" "test" {
  name                    = %[1]q
  description             = "Updated description"
  assessment_schedule = "Disabled"

  app_template {
    version = "2.0"

    resource {
      name = "updated-lambda"
      type = "AWS::Lambda::Function"

      logical_resource_id {
        identifier         = "UpdatedLambda"
        logical_stack_name = "updated-stack"
      }
    }

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }

    app_component {
      name           = "updated-compute-tier"
      type           = "AWS::ResilienceHub::ComputeAppComponent"
      resource_names = ["updated-lambda"]
    }
  }
}
`, rName)
}

func testAccAppConfig_updateResourceMapping(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehub_app" "test" {
  name                    = %[1]q
  description             = "Updated description"
  assessment_schedule = "Disabled"

  app_template {
    version = "2.0"

    resource {
      name = "updated-lambda"
      type = "AWS::Lambda::Function"

      logical_resource_id {
        identifier            = "UpdatedLambda"
        terraform_source_name = "updated-terraform-source"
      }
    }

    app_component {
      name           = "appcommon"
      type           = "AWS::ResilienceHub::AppCommonAppComponent"
      resource_names = []
    }

    app_component {
      name           = "updated-compute-tier"
      type           = "AWS::ResilienceHub::ComputeAppComponent"
      resource_names = ["updated-lambda"]
    }
  }

  resource_mapping {
    mapping_type          = "Terraform"
    resource_name         = "updated-lambda"
    terraform_source_name = "updated-terraform-source"

    physical_resource_id {
      type       = "Native"
      identifier = "s3://updated-terraform-bucket/terraform.tfstate"
    }
  }
}
`, rName)
}
