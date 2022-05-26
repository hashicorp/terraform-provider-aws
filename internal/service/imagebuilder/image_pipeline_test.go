package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
)

func TestAccImageBuilderImagePipeline_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageRecipeResourceName := "aws_imagebuilder_image_recipe.test"
	infrastructureConfigurationResourceName := "aws_imagebuilder_infrastructure_configuration.test"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", fmt.Sprintf("image-pipeline/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_last_run", ""),
					resource.TestCheckResourceAttr(resourceName, "date_next_run", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "distribution_configuration_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "720"),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", imagebuilder.PipelineStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfimagebuilder.ResourceImagePipeline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_distributionARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	distributionConfigurationResourceName := "aws_imagebuilder_distribution_configuration.test"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_distributionConfigurationARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "distribution_configuration_arn", distributionConfigurationResourceName, "arn"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "distribution_configuration_arn", distributionConfigurationResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_enhancedImageMetadataEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_enhancedMetadataEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", "false"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", "true"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_imageRecipeARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageRecipeResourceName := "aws_imagebuilder_image_recipe.test"
	imageRecipeResourceName2 := "aws_imagebuilder_image_recipe.test2"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName, "arn"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_containerRecipeARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	containerRecipeResourceName := "aws_imagebuilder_container_recipe.test"
	containerRecipeResourceName2 := "aws_imagebuilder_container_recipe.test2"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_containerRecipeARN1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "container_recipe_arn", containerRecipeResourceName, "arn"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "container_recipe_arn", containerRecipeResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_ImageTests_imageTestsEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_testsConfigurationTestsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", "false"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", "true"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_ImageTests_timeoutMinutes(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_testsConfigurationTimeoutMinutes(rName, 721),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "722"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_infrastructureARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	infrastructureConfigurationResourceName := "aws_imagebuilder_infrastructure_configuration.test"
	infrastructureConfigurationResourceName2 := "aws_imagebuilder_infrastructure_configuration.test2"
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName, "arn"),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_Schedule_pipelineExecutionStartCondition(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_scheduleExecutionStartCondition(rName, imagebuilder.PipelineExecutionStartConditionExpressionMatchAndDependencyUpdatesAvailable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.pipeline_execution_start_condition", imagebuilder.PipelineExecutionStartConditionExpressionMatchAndDependencyUpdatesAvailable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_scheduleExecutionStartCondition(rName, imagebuilder.PipelineExecutionStartConditionExpressionMatchOnly),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.pipeline_execution_start_condition", imagebuilder.PipelineExecutionStartConditionExpressionMatchOnly),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_Schedule_scheduleExpression(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_scheduleExpression(rName, "cron(1 0 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(1 0 * * ? *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_scheduleExpression(rName, "cron(2 0 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(2 0 * * ? *)"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_Schedule_timezone(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_scheduleTimezone(rName, "cron(1 0 * * ? *)", "Etc/UTC"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(1 0 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.timezone", "Etc/UTC"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_scheduleTimezone(rName, "cron(1 0 * * ? *)", "America/Los_Angeles"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(1 0 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.timezone", "America/Los_Angeles"),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_status(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_status(rName, imagebuilder.PipelineStatusDisabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", imagebuilder.PipelineStatusDisabled),
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
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", imagebuilder.PipelineStatusEnabled),
				),
			},
		},
	})
}

func TestAccImageBuilderImagePipeline_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_pipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImagePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImagePipelineConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImagePipelineConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccImagePipelineConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImagePipelineExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckImagePipelineDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_image_pipeline" {
			continue
		}

		input := &imagebuilder.GetImagePipelineInput{
			ImagePipelineArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetImagePipeline(input)

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

func testAccCheckImagePipelineExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

		input := &imagebuilder.GetImagePipelineInput{
			ImagePipelineArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetImagePipeline(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image Pipeline (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccImagePipelineBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

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

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}
`, rName)
}

func testAccImagePipelineConfig_description(rName string, description string) string {
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  description                      = %[2]q
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName, description))
}

func testAccImagePipelineConfig_distributionConfigurationARN1(rName string) string {
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  enhanced_image_metadata_enabled  = %[2]t
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName, enhancedImageMetadataEnabled))
}

func testAccImagePipelineConfig_recipeARN2(rName string) string {
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_imagebuilder_container_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  name           = %[1]q
  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-latest/x.x.x"
  version        = "1.0.0"

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}

resource "aws_imagebuilder_image_pipeline" "test" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_containerRecipeARN2(rName string) string {
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

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

func testAccImagePipelineConfig_testsConfigurationTestsEnabled(rName string, imageTestsEnabled bool) string {
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName))
}

func testAccImagePipelineConfig_scheduleExecutionStartCondition(rName string, pipelineExecutionStartCondition string) string {
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_pipeline" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
  status                           = %[2]q
}
`, rName, status))
}

func testAccImagePipelineConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccImagePipelineBaseConfig(rName),
		fmt.Sprintf(`
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
