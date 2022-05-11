package imagebuilder_test

import (
	"fmt"
	"regexp"
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

func TestAccImageBuilderContainerRecipe_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", regexp.MustCompile(fmt.Sprintf("container-recipe/%s/1.0.0", rName))),
					resource.TestCheckResourceAttr(resourceName, "component.#", "1"),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "component.0.component_arn", "imagebuilder", "aws", "component/update-linux/x.x.x"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "container_type", "DOCKER"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttrSet(resourceName, "dockerfile_template_data"),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner"),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "parent_image", "imagebuilder", "aws", "image/amazon-linux-x86-2/x.x.x"),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "target_repository.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_repository.0.repository_name", "aws_ecr_repository.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_repository.0.service", "ECR"),
					resource.TestCheckResourceAttr(resourceName, "version", "1.0.0"),
					resource.TestCheckResourceAttr(resourceName, "working_directory", ""),
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

func TestAccImageBuilderContainerRecipe_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfimagebuilder.ResourceContainerRecipe(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderContainerRecipe_component(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeComponentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", "2"),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "component.0.component_arn", "imagebuilder", "aws", "component/update-linux/x.x.x"),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "component.1.component_arn", "imagebuilder", "aws", "component/aws-cli-version-2-linux/x.x.x"),
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

func TestAccImageBuilderContainerRecipe_componentParameter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeComponentParameterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.0.name", "Parameter1"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.0.value", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.1.name", "Parameter2"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.1.value", "Value2"),
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

func TestAccImageBuilderContainerRecipe_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
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

func TestAccImageBuilderContainerRecipe_dockerfileTemplateURI(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDockerfileTemplateURIConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "dockerfile_template_data"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dockerfile_template_uri"},
			},
		},
	})
}

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMapping_deviceName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingDeviceNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.device_name", "/dev/xvda"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_deleteOnTermination(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSDeleteOnTerminationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.delete_on_termination", "true"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_encrypted(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSEncryptedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.encrypted", "true"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_iops(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSIOPSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.iops", "100"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSKMSKeyIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_snapshotID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ebsSnapshotResourceName := "aws_ebs_snapshot.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSSnapshotIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.snapshot_id", ebsSnapshotResourceName, "id"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_volumeSize(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSVolumeSizeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.volume_size", "20"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_volumeType(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSVolumeTypeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.volume_type", imagebuilder.EbsVolumeTypeGp2),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMapping_noDevice(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingNoDeviceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.no_device", "true"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMapping_virtualName(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationBlockDeviceMappingVirtualNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.virtual_name", "ephemeral0"),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_Image(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageDataSourceName := "data.aws_ami.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeInstanceConfigurationImageConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.image", imageDataSourceName, "id"),
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

func TestAccImageBuilderContainerRecipe_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeKmsKeyIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccImageBuilderContainerRecipe_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
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
				Config: testAccContainerRecipeTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccContainerRecipeTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccImageBuilderContainerRecipe_workingDirectory(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeWorkingDirectoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "working_directory", "/tmp"),
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

func testAccCheckContainerRecipeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_container_recipe" {
			continue
		}

		input := &imagebuilder.GetContainerRecipeInput{
			ContainerRecipeArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetContainerRecipe(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Container Recipe (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Container Recipe (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckContainerRecipeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

		input := &imagebuilder.GetContainerRecipeInput{
			ContainerRecipeArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetContainerRecipe(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Container Recipe (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccContainerRecipeBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_repository" "test" {
  name = %[1]q
}

`, rName)
}

func testAccContainerRecipeNameConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name           = %[1]q
  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}
`, rName))
}

func testAccContainerRecipeComponentConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name           = %[1]q
  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/aws-cli-version-2-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}
`, rName))
}

func testAccContainerRecipeComponentParameterConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_component" "test" {
  data = <<EOF
phases:
  - name: build
    steps:
      - name: example
        action: ExecuteBash
        inputs:
          commands:
            - echo {{ Parameter1 }}
            - echo {{ Parameter2 }}
parameters:
  - Parameter1:
      type: string
  - Parameter2:
      type: string  
schemaVersion: 1.0
EOF

  name     = %[1]q
  platform = "Linux"
  version  = "1.0.0"
}

resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = aws_imagebuilder_component.test.arn

    parameter {
      name  = "Parameter1"
      value = "Value1"
    }

    parameter {
      name  = "Parameter2"
      value = "Value2"
    }
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}
`, rName))
}

func testAccContainerRecipeDescriptionConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name        = %[1]q
  description = "description"

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}
`, rName))
}

func testAccContainerDockerfileTemplateURIConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = "Dockerfile"
  content = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF
}

resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  depends_on = [
    aws_s3_object.test
  ]
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingDeviceNameConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      device_name = "/dev/xvda"
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSDeleteOnTerminationConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      ebs {
        delete_on_termination = true
      }
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSEncryptedConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      ebs {
        encrypted = true
      }
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSIOPSConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      ebs {
        iops = 100
      }
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSKMSKeyIDConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      ebs {
        kms_key_id = aws_kms_key.test.arn
      }
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSSnapshotIDConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}

resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      ebs {
        snapshot_id = aws_ebs_snapshot.test.id
      }
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSVolumeSizeConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      ebs {
        volume_size = 20
      }
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingEBSVolumeTypeConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      ebs {
        volume_type = "gp2"
      }
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingNoDeviceConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      no_device = true
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationBlockDeviceMappingVirtualNameConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    block_device_mapping {
      virtual_name = "ephemeral0"
    }
  }
}
`, rName))
}

func testAccContainerRecipeInstanceConfigurationImageConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  instance_configuration {
    image = data.aws_ami.test.id
  }
}
`, rName))
}

func testAccContainerRecipeKmsKeyIDConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  kms_key_id = aws_kms_key.test.arn
}
`, rName))
}

func testAccContainerRecipeTags1Config(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccContainerRecipeTags2Config(rName string, tagKey1 string, tagValue1, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccContainerRecipeWorkingDirectoryConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name = %[1]q

  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
  version        = "1.0.0"

  component {
    component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }

  working_directory = "/tmp"
}
`, rName))
}
