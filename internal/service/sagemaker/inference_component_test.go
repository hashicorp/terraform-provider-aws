// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerInferenceComponent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "endpoint_arn", "sagemaker", fmt.Sprintf("endpoint/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
					resource.TestCheckResourceAttr(resourceName, "variant_name", "variant-1"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, "runtime_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_config.0.copy_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_config.0.current_copy_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_config.0.desired_copy_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.min_memory_required_in_mb", "1024"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.number_of_cpu_cores_required", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "specification.0.container.0.resolved_image"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "InService"),
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

func TestAccSageMakerInferenceComponent_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsagemaker.ResourceInferenceComponent(), resourceName),
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

func TestAccSageMakerInferenceComponent_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccInferenceComponentConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccInferenceComponentConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerInferenceComponent_runtimeConfigUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_runtimeConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "runtime_config.0.copy_count", "1"),
				),
			},
			{
				Config: testAccInferenceComponentConfig_runtimeConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "runtime_config.0.copy_count", "2"),
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

func TestAccSageMakerInferenceComponent_container(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_container(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.container.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "specification.0.container.0.image"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.container.0.environment.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.container.0.environment.MY_ENV", "my_value"),
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

func TestAccSageMakerInferenceComponent_startupParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_startupParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.startup_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.startup_parameters.0.container_startup_health_check_timeout_in_seconds", "120"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.startup_parameters.0.model_data_download_timeout_in_seconds", "120"),
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

func TestAccSageMakerInferenceComponent_dataCacheConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_dataCacheConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.data_cache_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.data_cache_config.0.enable_caching", "true"),
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

func TestAccSageMakerInferenceComponent_schedulingConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_schedulingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.scheduling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.scheduling_config.0.placement_strategy", "SPREAD"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.scheduling_config.0.availability_zone_balance.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.scheduling_config.0.availability_zone_balance.0.enforcement_mode", "PERMISSIVE"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.scheduling_config.0.availability_zone_balance.0.max_imbalance", "1"),
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

func TestAccSageMakerInferenceComponent_containerImage(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_containerImage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "specification.0.container.0.image"),
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

func TestAccSageMakerInferenceComponent_deploymentConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// deployment_config is an update-only rollout directive that the API rejects
				// unless it accompanies a spec or runtime change. Create with copy_count 2
				// and no deployment_config, then in the next step bump copy_count to 3 and
				// attach deployment_config to govern that rollout. maximum_batch_size uses
				// CAPACITY_PERCENT so it is independent of the copy count (the API validates
				// a COPY_COUNT batch against the current count, which is fragile in a test).
				Config: testAccInferenceComponentConfig_runtimeConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "0"),
				),
			},
			{
				Config: testAccInferenceComponentConfig_deploymentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.wait_interval_in_seconds", "60"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.maximum_execution_timeout_in_seconds", "1800"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.maximum_batch_size.0.type", "CAPACITY_PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.maximum_batch_size.0.value", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.rollback_maximum_batch_size.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.rollback_maximum_batch_size.0.type", "COPY_COUNT"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.rolling_update_policy.0.rollback_maximum_batch_size.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config.0.auto_rollback_configuration.0.alarms.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"deployment_config"},
			},
		},
	})
}

func TestAccSageMakerInferenceComponent_computeResourceRequirements(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_computeResourceFull(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.min_memory_required_in_mb", "1024"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.max_memory_required_in_mb", "2048"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.number_of_cpu_cores_required", "1"),
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

func TestAccSageMakerInferenceComponent_acceleratorDevices(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_acceleratorDevices(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.number_of_accelerator_devices_required", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.scheduling_config.0.placement_strategy", "BINPACK"),
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

func TestAccSageMakerInferenceComponent_modelName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_modelName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.model_name", rName),
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

func TestAccSageMakerInferenceComponent_specifications(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_specifications(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "specifications.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "specifications.0.instance_type", "ml.c5.xlarge"),
					resource.TestCheckResourceAttr(resourceName, "specifications.1.instance_type", "ml.c5.2xlarge"),
					resource.TestCheckResourceAttr(resourceName, "specifications.0.model_name", rName),
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

func TestAccSageMakerInferenceComponent_adapter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.adapter"
	baseResourceName := "aws_sagemaker_inference_component.base"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_adapter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, baseResourceName),
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					// An adapter inference component references a base component and inherits
					// its compute resources; the API echoes those inherited resources back on
					// read, so the block is populated even though the config omits it.
					resource.TestCheckResourceAttrPair(resourceName, "specification.0.base_inference_component_name", baseResourceName, names.AttrName),
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

func TestAccSageMakerInferenceComponent_containerMetricsConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_containerMetricsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.container.0.container_metrics_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.container.0.container_metrics_config.0.metrics_endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.container.0.container_metrics_config.0.metrics_endpoint.0.metrics_endpoint_path", "/metrics"),
					resource.TestCheckResourceAttr(resourceName, "specification.0.container.0.container_metrics_config.0.metrics_endpoint.0.metric_publish_frequency_in_seconds", "30"),
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

func TestAccSageMakerInferenceComponent_specificationUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_inference_component.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInferenceComponentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccInferenceComponentConfig_specificationMemory(rName, 1024),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.min_memory_required_in_mb", "1024"),
				),
			},
			{
				Config: testAccInferenceComponentConfig_specificationMemory(rName, 2048),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInferenceComponentExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "specification.0.compute_resource_requirements.0.min_memory_required_in_mb", "2048"),
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

func testAccCheckInferenceComponentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_inference_component" {
				continue
			}

			_, err := tfsagemaker.FindInferenceComponentByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker Inference Component (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckInferenceComponentExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		_, err := tfsagemaker.FindInferenceComponentByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccInferenceComponentConfig_base(rName string) string {
	return testAccInferenceComponentConfig_baseInstanceType(rName, "ml.c5.xlarge")
}

func testAccInferenceComponentConfig_baseInstanceType(rName, instanceType string) string {
	return fmt.Sprintf(`
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

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_iam_role_policy" "s3_access" {
  role = aws_iam_role.test.name
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:s3:::%[1]s",
          "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*",
        ]
      }
    ]
  })
}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-tensorflow-serving"
  image_tag       = "1.12-cpu"
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket       = aws_s3_bucket.test.bucket
  key          = "model/model.tar.gz"
  content_type = "application/gzip"
  source       = "test-fixtures/sagemaker-tensorflow-serving-test-model.tar.gz"
}

resource "aws_sagemaker_endpoint_configuration" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  production_variants {
    variant_name           = "variant-1"
    instance_type          = %[2]q
    initial_instance_count = 1

    managed_instance_scaling {
      status             = "ENABLED"
      min_instance_count = 1
      max_instance_count = 1
    }

    routing_config {
      routing_strategy = "LEAST_OUTSTANDING_REQUESTS"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3_access]
}

resource "aws_sagemaker_endpoint" "test" {
  endpoint_config_name = aws_sagemaker_endpoint_configuration.test.name
  name                 = %[1]q

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3_access]
}
`, rName, instanceType)
}

func testAccInferenceComponentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }

    startup_parameters {
      container_startup_health_check_timeout_in_seconds = 300
      model_data_download_timeout_in_seconds            = 300
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccInferenceComponentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccInferenceComponentConfig_runtimeConfig(rName string, copyCount int) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = %[2]d
  }
}
`, rName, copyCount))
}

func testAccInferenceComponentConfig_container(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"

      environment = {
        MY_ENV = "my_value"
      }
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_startupParameters(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }

    startup_parameters {
      container_startup_health_check_timeout_in_seconds = 120
      model_data_download_timeout_in_seconds            = 120
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_dataCacheConfig(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }

    data_cache_config {
      enable_caching = true
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_schedulingConfig(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }

    scheduling_config {
      placement_strategy = "SPREAD"

      availability_zone_balance {
        enforcement_mode = "PERMISSIVE"
        max_imbalance    = 1
      }
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_containerImage(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_deploymentConfig(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
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

resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 3
  }

  deployment_config {
    rolling_update_policy {
      maximum_batch_size {
        type  = "CAPACITY_PERCENT"
        value = 50
      }

      wait_interval_in_seconds             = 60
      maximum_execution_timeout_in_seconds = 1800

      rollback_maximum_batch_size {
        type  = "COPY_COUNT"
        value = 1
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

func testAccInferenceComponentConfig_computeResourceFull(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      max_memory_required_in_mb    = 2048
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_modelName(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3_access]
}

resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    model_name = aws_sagemaker_model.test.name

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_specifications(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_model" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  primary_container {
    image          = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
    model_data_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy.s3_access]
}

resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specifications {
    instance_type = "ml.c5.xlarge"
    model_name    = aws_sagemaker_model.test.name

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  specifications {
    instance_type = "ml.c5.2xlarge"
    model_name    = aws_sagemaker_model.test.name

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_containerMetricsConfig(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"

      container_metrics_config {
        metrics_endpoint {
          metrics_endpoint_path               = "/metrics"
          metric_publish_frequency_in_seconds = 30
        }
      }
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_specificationMemory(rName string, minMemory int) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = %[2]d
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName, minMemory))
}

func testAccInferenceComponentConfig_acceleratorDevices(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_baseInstanceType(rName, "ml.g5.xlarge"), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "test" {
  name          = %[1]q
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb              = 1024
      number_of_accelerator_devices_required = 1
    }

    scheduling_config {
      placement_strategy = "BINPACK"

      availability_zone_balance {
        enforcement_mode = "PERMISSIVE"
      }
    }
  }

  runtime_config {
    copy_count = 1
  }
}
`, rName))
}

func testAccInferenceComponentConfig_adapter(rName string) string {
	return acctest.ConfigCompose(testAccInferenceComponentConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_inference_component" "base" {
  name          = "%[1]s-base"
  endpoint_name = aws_sagemaker_endpoint.test.name
  variant_name  = "variant-1"

  specification {
    container {
      image        = data.aws_sagemaker_prebuilt_ecr_image.test.registry_path
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }

    compute_resource_requirements {
      min_memory_required_in_mb    = 1024
      number_of_cpu_cores_required = 1
    }
  }

  runtime_config {
    copy_count = 1
  }
}

resource "aws_sagemaker_inference_component" "adapter" {
  name          = "%[1]s-adapter"
  endpoint_name = aws_sagemaker_endpoint.test.name

  specification {
    base_inference_component_name = aws_sagemaker_inference_component.base.name

    container {
      artifact_url = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
    }
  }
}
`, rName))
}
