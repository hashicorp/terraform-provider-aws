// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaWebFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var function lambdaweb.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
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

func TestAccLambdaWebFunction_update(t *testing.T) {
	ctx := acctest.Context(t)
	var function lambdaweb.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function.test"
	var initialRevisionID string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
					resource.TestCheckResourceAttrWith(resourceName, "latest_revision_id", func(v string) error {
						if v == "" {
							return errors.New("latest_revision_id is empty")
						}
						initialRevisionID = v
						return nil
					}),
				),
			},
			{
				Config: testAccFunctionConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FunctionStateActive)),
					// A revision_config change must roll a new immutable revision.
					resource.TestCheckResourceAttrWith(resourceName, "latest_revision_id", func(v string) error {
						if v == "" {
							return errors.New("latest_revision_id is empty")
						}
						if v == initialRevisionID {
							return fmt.Errorf("expected a new revision, still %s", v)
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "revision_config.0.description", "updated by acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "revision_config.0.service_config.0.timeout_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "revision_config.0.service_config.0.max_concurrency_per_environment", "10"),
					resource.TestCheckResourceAttr(resourceName, "revision_config.0.service_config.0.environment_variables.APP_ENV", "acctest"),
				),
			},
		},
	})
}

func TestAccLambdaWebFunction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var function lambdaweb.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdaweb.ResourceFunction, resourceName),
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

func testAccCheckFunctionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdaweb_function" {
				continue
			}

			_, err := tflambdaweb.FindFunctionByName(ctx, conn, rs.Primary.Attributes["function_name"])
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

func testAccCheckFunctionExists(ctx context.Context, t *testing.T, n string, v *lambdaweb.GetWebFunctionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		out, err := tflambdaweb.FindFunctionByName(ctx, conn, rs.Primary.Attributes["function_name"])
		if err != nil {
			return smarterr.NewError(err)
		}

		*v = *out

		return nil
	}
}

func testAccFunctionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_base(rName), fmt.Sprintf(`
resource "aws_lambdaweb_function" "test" {
  depends_on = [aws_s3_bucket_policy.test, aws_s3_bucket_versioning.test, aws_iam_role_policy_attachment.test]

  function_name = %[1]q

  revision_config {
    build_config {
      runtime_config {
        runtime = "nodejs24.x"
      }

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

func testAccFunctionConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_base(rName), fmt.Sprintf(`
resource "aws_lambdaweb_function" "test" {
  depends_on = [aws_s3_bucket_policy.test, aws_s3_bucket_versioning.test, aws_iam_role_policy_attachment.test]

  function_name = %[1]q

  revision_config {
    description = "updated by acceptance test"

    build_config {
      runtime_config {
        runtime = "nodejs24.x"
      }

      code_config {
        s3_object {
          bucket = aws_s3_object.test.bucket
          key    = aws_s3_object.test.key
        }
      }
    }

    service_config {
      execution_role_arn              = aws_iam_role.test.arn
      timeout_seconds                 = 60
      max_concurrency_per_environment = 10

      environment_variables = {
        APP_ENV = "acctest"
      }
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

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "AllowLambdaAccess"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      Action   = ["s3:GetObject", "s3:GetObjectVersion"]
      Resource = "${aws_s3_bucket.test.arn}/*"
    }]
  })
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket_versioning.test.bucket
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
resource "aws_lambdaweb_function" "test" {
  depends_on = [aws_s3_bucket_policy.test, aws_s3_bucket_versioning.test, aws_iam_role_policy_attachment.test]

  function_name = %[1]q

  revision_config {
    build_config {
      runtime_config {
        runtime = "nodejs24.x"
      }

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
