// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderImagePipeline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageRecipeResourceName := "aws_imagebuilder_image_recipe.test"
	infrastructureConfigurationResourceName := "aws_imagebuilder_infrastructure_configuration.test"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "imagebuilder", fmt.Sprintf("image-pipeline/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_last_run", ""),
					resource.TestCheckResourceAttr(resourceName, "date_next_run", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "distribution_configuration_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "720"),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, imagebuilder.PipelineStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccImageBuilderImagePipeline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfimagebuilder.ResourceImagePipeline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_distributionARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	distributionConfigurationResourceName := "aws_imagebuilder_distribution_configuration.test"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_distributionConfigurationARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "distribution_configuration_arn", distributionConfigurationResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_distributionConfigurationARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "distribution_configuration_arn", distributionConfigurationResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_enhancedImageMetadataEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_enhancedMetadataEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_enhancedMetadataEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_imageRecipeARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageRecipeResourceName := "aws_imagebuilder_image_recipe.test"
	imageRecipeResourceName2 := "aws_imagebuilder_image_recipe.test2"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_recipeARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_containerRecipeARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerRecipeResourceName := "aws_imagebuilder_container_recipe.test"
	containerRecipeResourceName2 := "aws_imagebuilder_container_recipe.test2"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_containerRecipeARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "container_recipe_arn", containerRecipeResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_containerRecipeARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "container_recipe_arn", containerRecipeResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_ImageScanning_imageScanningEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_testsConfigurationScanningEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.image_scanning_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_testsConfigurationScanningEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.image_scanning_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_ImageScanning_imageScanningEnabledAdvanced(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_testsConfigurationScanningEnabledAdvanced(rName, []string{"a", "b"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.image_scanning_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.ecr_configuration.0.repository_name", rName),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_scanning_configuration.0.ecr_configuration.0.container_tags.*", "b"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_scanning_configuration.0.ecr_configuration.0.container_tags.*", "a"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_testsConfigurationScanningEnabledAdvanced(rName, []string{"a", "c"}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.image_scanning_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.ecr_configuration.0.repository_name", rName),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_scanning_configuration.0.ecr_configuration.0.container_tags.*", "c"),
					resource.TestCheckTypeSetElemAttr(resourceName, "image_scanning_configuration.0.ecr_configuration.0.container_tags.*", "a"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_ImageTests_imageTestsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_testsConfigurationTestsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_testsConfigurationTestsEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_ImageTests_timeoutMinutes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_testsConfigurationTimeoutMinutes(rName, 721),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "721"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_testsConfigurationTimeoutMinutes(rName, 722),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "722"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_infrastructureARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	infrastructureConfigurationResourceName := "aws_imagebuilder_infrastructure_configuration.test"
	infrastructureConfigurationResourceName2 := "aws_imagebuilder_infrastructure_configuration.test2"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_infrastructureConfigurationARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName2, names.AttrARN),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_Schedule_pipelineExecutionStartCondition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_scheduleExecutionStartCondition(rName, imagebuilder.PipelineExecutionStartConditionExpressionMatchAndDependencyUpdatesAvailable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.pipeline_execution_start_condition", imagebuilder.PipelineExecutionStartConditionExpressionMatchAndDependencyUpdatesAvailable),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"date_next_run"},
			},
			{
				Config: testAccImagePipelineConfig_scheduleExecutionStartCondition(rName, imagebuilder.PipelineExecutionStartConditionExpressionMatchOnly),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.pipeline_execution_start_condition", imagebuilder.PipelineExecutionStartConditionExpressionMatchOnly),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_Schedule_scheduleExpression(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_scheduleExpression(rName, "cron(1 0 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(1 0 * * ? *)"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"date_next_run"},
			},
			{
				Config: testAccImagePipelineConfig_scheduleExpression(rName, "cron(2 0 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(2 0 * * ? *)"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_Schedule_timezone(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_scheduleTimezone(rName, "cron(1 0 * * ? *)", "Etc/UTC"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(1 0 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.timezone", "Etc/UTC"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"date_next_run"},
			},
			{
				Config: testAccImagePipelineConfig_scheduleTimezone(rName, "cron(1 0 * * ? *)", "America/Los_Angeles"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(1 0 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.timezone", "America/Los_Angeles"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_status(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_status(rName, imagebuilder.PipelineStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, imagebuilder.PipelineStatusDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_status(rName, imagebuilder.PipelineStatusEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, imagebuilder.PipelineStatusEnabled),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccImagePipelineConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_workflow(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_workflow(rName, imagebuilder.OnWorkflowFailureAbort, "test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.on_failure", imagebuilder.OnWorkflowFailureAbort),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.parallel_group", "test1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_workflow(rName, imagebuilder.OnWorkflowFailureContinue, "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.on_failure", imagebuilder.OnWorkflowFailureContinue),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.parallel_group", "test2"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_workflowParameter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImagePipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_workflowParameter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.parameter.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_workflowParameter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.parameter.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckImagePipelineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_image_pipeline" {
				continue
			}

			input := &imagebuilder.GetImagePipelineInput{
				ImagePipelineArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetImagePipelineWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Image Builder Image Pipeline (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Image Builder Image Pipeline (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckImagePipelineExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		input := &imagebuilder.GetImagePipelineInput{
			ImagePipelineArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetImagePipelineWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image Pipeline (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccImagePipelineConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.role.name
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}

resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}

resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_ecr_repository" "test" {
  name                 = %[1]q
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = false
  }
}

resource "aws_imagebuilder_container_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  dockerfile_template_data = <<EOF
  FROM {{{ imagebuilder:parentImage }}}
  {{{ imagebuilder:environments }}}
  {{{ imagebuilder:components }}}
  EOF

  container_type    = "DOCKER"
  name              = %[1]q
  parent_image      = "amazonlinux:latest"
  working_directory = "/tmp"
  version           = "1.0.0"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName)
}

func testAccImagePipelineConfig_description(rName string, description string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  description                      = %[2]q
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName, description))
}

func testAccImagePipelineConfig_distributionConfigurationARN1(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_distribution_configuration" "test" {
  name = "%[1]s-1"

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_imagebuilder_image_pipeline" "test" {
  distribution_configuration_arn   = aws_imagebuilder_distribution_configuration.test.arn
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_distributionConfigurationARN2(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_distribution_configuration" "test" {
  name = "%[1]s-2"

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_imagebuilder_image_pipeline" "test" {
  distribution_configuration_arn   = aws_imagebuilder_distribution_configuration.test.arn
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_enhancedMetadataEnabled(rName string, enhancedImageMetadataEnabled bool) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  enhanced_image_metadata_enabled  = %[2]t
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName, enhancedImageMetadataEnabled))
}

func testAccImagePipelineConfig_recipeARN2(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test2" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = "%[1]s-2"
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test2.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_containerRecipeARN1(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_containerRecipeARN2(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test2" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  name           = "%[1]s-2"
  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-latest/x.x.x"
  version        = "1.0.0"

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}

resource "aws_imagebuilder_image_pipeline" "test" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test2.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_testsConfigurationScanningEnabled(rName string, imageScanningEnabled bool) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  image_scanning_configuration {
    image_scanning_enabled = %[2]t
  }
}
`, rName, imageScanningEnabled))
}

func testAccImagePipelineConfig_testsConfigurationScanningEnabledAdvanced(rName string, imageTags []string) string {
	commaSepImageTags := ""
	if len(imageTags) > 0 {
		commaSepImageTags = "\"" + strings.Join(imageTags, "\", \"") + "\""
	}

	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  image_scanning_configuration {
    image_scanning_enabled = true

    ecr_configuration {
      container_tags  = [%[2]s]
      repository_name = aws_ecr_repository.test.name
    }
  }
}
`, rName, commaSepImageTags))
}

func testAccImagePipelineConfig_testsConfigurationTestsEnabled(rName string, imageTestsEnabled bool) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  image_tests_configuration {
    image_tests_enabled = %[2]t
  }
}
`, rName, imageTestsEnabled))
}

func testAccImagePipelineConfig_testsConfigurationTimeoutMinutes(rName string, timeoutMinutes int) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  image_tests_configuration {
    timeout_minutes = %[2]d
  }
}
`, rName, timeoutMinutes))
}

func testAccImagePipelineConfig_infrastructureConfigurationARN2(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_infrastructure_configuration" "test2" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = "%[1]s-2"
}

resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test2.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_scheduleExecutionStartCondition(rName string, pipelineExecutionStartCondition string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  schedule {
    pipeline_execution_start_condition = %[2]q
    schedule_expression                = "cron(0 0 * * ? *)"
  }
}
`, rName, pipelineExecutionStartCondition))
}

func testAccImagePipelineConfig_scheduleExpression(rName string, scheduleExpression string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  schedule {
    schedule_expression = %[2]q
  }
}
`, rName, scheduleExpression))
}

func testAccImagePipelineConfig_scheduleTimezone(rName string, scheduleExpression string, timezone string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  schedule {
    schedule_expression = %[2]q
    timezone            = %[3]q
  }
}
`, rName, scheduleExpression, timezone))
}
func testAccImagePipelineConfig_status(rName string, status string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
  status                           = %[2]q
}
`, rName, status))
}

func testAccImagePipelineConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccImagePipelineConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccImagePipelineConfig_workflow(rName, onFailure, parallelGroup string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"
  EOT
}

resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  execution_role = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/imagebuilder.amazonaws.com/AWSServiceRoleForImageBuilder"

  workflow {
    on_failure     = %[2]q
    parallel_group = %[3]q
    workflow_arn   = aws_imagebuilder_workflow.test.arn
  }
}
`, rName, onFailure, parallelGroup))
}

func testAccImagePipelineConfig_workflowParameter(rName string) string {
	return acctest.ConfigCompose(testAccImagePipelineConfig_base(rName), fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}

resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q

  execution_role = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/imagebuilder.amazonaws.com/AWSServiceRoleForImageBuilder"

  workflow {
    on_failure     = "ABORT"
    parallel_group = "test"
    workflow_arn   = aws_imagebuilder_workflow.test.arn

    parameter {
      name  = "waitForActionAtEnd"
      value = "true"
    }
  }
}
`, rName))
}
