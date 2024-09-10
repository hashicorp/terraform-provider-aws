// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/computeoptimizer/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcomputeoptimizer "github.com/hashicorp/terraform-provider-aws/internal/service/computeoptimizer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRecommendationPreferences_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RecommendationPreferencesDetail
	resourceName := "aws_computeoptimizer_recommendation_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComputeOptimizerEndpointID)
			testAccPreCheckEnrollmentStatus(ctx, t, "Active")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecommendationPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecommendationPreferencesConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecommendationPreferencesExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enhanced_infrastructure_metrics"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("external_metrics_preference"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("inferred_workload_types"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("look_back_period"), knownvalue.StringExact("DAYS_32")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("preferred_resource"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceType), knownvalue.StringExact("Ec2Instance")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("savings_estimation_mode"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact("AccountId"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("utilization_preference"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRecommendationPreferences_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RecommendationPreferencesDetail
	resourceName := "aws_computeoptimizer_recommendation_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComputeOptimizerEndpointID)
			testAccPreCheckEnrollmentStatus(ctx, t, "Active")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecommendationPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecommendationPreferencesConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecommendationPreferencesExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcomputeoptimizer.ResourceRecommendationPreferences, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRecommendationPreferences_preferredResources(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RecommendationPreferencesDetail
	resourceName := "aws_computeoptimizer_recommendation_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComputeOptimizerEndpointID)
			testAccPreCheckEnrollmentStatus(ctx, t, "Active")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecommendationPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecommendationPreferencesConfig_preferredResources,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecommendationPreferencesExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enhanced_infrastructure_metrics"), knownvalue.StringExact("Active")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("external_metrics_preference"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("external_metrics_preference"), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrSource: knownvalue.StringExact("Datadog"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("inferred_workload_types"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("look_back_period"), knownvalue.StringExact("DAYS_93")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("preferred_resource"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("preferred_resource"), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"exclude_list": knownvalue.Null(),
								"include_list": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("m5.xlarge"),
									knownvalue.StringExact("r5"),
								}),
								names.AttrName: knownvalue.StringExact("Ec2InstanceTypes"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceType), knownvalue.StringExact("Ec2Instance")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("savings_estimation_mode"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact("AccountId"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("utilization_preference"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRecommendationPreferences_utilizationPreferences(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RecommendationPreferencesDetail
	resourceName := "aws_computeoptimizer_recommendation_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ComputeOptimizerEndpointID)
			testAccPreCheckEnrollmentStatus(ctx, t, "Active")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ComputeOptimizerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRecommendationPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRecommendationPreferencesConfig_utilizationPreferences,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecommendationPreferencesExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enhanced_infrastructure_metrics"), knownvalue.StringExact("Active")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("external_metrics_preference"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("inferred_workload_types"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("look_back_period"), knownvalue.StringExact("DAYS_93")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("preferred_resource"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceType), knownvalue.StringExact("Ec2Instance")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("savings_estimation_mode"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact("AccountId"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("utilization_preference"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("utilization_preference"), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrMetricName: knownvalue.StringExact("CpuUtilization"),
								"metric_parameters": knownvalue.ListExact(
									[]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"headroom":  knownvalue.StringExact("PERCENT_20"),
											"threshold": knownvalue.StringExact("P95"),
										}),
									},
								),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrMetricName: knownvalue.StringExact("MemoryUtilization"),
								"metric_parameters": knownvalue.ListExact(
									[]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"headroom":  knownvalue.StringExact("PERCENT_30"),
											"threshold": knownvalue.Null(),
										}),
									},
								),
							}),
						}),
					),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRecommendationPreferencesConfig_utilizationPreferencesUpdated,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRecommendationPreferencesExists(ctx, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("enhanced_infrastructure_metrics"), knownvalue.StringExact("Inactive")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("external_metrics_preference"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("inferred_workload_types"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("look_back_period"), knownvalue.StringExact("DAYS_14")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("preferred_resource"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceType), knownvalue.StringExact("Ec2Instance")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("savings_estimation_mode"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrScope), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectPartial(map[string]knownvalue.Check{
								names.AttrName: knownvalue.StringExact("AccountId"),
							}),
						}),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("utilization_preference"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("utilization_preference"), knownvalue.ListExact(
						[]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrMetricName: knownvalue.StringExact("CpuUtilization"),
								"metric_parameters": knownvalue.ListExact(
									[]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"headroom":  knownvalue.StringExact("PERCENT_0"),
											"threshold": knownvalue.StringExact("P90"),
										}),
									},
								),
							}),
						}),
					),
				},
			},
		},
	})
}

func testAccCheckRecommendationPreferencesExists(ctx context.Context, n string, v *awstypes.RecommendationPreferencesDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)

		output, err := tfcomputeoptimizer.FindRecommendationPreferencesByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrResourceType], rs.Primary.Attributes["scope.0.name"], rs.Primary.Attributes["scope.0.value"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRecommendationPreferencesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ComputeOptimizerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_computeoptimizer_recommendation_preferences" {
				continue
			}

			_, err := tfcomputeoptimizer.FindRecommendationPreferencesByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrResourceType], rs.Primary.Attributes["scope.0.name"], rs.Primary.Attributes["scope.0.value"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Compute Optimizer Recommendation Preferences %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccRecommendationPreferencesConfig_basic = `
data "aws_caller_identity" "current" {}

resource "aws_computeoptimizer_recommendation_preferences" "test" {
  resource_type = "Ec2Instance"
  scope {
    name  = "AccountId"
    value = data.aws_caller_identity.current.account_id
  }

  look_back_period = "DAYS_32"
}
`

const testAccRecommendationPreferencesConfig_preferredResources = `
data "aws_caller_identity" "current" {}

resource "aws_computeoptimizer_recommendation_preferences" "test" {
  resource_type = "Ec2Instance"
  scope {
    name  = "AccountId"
    value = data.aws_caller_identity.current.account_id
  }

  enhanced_infrastructure_metrics = "Active"

  external_metrics_preference {
    source = "Datadog"
  }

  preferred_resource {
    include_list = ["m5.xlarge", "r5"]
    name         = "Ec2InstanceTypes"
  }
}
`

const testAccRecommendationPreferencesConfig_utilizationPreferences = `
data "aws_caller_identity" "current" {}

resource "aws_computeoptimizer_recommendation_preferences" "test" {
  resource_type = "Ec2Instance"
  scope {
    name  = "AccountId"
    value = data.aws_caller_identity.current.account_id
  }

  enhanced_infrastructure_metrics = "Active"

  utilization_preference {
    metric_name = "CpuUtilization"
    metric_parameters {
      headroom  = "PERCENT_20"
      threshold = "P95"
    }
  }

  utilization_preference {
    metric_name = "MemoryUtilization"
    metric_parameters {
      headroom = "PERCENT_30"
    }
  }
}
`

const testAccRecommendationPreferencesConfig_utilizationPreferencesUpdated = `
data "aws_caller_identity" "current" {}

resource "aws_computeoptimizer_recommendation_preferences" "test" {
  resource_type = "Ec2Instance"
  scope {
    name  = "AccountId"
    value = data.aws_caller_identity.current.account_id
  }

  enhanced_infrastructure_metrics = "Inactive"

  utilization_preference {
    metric_name = "CpuUtilization"
    metric_parameters {
      headroom  = "PERCENT_0"
      threshold = "P90"
    }
  }
}
`
