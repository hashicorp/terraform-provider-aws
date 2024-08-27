// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccImageBuilderImage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageRecipeResourceName := "aws_imagebuilder_image_recipe.test"
	infrastructureConfigurationResourceName := "aws_imagebuilder_infrastructure_configuration.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "imagebuilder", regexache.MustCompile(fmt.Sprintf("image/%s/1.0.0/[1-9][0-9]*", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "container_recipe_arn"),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckNoResourceAttr(resourceName, "distribution_configuration_arn"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "720"),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "os_version", "Amazon Linux 2"),
					resource.TestCheckResourceAttr(resourceName, "output_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, names.AttrVersion, regexache.MustCompile(`1.0.0/[1-9][0-9]*`)),
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

func TestAccImageBuilderImage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfimagebuilder.ResourceImage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderImage_distributionARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	distributionConfigurationResourceName := "aws_imagebuilder_distribution_configuration.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_distributionConfigurationARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "distribution_configuration_arn", distributionConfigurationResourceName, names.AttrARN),
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

func TestAccImageBuilderImage_enhancedImageMetadataEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_enhancedMetadataEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", acctest.CtFalse),
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

func TestAccImageBuilderImage_ImageTests_imageTestsEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_testsConfigurationTestsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", acctest.CtFalse),
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

func TestAccImageBuilderImage_ImageTests_timeoutMinutes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_testsConfigurationTimeoutMinutes(rName, 721),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "721"),
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

func TestAccImageBuilderImage_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
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
				Config: testAccImageConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccImageConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderImage_containerRecipeARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"
	containerRecipeResourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_containerRecipe(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "container_recipe_arn", containerRecipeResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccImageBuilderImage_imageScanningConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Inspector2EndpointID)
			acctest.PreCheckInspector2(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_imageScanningConfigurationEnabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccImageBuilderImage_outputResources_containers(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"
	regionDataSourceName := "data.aws_region.current"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_containerRecipe(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_resources.0.containers.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "output_resources.0.containers.0.image_uris.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "output_resources.0.containers.0.region", regionDataSourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccImageBuilderImage_workflows(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_workflows(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "workflow.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "workflow.0.workflow_arn", "aws_imagebuilder_workflow.test_build", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.parameter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.parameter.0.name", "foo"),
					resource.TestCheckResourceAttr(resourceName, "workflow.0.parameter.0.value", "bar"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.on_failure", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "workflow.1.parallel_group", "baz"),
				),
			},
		},
	})
}

func testAccCheckImageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_image_pipeline" {
				continue
			}

			input := &imagebuilder.GetImageInput{
				ImageBuildVersionArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetImageWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Image Builder Image (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Image Builder Image (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckImageExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		input := &imagebuilder.GetImageInput{
			ImageBuildVersionArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetImageWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccImageBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_imagebuilder_component" "update-linux" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/1.0.0"
}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.AmazonSSMManagedInstanceCore,
    aws_iam_role_policy_attachment.EC2InstanceProfileForImageBuilder,
  ]
}

resource "aws_iam_role" "test" {
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

resource "aws_iam_role_policy_attachment" "AmazonSSMManagedInstanceCore" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilder" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilder"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  ingress {
    from_port = 0
    protocol  = -1
    self      = true
    to_port   = 0
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id
}

resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = data.aws_imagebuilder_component.update-linux.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_default_security_group.test.id]
  subnet_id             = aws_subnet.test.id

  depends_on = [aws_default_route_table.test]
}
`, rName)
}

func testAccImageConfig_distributionConfigurationARN(rName string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}

resource "aws_imagebuilder_image" "test" {
  distribution_configuration_arn   = aws_imagebuilder_distribution_configuration.test.arn
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName))
}

func testAccImageConfig_enhancedMetadataEnabled(rName string, enhancedImageMetadataEnabled bool) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  enhanced_image_metadata_enabled  = %[2]t
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName, enhancedImageMetadataEnabled))
}

func testAccImageConfig_testsConfigurationTestsEnabled(rName string, imageTestsEnabled bool) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  image_tests_configuration {
    image_tests_enabled = %[2]t
  }
}
`, rName, imageTestsEnabled))
}

func testAccImageConfig_testsConfigurationTimeoutMinutes(rName string, timeoutMinutes int) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  image_tests_configuration {
    timeout_minutes = %[2]d
  }
}
`, rName, timeoutMinutes))
}

func testAccImageConfig_required(rName string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`)
}

func testAccImageConfig_workflows(rName string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test_build" {
  name    = join("-", [%[1]q, "build"])
  version = "1.0.0"
  type    = "BUILD"

  data = <<-EOT
name: ${join("-", [%[1]q, "build"])}
description: Workflow to build an AMI
schemaVersion: 1.0

parameters:
  - name: foo
    type: string

steps:
  - name: LaunchBuildInstance
    action: LaunchInstance
    onFailure: Abort
    inputs:
      waitFor: "ssmAgent"

  - name: UpdateSSMAgent
    action: RunCommand
    onFailure: Abort
    inputs:
      documentName: "AWS-UpdateSSMAgent"
      instanceId.$: "$.stepOutputs.LaunchBuildInstance.instanceId"
      parameters:
        allowDowngrade:
          - "false"

  - name: ApplyBuildComponents
    action: ExecuteComponents
    onFailure: Abort
    inputs:
      instanceId.$: "$.stepOutputs.LaunchBuildInstance.instanceId"

  - name: InventoryCollection
    action: CollectImageMetadata
    onFailure: Abort
    if:
      and:
        - stringEquals: "AMI"
          value: "$.imagebuilder.imageType"
        - booleanEquals: true
          value: "$.imagebuilder.collectImageMetadata"
    inputs:
      instanceId.$: "$.stepOutputs.LaunchBuildInstance.instanceId"

  - name: RunSanitizeScript
    action: SanitizeInstance
    onFailure: Abort
    if:
      and:
        - stringEquals: "AMI"
          value: "$.imagebuilder.imageType"
        - stringEquals: "Linux"
          value: "$.imagebuilder.platform"
    inputs:
      instanceId.$: "$.stepOutputs.LaunchBuildInstance.instanceId"

  - name: RunSysPrepScript
    action: RunSysPrep
    onFailure: Abort
    if:
      and:
        - stringEquals: "AMI"
          value: "$.imagebuilder.imageType"
        - stringEquals: "Windows"
          value: "$.imagebuilder.platform"
    inputs:
      instanceId.$: "$.stepOutputs.LaunchBuildInstance.instanceId"

  - name: CreateOutputAMI
    action: CreateImage
    onFailure: Abort
    if:
      stringEquals: "AMI"
      value: "$.imagebuilder.imageType"
    inputs:
      instanceId.$: "$.stepOutputs.LaunchBuildInstance.instanceId"

  - name: TerminateBuildInstance
    action: TerminateInstance
    onFailure: Continue
    inputs:
      instanceId.$: "$.stepOutputs.LaunchBuildInstance.instanceId"

outputs:
  - name: "ImageId"
    value: "$.stepOutputs.CreateOutputAMI.imageId"
  EOT
}

resource "aws_imagebuilder_workflow" "test_test" {
  name    = join("-", [%[1]q, "test"])
  version = "1.0.0"
  type    = "TEST"

  data = <<-EOT
name: ${join("-", [%[1]q, "test"])}
description: Workflow to test an AMI
schemaVersion: 1.0

steps:
  - name: LaunchTestInstance
    action: LaunchInstance
    onFailure: Abort
    inputs:
      waitFor: "ssmAgent"

  - name: CollectImageScanFindings
    action: CollectImageScanFindings
    onFailure: Continue
    if:
      and:
        - booleanEquals: true
          value: "$.imagebuilder.collectImageScanFindings"
        - or:
            - stringEquals: "Linux"
              value: "$.imagebuilder.platform"
            - stringEquals: "Windows"
              value: "$.imagebuilder.platform"
    inputs:
      instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

  - name: ApplyTestComponents
    action: ExecuteComponents
    onFailure: Abort
    inputs:
      instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

  - name: TerminateTestInstance
    action: TerminateInstance
    onFailure: Continue
    inputs:
      instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"
  EOT
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test_execute" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "imagebuilder.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = join("-", [%[1]q, "execute"])
}

data "aws_iam_policy" "AWSServiceRoleForImageBuilder" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/aws-service-role/AWSServiceRoleForImageBuilder"
}

resource "aws_iam_policy" "test_execute_service_policy" {
  name   = join("-", [%[1]q, "execute-service"])
  policy = data.aws_iam_policy.AWSServiceRoleForImageBuilder.policy
}

resource "aws_iam_role_policy_attachment" "test_execute_service" {
  policy_arn = aws_iam_policy.test_execute_service_policy.arn
  role       = aws_iam_role.test_execute.name
}

resource "aws_iam_policy" "test_execute" {
  name = join("-", [%[1]q, "execute"])
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action   = "ssm:SendCommand"
      Effect   = "Allow"
      Resource = "arn:${data.aws_partition.current.partition}:ssm:${data.aws_region.current.id}::document/AWS-UpdateSSMAgent"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test_execute" {
  policy_arn = aws_iam_policy.test_execute.arn
  role       = aws_iam_role.test_execute.name
}

resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  execution_role                   = aws_iam_role.test_execute.arn

  workflow {
    workflow_arn = aws_imagebuilder_workflow.test_build.arn

    parameter {
      name  = "foo"
      value = "bar"
    }
  }

  workflow {
    workflow_arn   = aws_imagebuilder_workflow.test_test.arn
    on_failure     = "CONTINUE"
    parallel_group = "baz"
  }

  depends_on = [
    aws_iam_role_policy_attachment.test_execute,
    aws_iam_role_policy_attachment.test_execute_service
  ]
}
`, rName),
	)
}
func testAccImageConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccImageConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccImageConfig_containerRecipeBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  ingress {
    from_port = 0
    protocol  = -1
    self      = true
    to_port   = 0
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
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

resource "aws_iam_role_policy_attachment" "AmazonSSMManagedInstanceCore" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilderECRContainerBuilds" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"
  role       = aws_iam_role.test.name
}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.AmazonSSMManagedInstanceCore,
    aws_iam_role_policy_attachment.EC2InstanceProfileForImageBuilderECRContainerBuilds
  ]
}

resource "aws_ecr_repository" "test" {
  name         = %[1]q
  force_delete = true
}

data "aws_imagebuilder_component" "update-linux" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/1.0.0"
}

resource "aws_imagebuilder_container_recipe" "test" {
  component {
    component_arn = data.aws_imagebuilder_component.update-linux.arn
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

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_default_security_group.test.id]
  subnet_id             = aws_subnet.test.id

  depends_on = [aws_default_route_table.test]
}
	`, rName)
}

func testAccImageConfig_containerRecipe(rName string) string {
	return acctest.ConfigCompose(
		testAccImageConfig_containerRecipeBase(rName),
		`
resource "aws_imagebuilder_image" "test" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`)
}

func testAccImageConfig_imageScanningConfigurationEnabled(rName string) string {
	return acctest.ConfigCompose(
		testAccImageConfig_containerRecipeBase(rName),
		`
data "aws_caller_identity" "current" {}

resource "aws_inspector2_enabler" "test" {
  account_ids    = [data.aws_caller_identity.current.account_id]
  resource_types = ["ECR"]
}

resource "aws_imagebuilder_image" "test" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  image_scanning_configuration {
    image_scanning_enabled = true

    ecr_configuration {
      repository_name = aws_ecr_repository.test.name
      container_tags  = ["foo", "bar"]
    }
  }

  depends_on = [aws_inspector2_enabler.test]
}
`)
}
