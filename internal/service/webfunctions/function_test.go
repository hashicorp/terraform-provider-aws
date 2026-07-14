// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webfunctions_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/webfunctions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/webfunctions/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfwebfunctions "github.com/hashicorp/terraform-provider-aws/internal/service/webfunctions"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWebFunctionsFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var function webfunctions.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_webfunctions_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WebFunctions)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WebFunctionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`web-function/.+`)),
					resource.TestCheckResourceAttr(resourceName, "function_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FunctionStateActive)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     rName,
				ImportStateVerifyIgnore: []string{
					"revision_config",
					"endpoint_config",
				},
			},
		},
	})
}

func TestAccWebFunctionsFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var function webfunctions.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_webfunctions_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WebFunctions)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WebFunctionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfwebfunctions.ResourceFunction, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWebFunctionsFunction_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var function webfunctions.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_webfunctions_function.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.WebFunctions)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WebFunctionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     rName,
				ImportStateVerifyIgnore: []string{
					"revision_config",
					"endpoint_config",
				},
			},
			{
				Config: testAccFunctionConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccFunctionConfig_tags0(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapSizeExact(0)),
				},
			},
		},
	})
}

func testAccCheckFunctionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WebFunctionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_webfunctions_function" {
				continue
			}

			_, err := tfwebfunctions.FindFunctionByName(ctx, conn, rs.Primary.Attributes["function_name"])
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return smarterr.NewError(err)
			}

			return fmt.Errorf("Web Functions Function %s still exists", rs.Primary.Attributes["function_name"])
		}

		return nil
	}
}

func testAccCheckFunctionExists(ctx context.Context, t *testing.T, n string, v *webfunctions.GetWebFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WebFunctionsClient(ctx)

		out, err := tfwebfunctions.FindFunctionByName(ctx, conn, rs.Primary.Attributes["function_name"])
		if err != nil {
			return smarterr.NewError(err)
		}

		*v = *out

		return nil
	}
}

func testAccFunctionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_base(rName), fmt.Sprintf(`
resource "aws_webfunctions_function" "test" {
  function_name = %[1]q

  revision_config {
    build_config {
      runtime = "nodejs24.x"

      code_config {
        s3_object {
          bucket = aws_s3_object.test.bucket
          key    = aws_s3_object.test.key
        }
      }
    }

    service_config {
      execution_role_arn = aws_iam_role.test.arn
    }
  }

  endpoint_config {
    endpoint_name = "default"
    endpoint_type = "HomeRegion"
    auth_type     = "ApplicationManaged"
    regions       = [data.aws_region.current.region]
  }
}
`, rName))
}

func testAccFunctionConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "function.zip"
  source = "test-fixtures/function.zip"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}
`, rName)
}

func testAccFunctionConfig_tagsBase(rName, tagsBlock string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_base(rName), fmt.Sprintf(`
resource "aws_webfunctions_function" "test" {
  function_name = %[1]q

  revision_config {
    build_config {
      runtime = "nodejs24.x"

      code_config {
        s3_object {
          bucket = aws_s3_object.test.bucket
          key    = aws_s3_object.test.key
        }
      }
    }

    service_config {
      execution_role_arn = aws_iam_role.test.arn
    }
  }

  endpoint_config {
    endpoint_name = "default"
    endpoint_type = "HomeRegion"
    auth_type     = "ApplicationManaged"
    regions       = [data.aws_region.current.region]
  }

%[2]s
}
`, rName, tagsBlock))
}

func testAccFunctionConfig_tags0(rName string) string {
	return testAccFunctionConfig_tagsBase(rName, `  tags = {}`)
}

func testAccFunctionConfig_tags1(rName, key1, value1 string) string {
	return testAccFunctionConfig_tagsBase(rName, fmt.Sprintf(`
  tags = {
    %[1]q = %[2]q
  }
`, key1, value1))
}

func testAccFunctionConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return testAccFunctionConfig_tagsBase(rName, fmt.Sprintf(`
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
`, key1, value1, key2, value2))
}
