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

func TestAccImageBuilderContainerRecipe_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "imagebuilder", regexache.MustCompile(fmt.Sprintf("container-recipe/%s/1.0.0", rName))),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct1),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "component.0.component_arn", "imagebuilder", "aws", "component/update-linux/x.x.x"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "container_type", "DOCKER"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttrSet(resourceName, "dockerfile_template_data"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEncrypted, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwner),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "parent_image", "imagebuilder", "aws", "image/amazon-linux-x86-2/x.x.x"),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target_repository.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target_repository.0.repository_name", "aws_ecr_repository.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "target_repository.0.service", "ECR"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1.0.0"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfimagebuilder.ResourceContainerRecipe(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderContainerRecipe_component(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_component(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_componentParameter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_dockerfileTemplateURI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingDeviceName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSDeleteOnTermination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.delete_on_termination", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.encrypted", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSIOPS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.kms_key_id", kmsKeyResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ebsSnapshotResourceName := "aws_ebs_snapshot.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSSnapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.snapshot_id", ebsSnapshotResourceName, names.AttrID),
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

func TestAccImageBuilderContainerRecipe_InstanceConfiguration_BlockDeviceMappingEBS_throughput(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSThroughput(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.0.throughput", "200"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSVolumeSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSVolumeType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.ebs.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingNoDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.0.no_device", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingVirtualName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.0.block_device_mapping.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageDataSourceName := "data.aws_ami.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_instanceConfigurationImage(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "instance_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "instance_configuration.0.image", imageDataSourceName, names.AttrID),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
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
				Config: testAccContainerRecipeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccContainerRecipeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderContainerRecipe_workingDirectory(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeConfig_workingDirectory(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
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

func TestAccImageBuilderContainerRecipe_platformOverride(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// A public ecr image can only be used if platform override is set, so this test will only pass if it is set correctly
				Config: testAccContainerRecipeConfig_platformOverride(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform", "Linux"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"platform_override"},
			},
		},
	})
}

func testAccCheckContainerRecipeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_container_recipe" {
				continue
			}

			input := &imagebuilder.GetContainerRecipeInput{
				ContainerRecipeArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetContainerRecipeWithContext(ctx, input)

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
}

func testAccCheckContainerRecipeExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		input := &imagebuilder.GetContainerRecipeInput{
			ContainerRecipeArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetContainerRecipeWithContext(ctx, input)

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

func testAccContainerRecipeConfig_name(rName string) string {
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

func testAccContainerRecipeConfig_component(rName string) string {
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

func testAccContainerRecipeConfig_componentParameter(rName string) string {
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

func testAccContainerRecipeConfig_description(rName string) string {
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

func testAccContainerRecipeConfig_dockerfileTemplateURI(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingDeviceName(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSDeleteOnTermination(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSEncrypted(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSIOPS(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSKMSKeyID(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSSnapshotID(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSThroughput(rName string, throughput int) string {
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
        throughput  = %[2]d
        volume_type = "gp3"
      }
    }
  }
}
`, rName, throughput))
}

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSVolumeSize(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingEBSVolumeType(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingNoDevice(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationBlockDeviceMappingVirtualName(rName string) string {
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

func testAccContainerRecipeConfig_instanceConfigurationImage(rName string) string {
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

func testAccContainerRecipeConfig_kmsKeyID(rName string) string {
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

func testAccContainerRecipeConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
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

func testAccContainerRecipeConfig_tags2(rName string, tagKey1 string, tagValue1, tagKey2 string, tagValue2 string) string {
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

func testAccContainerRecipeConfig_workingDirectory(rName string) string {
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

func testAccContainerRecipeConfig_platformOverride(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  name              = %[1]q
  container_type    = "DOCKER"
  parent_image      = "public.ecr.aws/amazonlinux/amazonlinux:latest"
  version           = "1.0.0"
  platform_override = "Linux"

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
