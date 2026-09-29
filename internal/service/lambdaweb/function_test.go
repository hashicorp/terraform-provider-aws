// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdaweb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflambdaweb "github.com/hashicorp/terraform-provider-aws/internal/service/lambdaweb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// testAccPreCheck skips acceptance tests in regions where the Lambda Web API
// is not yet available (the service is rolling out region by region pre-GA).
func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

	input := lambdaweb.ListWebFunctionsInput{}
	_, err := conn.ListWebFunctions(ctx, &input)

	// Regions where the Lambda Web API has not been rolled out yet respond
	// with an AccessDeniedException that the pre-GA SDK surfaces without an
	// error code, so match on the message as well.
	if err != nil && strings.Contains(err.Error(), "Unable to determine service/operation name to be authorized") {
		t.Skipf("skipping acceptance testing: Lambda Web API not available in this region: %s", err)
	}

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func TestAccLambdaWebFunction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var function lambdaweb.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
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
					// The service explains every state, including the healthy one,
					// and the attribute is what surfaces a Pending or Failed cause.
					resource.TestCheckResourceAttrWith(resourceName, "state_reason", func(v string) error {
						if v == "" {
							return errors.New("state_reason should not be empty")
						}
						return nil
					}),
					// Unset scaling and throttling are not reported by the API:
					// account-level defaults apply server-side, invisibly.
					resource.TestCheckNoResourceAttr(resourceName, "endpoint_config.0.scaling_config.max_environments"),
					resource.TestCheckNoResourceAttr(resourceName, "endpoint_config.0.throttle_config.rate_limit"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     rName,
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
			testAccPreCheck(ctx, t)
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

// TestAccLambdaWebFunction_endpointFollowsLatestRevision covers the inline
// endpoint's auto_deployment_mode default: without an explicit mode the API
// would create the endpoint Disabled and pin it to the first revision, so a
// revision_config change would publish a revision that never receives traffic.
func TestAccLambdaWebFunction_endpointFollowsLatestRevision(t *testing.T) {
	ctx := acctest.Context(t)
	var function lambdaweb.GetWebFunctionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdaweb_function.test"
	var initialRevisionID string

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFunctionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionConfig_revisionDescription(rName, "first"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
					resource.TestCheckResourceAttr(resourceName, "endpoint_config.0.auto_deployment_mode", string(awstypes.AutoDeploymentModeLatestRevision)),
					resource.TestCheckResourceAttrWith(resourceName, "latest_revision_id", func(v string) error {
						if v == "" {
							return errors.New("latest_revision_id is empty")
						}
						initialRevisionID = v
						return nil
					}),
					testAccCheckFunctionEndpointServesLatestRevision(ctx, t, resourceName),
				),
			},
			{
				Config: testAccFunctionConfig_revisionDescription(rName, "second"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFunctionExists(ctx, t, resourceName, &function),
					resource.TestCheckResourceAttr(resourceName, "endpoint_config.0.auto_deployment_mode", string(awstypes.AutoDeploymentModeLatestRevision)),
					resource.TestCheckResourceAttr(resourceName, "revision_config.0.description", "second"),
					resource.TestCheckResourceAttrWith(resourceName, "latest_revision_id", func(v string) error {
						if v == initialRevisionID {
							return fmt.Errorf("expected a new revision, still %s", v)
						}
						return nil
					}),
					testAccCheckFunctionEndpointServesLatestRevision(ctx, t, resourceName),
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
			testAccPreCheck(ctx, t)
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

// testAccCheckFunctionEndpointServesLatestRevision asserts that the inline
// endpoint routes all traffic to the function's latest_revision_id. Traffic
// moves asynchronously after a revision is published, so poll.
func testAccCheckFunctionEndpointServesLatestRevision(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		functionName := rs.Primary.Attributes["function_name"]
		endpointName := rs.Primary.Attributes["endpoint_config.0.endpoint_name"]
		want := rs.Primary.Attributes["latest_revision_id"]
		if want == "" {
			return errors.New("latest_revision_id is empty")
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaWebClient(ctx)

		var got []awstypes.RevisionWeight
		err := tfresource.WaitUntil(ctx, 5*time.Minute, func(ctx context.Context) (bool, error) {
			out, err := tflambdaweb.FindEndpointByName(ctx, conn, functionName, endpointName)
			if err != nil {
				return false, err
			}
			got = out.RevisionWeights
			return len(got) == 1 && aws.ToString(got[0].RevisionId) == want && got[0].Weight == 100, nil
		}, tfresource.WaitOpts{PollInterval: 10 * time.Second})
		if err != nil {
			return fmt.Errorf("endpoint %s of Lambda Web Function %s does not serve latest revision %s (revision weights: %+v): %w", endpointName, functionName, want, got, err)
		}

		return nil
	}
}

func testAccFunctionConfig_revisionDescription(rName, description string) string {
	return acctest.ConfigCompose(testAccFunctionConfig_base(rName), fmt.Sprintf(`
resource "aws_lambdaweb_function" "test" {
  depends_on = [aws_s3_bucket_policy.test, aws_s3_bucket_versioning.test, aws_iam_role_policy_attachment.test]

  function_name = %[1]q

  revision_config {
    description = %[2]q

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

  # auto_deployment_mode deliberately omitted: HomeRegion defaults to LatestRevision.
  endpoint_config {
    endpoint_name = "default"
    endpoint_type = "HomeRegion"
    auth_type     = "ApplicationManaged"
  }
}
`, rName, description))
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
resource "aws_cloudwatch_log_group" "test" {
  name              = "/aws/lambda/web/%[1]s-acctest"
  retention_in_days = 1
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AccountRoot"
        Effect    = "Allow"
        Principal = { AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root" }
        Action    = "kms:*"
        Resource  = "*"
      },
      {
        Sid       = "LambdaWebService"
        Effect    = "Allow"
        Principal = { Service = "lambda.amazonaws.com" }
        Action    = ["kms:Decrypt", "kms:GenerateDataKey", "kms:DescribeKey"]
        Resource  = "*"
      }
    ]
  })
}

data "aws_caller_identity" "current" {}

resource "aws_lambdaweb_function" "test" {
  depends_on = [aws_s3_bucket_policy.test, aws_s3_bucket_versioning.test, aws_iam_role_policy_attachment.test]

  function_name = %[1]q

  revision_config {
    description = "updated by acceptance test"
    kms_key_arn = aws_kms_key.test.arn

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

      telemetry_config {
        logging_config {
          log_group             = aws_cloudwatch_log_group.test.name
          application_log_level = "INFO"
          system_log_level      = "WARN"
        }
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
