// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/observabilityadmin/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfobservabilityadmin "github.com/hashicorp/terraform-provider-aws/internal/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTelemetryPipelinePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

	input := observabilityadmin.ListTelemetryPipelinesInput{}
	_, err := conn.ListTelemetryPipelines(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRandomTelemetryPipelineName(t *testing.T) string {
	return fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(t, 20, sdkacctest.CharSetAlpha))
}

func TestAccObservabilityAdminTelemetryPipeline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pipeline awstypes.TelemetryPipeline
	rName := testAccRandomTelemetryPipelineName(t)
	resourceName := "aws_observabilityadmin_telemetry_pipeline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccTelemetryPipelinePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryPipelineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryPipelineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryPipelineExists(ctx, t, resourceName, &pipeline),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("observabilityadmin", regexache.MustCompile(`telemetry-pipeline/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"body": knownvalue.NotNull(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccObservabilityAdminTelemetryPipeline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var pipeline awstypes.TelemetryPipeline
	rName := testAccRandomTelemetryPipelineName(t)
	resourceName := "aws_observabilityadmin_telemetry_pipeline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccTelemetryPipelinePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryPipelineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryPipelineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryPipelineExists(ctx, t, resourceName, &pipeline),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfobservabilityadmin.ResourceTelemetryPipeline, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccObservabilityAdminTelemetryPipeline_configurationUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var pipeline awstypes.TelemetryPipeline
	rName := testAccRandomTelemetryPipelineName(t)
	resourceName := "aws_observabilityadmin_telemetry_pipeline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccTelemetryPipelinePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryPipelineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryPipelineConfig_vpcFlow(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryPipelineExists(ctx, t, resourceName, &pipeline),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"body": knownvalue.NotNull(),
						}),
					})),
				},
			},
			{
				Config: testAccTelemetryPipelineConfig_vpcFlowWithProcessor(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryPipelineExists(ctx, t, resourceName, &pipeline),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"body": knownvalue.NotNull(),
						}),
					})),
				},
			},
		},
	})
}

func TestAccObservabilityAdminTelemetryPipeline_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var pipeline awstypes.TelemetryPipeline
	rName := testAccRandomTelemetryPipelineName(t)
	resourceName := "aws_observabilityadmin_telemetry_pipeline.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccTelemetryPipelinePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ObservabilityAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTelemetryPipelineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTelemetryPipelineConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryPipelineExists(ctx, t, resourceName, &pipeline),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccTelemetryPipelineConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryPipelineExists(ctx, t, resourceName, &pipeline),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccTelemetryPipelineConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTelemetryPipelineExists(ctx, t, resourceName, &pipeline),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckTelemetryPipelineDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_observabilityadmin_telemetry_pipeline" {
				continue
			}

			_, err := tfobservabilityadmin.FindTelemetryPipelineByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Observability Admin Telemetry Pipeline %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckTelemetryPipelineExists(ctx context.Context, t *testing.T, n string, v *awstypes.TelemetryPipeline) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ObservabilityAdminClient(ctx)

		output, err := tfobservabilityadmin.FindTelemetryPipelineByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccTelemetryPipelineConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "observabilityadmin.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:logs:*:${data.aws_caller_identity.current.account_id}:*"
    }]
  })
}
`, rName)
}

func testAccTelemetryPipelineConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTelemetryPipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_pipeline" "test" {
  name = %[1]q

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.test.arn
            }
            log_event_metadata = {
              data_source_name = replace(%[1]q, "-", "_")
              data_source_type = "default"
            }
          }
        }
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccTelemetryPipelineConfig_vpcFlow(rName string) string {
	return acctest.ConfigCompose(testAccTelemetryPipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_pipeline" "test" {
  name = %[1]q

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.test.arn
            }
            log_event_metadata = {
              data_source_name = "amazon_vpc"
              data_source_type = "flow"
            }
          }
        }
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccTelemetryPipelineConfig_vpcFlowWithProcessor(rName string) string {
	return acctest.ConfigCompose(testAccTelemetryPipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_pipeline" "test" {
  name = %[1]q

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.test.arn
            }
            log_event_metadata = {
              data_source_name = "amazon_vpc"
              data_source_type = "flow"
            }
          }
        }
        processor = [{
          ocsf = {
            schema = {
              vpc_flow = null
            }
            version         = "1.5"
            mapping_version = "1.5.0"
          }
        }]
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccTelemetryPipelineConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccTelemetryPipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_pipeline" "test" {
  name = %[1]q

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.test.arn
            }
            log_event_metadata = {
              data_source_name = "aws_cloudtrail"
              data_source_type = "data"
            }
          }
        }
        processor = [{
          ocsf = {
            schema = {
              cloud_trail = null
            }
            version         = "1.5"
            mapping_version = "1.5.0"
          }
        }]
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, tag1Key, tag1Value))
}

func testAccTelemetryPipelineConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccTelemetryPipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_observabilityadmin_telemetry_pipeline" "test" {
  name = %[1]q

  configuration {
    body = yamlencode({
      pipeline = {
        source = {
          cloudwatch_logs = {
            aws = {
              sts_role_arn = aws_iam_role.test.arn
            }
            log_event_metadata = {
              data_source_name = "aws_cloudtrail"
              data_source_type = "data"
            }
          }
        }
        processor = [{
          ocsf = {
            schema = {
              cloud_trail = null
            }
            version         = "1.5"
            mapping_version = "1.5.0"
          }
        }]
        sink = [{
          cloudwatch_logs = {
            log_group = "@original"
          }
        }]
      }
    })
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}
