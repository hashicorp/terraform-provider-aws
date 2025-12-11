// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationsignals_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationsignals"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationsignals/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfapplicationsignals "github.com/hashicorp/terraform-provider-aws/internal/service/applicationsignals"
)

func TestAccApplicationSignalsServiceLevelObjective_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var servicelevelobjective awstypes.ServiceLevelObjective
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationsignals_service_level_objective.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// TODO - work out why this precheck fails even though sdk can create SLOs...
			//acctest.PreCheckPartitionHasService(t, names.ApplicationSignalsServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationSignalsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLevelObjectiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLevelObjectiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceLevelObjectiveExists(ctx, resourceName, &servicelevelobjective),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "application-signals", regexache.MustCompile(`slo/`+rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccServiceLevelObjectiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName, // The attribute that uniquely identifies the resource
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccApplicationSignalsServiceLevelObjective_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var servicelevelobjective awstypes.ServiceLevelObjective
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationsignals_service_level_objective.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationSignalsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLevelObjectiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLevelObjectiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceLevelObjectiveExists(ctx, resourceName, &servicelevelobjective),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfapplicationsignals.ResourceServiceLevelObjective, resourceName),
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

func TestAccApplicationSignalsServiceLevelObjective_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var before, after awstypes.ServiceLevelObjective
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationsignals_service_level_objective.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationSignalsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLevelObjectiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLevelObjectiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceLevelObjectiveExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, fmt.Sprintf("%s service level objective", rName)),
				),
			},
			{
				Config: testAccServiceLevelObjectiveConfig_update(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceLevelObjectiveExists(ctx, resourceName, &after),
					testAccCheckServiceLevelObjectiveNotRecreated(&before, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, fmt.Sprintf("%s service level objective updated", rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccServiceLevelObjectiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName, // The attribute that uniquely identifies the resource
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})

}

func TestAccApplicationSignalsServiceLevelObjective_full(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var servicelevelobjective awstypes.ServiceLevelObjective
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_applicationsignals_service_level_objective.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ApplicationSignalsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceLevelObjectiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLevelObjectiveConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServiceLevelObjectiveExists(ctx, resourceName, &servicelevelobjective),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "application-signals", regexache.MustCompile(`slo/`+rName)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccServiceLevelObjectiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
				ImportStateVerifyIgnore:              []string{"apply_immediately", "user"},
			},
		},
	})
}

func testAccCheckServiceLevelObjectiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ApplicationSignalsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_applicationsignals_service_level_objective" {
				continue
			}

			_, err := tfapplicationsignals.FindServiceLevelObjectiveByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ApplicationSignals, create.ErrActionCheckingDestroyed, tfapplicationsignals.ResNameServiceLevelObjective, rs.Primary.ID, err)
			}

			return create.Error(names.ApplicationSignals, create.ErrActionCheckingDestroyed, tfapplicationsignals.ResNameServiceLevelObjective, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckServiceLevelObjectiveExists(ctx context.Context, name string, servicelevelobjective *awstypes.ServiceLevelObjective) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ApplicationSignals, create.ErrActionCheckingExistence, tfapplicationsignals.ResNameServiceLevelObjective, name, errors.New("not found"))
		}

		if rs.Primary.Attributes["name"] == "" {
			return create.Error(names.ApplicationSignals, create.ErrActionCheckingExistence, tfapplicationsignals.ResNameServiceLevelObjective, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ApplicationSignalsClient(ctx)

		resp, err := tfapplicationsignals.FindServiceLevelObjectiveByID(ctx, conn, rs.Primary.Attributes["name"])
		if err != nil {
			return create.Error(names.ApplicationSignals, create.ErrActionCheckingExistence, tfapplicationsignals.ResNameServiceLevelObjective, rs.Primary.Attributes["name"], err)
		}

		*servicelevelobjective = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ApplicationSignalsClient(ctx)

	input := &applicationsignals.ListServiceLevelObjectivesInput{}

	_, err := conn.ListServiceLevelObjectives(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckServiceLevelObjectiveNotRecreated(before, after *awstypes.ServiceLevelObjective) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Arn), aws.ToString(after.Arn); before != after {
			return create.Error(
				names.ApplicationSignals,
				create.ErrActionCheckingNotRecreated,
				tfapplicationsignals.ResNameServiceLevelObjective,
				before+after,
				errors.New(fmt.Sprintf("recreated (before ARN: %s, after ARN: %s)", before, after)))
		}

		return nil
	}
}

func testAccServiceLevelObjectiveImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		name, ok := rs.Primary.Attributes[names.AttrName]
		if !ok {
			return "", fmt.Errorf("Name attribute not found in state for resource: %s", resourceName)
		}
		return name, nil
	}
}

func testAccServiceLevelObjectiveConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_applicationsignals_service_level_objective" "test" {
  name = %[1]q
  description = "%[1]s service level objective"
  goal {
    interval {
      rolling_interval {
        duration_unit = "DAY"
        duration      = 90
      }
    }
    attainment_goal   = 99.98
    warning_threshold = 99.9
  }
  sli {
    sli_metric {
      metric_data_queries {
        id = "m1"
        expression = "FILL(METRICS(), 0)"
        period = 60
        return_data = true
      }
    }
     comparison_operator = "LessThan"
     metric_threshold    = 2
  }
}
`, rName)
}

func testAccServiceLevelObjectiveConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_applicationsignals_service_level_objective" "test" {
  name = %[1]q
  description = "%[1]s service level objective updated"
  goal {
    interval {
      rolling_interval {
        duration_unit = "DAY"
        duration      = 90
      }
    }
    attainment_goal   = 99.98
    warning_threshold = 99.9
  }
  sli {
    sli_metric {
      metric_data_queries {
        id = "m1"
        expression = "FILL(METRICS(), 0)"
        period = 60
        return_data = true
      }
    }
     comparison_operator = "LessThan"
     metric_threshold    = 2
  }
}
`, rName)
}

func testAccServiceLevelObjectiveConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_applicationsignals_service_level_objective" "test" {
  name = %[1]q

  burn_rate_configurations {
    look_back_window_minutes = 60
  }

  goal {
    interval {
      rolling_interval {
        duration_unit = "DAY"
        duration      = 109
      }
    }
    attainment_goal   = 99.98
    warning_threshold = 99.9
  }

  request_based_sli {
    request_based_sli_metric {
      total_request_count_metric {
        metric_stat {
          metric {
            namespace  = "AWS/Lambda"
            metric_name = "Invocations"
            dimensions {
              name  = "Dimension1"
              value = "my-dimension-name"
            }
          }
          period = 60
          stat = "Sum"
        }
        id = "total"
        return_data = true
      }
      monitored_request_count_metric {
        bad_count_metric {
          id = "cwMetricNumerator"
          metric_stat {
            metric {
              namespace  = "AWS/ApplicationELB"
              metric_name = "HTTPCode_Target_5XX_Count"
              dimensions {
                name  = "LoadBalancer"
                value = "my-load-balancer"
              }
            }
            period = 60
            stat   = "Sum"
          }
          return_data = true
        }
        bad_count_metric {
          id = "pop"
          metric_stat {
            metric {
              namespace  = "AWS/Lambda"
              metric_name = "Errors"
              dimensions {
                name  = "LoadBalancer"
                value = "another-load-balancer"
              }
            }
            period = 60
            stat   = "Sum"
          }
          return_data = false
        }
      }
    }
  }
}`, rName)
}
