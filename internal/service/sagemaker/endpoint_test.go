// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

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
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("endpoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_config_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccSageMakerEndpoint_endpointConfigName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"
	sagemakerEndpointConfigurationResourceName1 := "aws_sagemaker_endpoint_configuration.test"
	sagemakerEndpointConfigurationResourceName2 := "aws_sagemaker_endpoint_configuration.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName1, names.AttrName),
				),
			},
			{
				Config: testAccEndpointConfig_endpointConfigNameUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName2, names.AttrName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
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

func TestAccSageMakerEndpoint_retainAllVariantProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "retain_all_variant_properties", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "exclude_retained_variant_properties.#", "0"),
				),
			},
			{
				Config: testAccEndpointConfig_retainAllVariantProperties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "retain_all_variant_properties", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "exclude_retained_variant_properties.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "exclude_retained_variant_properties.*", map[string]string{
						"variant_property_type": "DesiredInstanceCount",
					}),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are update-only arguments not returned by the API.
				ImportStateVerifyIgnore: []string{"retain_all_variant_properties", "retain_deployment_config", "exclude_retained_variant_properties"},
			},
		},
	})
}

func TestAccSageMakerEndpoint_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},
			{
				Config: testAccEndpointConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.bar", "baz"),
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

func TestAccSageMakerEndpoint_deploymentConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_deploymentBasic(rName, "ALL_AT_ONCE", 60),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.termination_wait_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.type", "ALL_AT_ONCE"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.wait_interval_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.canary_size.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.#", "0"),
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

func TestAccSageMakerEndpoint_deploymentConfig_blueGreen(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_deploymentBlueGreen(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.termination_wait_in_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.type", "LINEAR"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.wait_interval_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.canary_size.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.0.type", "INSTANCE_COUNT"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.0.traffic_routing_configuration.0.linear_step_size.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.#", "0"),
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

func TestAccSageMakerEndpoint_deploymentConfig_rolling(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_deploymentRolling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.0.alarms.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.blue_green_update_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.wait_interval_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.maximum_batch_size.0.type", "CAPACITY_PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.maximum_batch_size.0.value", "5"),
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

func TestAccSageMakerEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_sagemaker_endpoint.test", plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_sagemaker_endpoint.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// TestAccSageMakerEndpoint_changeEndpointConfig_sameName tests for an Endpoint Configuration being replaced due to a change,
// but since the Endpoint Configuration name doesn't change, the Endpoint does not pick up the change.
// NOTE: This is *not* what users want to happen!
func TestAccSageMakerEndpoint_changeEndpointConfig_sameName(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sagemaker_endpoint.test"
	sagemakerEndpointConfigurationResourceName := "aws_sagemaker_endpoint_configuration.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	variantName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	variantName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_changeEndpointConfig_variantName(rName, variantName1, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName, names.AttrName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.variant_name", variantName1),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.initial_instance_count", "2"),
				),
			},
			{
				Config: testAccEndpointConfig_changeEndpointConfig_variantName(rName, variantName2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName, names.AttrName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.variant_name", variantName2),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.initial_instance_count", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(sagemakerEndpointConfigurationResourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
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

// TestAccSageMakerEndpoint_changeEndpointConfig_namePrefix_createBeforeDestroy tests for an Endpoint Configuration being
// replaced due to a change and creating the replacement remote resource before destroying the original.
// This is the correct configuration.
func TestAccSageMakerEndpoint_changeEndpointConfig_namePrefix_createBeforeDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sagemaker_endpoint.test"
	sagemakerEndpointConfigurationResourceName := "aws_sagemaker_endpoint_configuration.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_changeEndpointConfig_namePrefix_createBeforeDestroy(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName, names.AttrName),
					acctest.CheckResourceAttrHasPrefix(sagemakerEndpointConfigurationResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.variant_name", rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.initial_instance_count", "2"),
				),
			},
			{
				Config: testAccEndpointConfig_changeEndpointConfig_namePrefix_createBeforeDestroy(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName, names.AttrName),
					acctest.CheckResourceAttrHasPrefix(sagemakerEndpointConfigurationResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.variant_name", rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.initial_instance_count", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(sagemakerEndpointConfigurationResourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
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

// TestAccSageMakerEndpoint_changeEndpointConfig_namePrefix_noCreateBeforeDestroy tests for an Endpoint Configuration being
// replaced due to a change, but not creating the replacement remote resource before destroying the original.
// NOTE: This results in an error
func TestAccSageMakerEndpoint_changeEndpointConfig_namePrefix_noCreateBeforeDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sagemaker_endpoint.test"
	sagemakerEndpointConfigurationResourceName := "aws_sagemaker_endpoint_configuration.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_changeEndpointConfig_namePrefix_noCreateBeforeDestroy(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName, names.AttrName),
					acctest.CheckResourceAttrHasPrefix(sagemakerEndpointConfigurationResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.variant_name", rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.initial_instance_count", "2"),
				),
			},
			{
				Config:      testAccEndpointConfig_changeEndpointConfig_namePrefix_noCreateBeforeDestroy(rName, 1),
				ExpectError: regexache.MustCompile(`Could not find endpoint configuration`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(sagemakerEndpointConfigurationResourceName, plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerEndpoint_AppAutoScaling_replaceEndpointConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_sagemaker_endpoint.test"
	sagemakerEndpointConfigurationResourceName := "aws_sagemaker_endpoint_configuration.test"

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	variantName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	variantName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointConfig_AppAutoScaling_basic(rName, variantName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName, names.AttrName),
					acctest.CheckResourceAttrHasPrefix(sagemakerEndpointConfigurationResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.variant_name", variantName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEndpointConfig_AppAutoScaling_basic(rName, variantName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_config_name", sagemakerEndpointConfigurationResourceName, names.AttrName),
					acctest.CheckResourceAttrHasPrefix(sagemakerEndpointConfigurationResourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(sagemakerEndpointConfigurationResourceName, "production_variants.0.variant_name", variantName2),
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

func testAccCheckEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_endpoint" {
				continue
			}

			_, err := tfsagemaker.FindEndpointByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker AI Endpoint (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckEndpointExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		_, err := tfsagemaker.FindEndpointByNameExcludeDeleting(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccEndpointConfig_model_base(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"

    actions = [
      "cloudwatch:PutMetricData",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "logs:CreateLogGroup",
      "logs:DescribeLogStreams",
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "s3:GetObject",
    ]

    resources = ["*"]
  }
}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.access.json
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "model.tar.gz"
  source = "test-fixtures/sagemaker-tensorflow-serving-test-model.tar.gz"
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-tensorflow-serving"
  image_tag       = "1.12-cpu"
}

resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_url = "https://${aws_s3_bucket.test.bucket_regional_domain_name}/${aws_s3_object.test.key}"
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName)
}

func testAccEndpointConfig_endpointConfiguration_base(rName string) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_model_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    initial_instance_count = 2
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = "variant-1"
  }
}
`, rName))
}

func testAccEndpointConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q
}
`, rName))
}

func testAccEndpointConfig_endpointConfigNameUpdate(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test2" {
  name = "%[1]s-updated"

  production_variants {
    initial_instance_count = 1
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = "variant-1"
  }
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test2.name
  name                 = %[1]q
}
`, rName))
}

func testAccEndpointConfig_retainAllVariantProperties(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_endpoint_configuration" "test2" {
  name = "%[1]s-updated"

  production_variants {
    initial_instance_count = 1
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = "variant-1"
  }
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test2.name
  name                 = %[1]q

  retain_all_variant_properties = true

  exclude_retained_variant_properties {
    variant_property_type = "DesiredInstanceCount"
  }
}
`, rName))
}

func testAccEndpointConfig_tags(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  tags = {
    foo = "bar"
  }
}
`, rName))
}

func testAccEndpointConfig_tagsUpdate(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  tags = {
    bar = "baz"
  }
}
`, rName))
}

func testAccEndpointConfig_deploymentBasic(rName, tType string, wait int) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  deployment_config {
    blue_green_update_policy {
      traffic_routing_configuration {
        type                     = %[2]q
        wait_interval_in_seconds = %[3]d
      }
    }
  }
}
`, rName, tType, wait))
}

func testAccEndpointConfig_deploymentBlueGreen(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  deployment_config {
    blue_green_update_policy {
      traffic_routing_configuration {
        type                     = "LINEAR"
        wait_interval_in_seconds = "60"

        linear_step_size {
          type  = "INSTANCE_COUNT"
          value = 1
        }
      }
    }

    auto_rollback_configuration {
      alarms {
        alarm_name = aws_cloudwatch_metric_alarm.test.alarm_name
      }
    }
  }
}
`, rName))
}

func testAccEndpointConfig_deploymentRolling(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_endpointConfiguration_base(rName), fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  deployment_config {
    auto_rollback_configuration {
      alarms {
        alarm_name = aws_cloudwatch_metric_alarm.test.alarm_name
      }
    }

    rolling_update_policy {
      wait_interval_in_seconds = 60

      maximum_batch_size {
        type  = "CAPACITY_PERCENT"
        value = 5
      }
    }
  }
}
`, rName))
}

func testAccEndpointConfig_changeEndpointConfig_variantName(rName, variantName string, instanceCount int) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_model_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name = %[1]q

  production_variants {
    initial_instance_count = %[3]d
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = %[2]q
  }
}
`, rName, variantName, instanceCount))
}

func testAccEndpointConfig_changeEndpointConfig_namePrefix_createBeforeDestroy(rName string, instanceCount int) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_model_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name_prefix = "%[1]s-"

  production_variants {
    initial_instance_count = %[2]d
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, instanceCount))
}

func testAccEndpointConfig_changeEndpointConfig_namePrefix_noCreateBeforeDestroy(rName string, instanceCount int) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_model_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name_prefix = "%[1]s-"

  production_variants {
    initial_instance_count = %[2]d
    initial_variant_weight = 1
    instance_type          = "ml.t2.medium"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = %[1]q
  }
}
`, rName, instanceCount))
}

func testAccEndpointConfig_AppAutoScaling_basic(rName, variantName string) string {
	return acctest.ConfigCompose(
		testAccEndpointConfig_model_base(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name_prefix = "%[1]s-"

  production_variants {
    initial_instance_count = 2
    initial_variant_weight = 1
    instance_type          = "ml.m5.large"
    model_name             = aws_sagemaker_model.test.name
    variant_name           = %[2]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_appautoscaling_target" "test" {
  service_namespace  = "sagemaker"
  resource_id        = "endpoint/${aws_sagemaker_endpoint.test.name}/variant/${aws_sagemaker_endpoint_configuration.test.production_variants[0].variant_name}"
  scalable_dimension = "sagemaker:variant:DesiredInstanceCount"
  min_capacity       = 1
  max_capacity       = 3
}

resource "aws_appautoscaling_policy" "test" {
  name = %[1]q

  service_namespace  = aws_appautoscaling_target.test.service_namespace
  resource_id        = aws_appautoscaling_target.test.resource_id
  scalable_dimension = aws_appautoscaling_target.test.scalable_dimension

  policy_type = "TargetTrackingScaling"

  target_tracking_scaling_policy_configuration {
    target_value = 70.0

    predefined_metric_specification {
      predefined_metric_type = "SageMakerVariantInvocationsPerInstance"
    }
  }
}
`, rName, variantName))
}
