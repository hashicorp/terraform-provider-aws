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
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderImageRecipe_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "imagebuilder", regexache.MustCompile(fmt.Sprintf("image-recipe/%s/1.0.0", rName))),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct1),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwner),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "parent_image", "imagebuilder", "aws", "image/amazon-linux-2-x86/x.x.x"),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1.0.0"),
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

func TestAccImageBuilderImageRecipe_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfimagebuilder.ResourceImageRecipe(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderImageRecipe_BlockDeviceMapping_deviceName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingDeviceName(rName, "/dev/xvdb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						names.AttrDeviceName: "/dev/xvdb",
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_deleteOnTermination(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSDeleteOnTermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"ebs.0.delete_on_termination": acctest.CtTrue,
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_encrypted(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSEncrypted(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"ebs.0.encrypted": acctest.CtTrue,
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_iops(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSIOPS(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"ebs.0.iops": "100",
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSKMSKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "block_device_mapping.*.ebs.0.kms_key_id", kmsKeyResourceName, names.AttrARN),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_snapshotID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ebsSnapshotResourceName := "aws_ebs_snapshot.test"
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSSnapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "block_device_mapping.*.ebs.0.snapshot_id", ebsSnapshotResourceName, names.AttrID),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_throughput(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSThroughput(rName, 200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"ebs.0.throughput": "200",
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_volumeSize(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSVolumeSize(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"ebs.0.volume_size": "20",
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_volumeTypeGP2(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSVolumeType(rName, imagebuilder.EbsVolumeTypeGp2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"ebs.0.volume_type": imagebuilder.EbsVolumeTypeGp2,
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMappingEBS_volumeTypeGP3(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingEBSVolumeType(rName, tfimagebuilder.EBSVolumeTypeGP3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"ebs.0.volume_type": tfimagebuilder.EBSVolumeTypeGP3,
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMapping_noDevice(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingNoDevice(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						"no_device": acctest.CtTrue,
					}),
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

func TestAccImageBuilderImageRecipe_BlockDeviceMapping_virtualName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_blockDeviceMappingVirtualName(rName, "ephemeral0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "block_device_mapping.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "block_device_mapping.*", map[string]string{
						names.AttrVirtualName: "ephemeral0",
					}),
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

func TestAccImageBuilderImageRecipe_component(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_component(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct3),
					resource.TestCheckResourceAttrPair(resourceName, "component.0.component_arn", "data.aws_imagebuilder_component.aws-cli-version-2-linux", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "component.1.component_arn", "data.aws_imagebuilder_component.update-linux", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "component.2.component_arn", "aws_imagebuilder_component.test", names.AttrARN),
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

func TestAccImageBuilderImageRecipe_componentParameter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_componentParameter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.0.name", "Parameter1"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.0.value", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.1.name", "Parameter2"),
					resource.TestCheckResourceAttr(resourceName, "component.0.parameter.1.value", "Value2"),
				),
			},
		},
	})
}

func TestAccImageBuilderImageRecipe_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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

func TestAccImageBuilderImageRecipe_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
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
				Config: testAccImageRecipeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccImageRecipeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderImageRecipe_workingDirectory(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_workingDirectory(rName, "/tmp"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
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

func TestAccImageBuilderImageRecipe_pipelineUpdateDependency(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_pipelineDependency(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
				),
			},
			{
				Config: testAccImageRecipeConfig_pipelineDependencyUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccImageBuilderImageRecipe_systemsManagerAgent(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_systemsManagerAgent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "systems_manager_agent.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "systems_manager_agent.0.uninstall_after_build", acctest.CtTrue),
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

func TestAccImageBuilderImageRecipe_updateDependency(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_componentUpdate(rName, "hello world"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "component.0.component_arn", "aws_imagebuilder_component.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageRecipeConfig_componentUpdate(rName, "hello world updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "component.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "component.0.component_arn", "aws_imagebuilder_component.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccImageBuilderImageRecipe_userDataBase64(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_userDataBase64(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "user_data_base64", itypes.Base64EncodeOnce([]byte("hello world"))),
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

func TestAccImageBuilderImageRecipe_windowsBaseImage(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageRecipeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageRecipeConfig_windows(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageRecipeExists(ctx, resourceName),
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

func testAccCheckImageRecipeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_image_recipe" {
				continue
			}

			input := &imagebuilder.GetImageRecipeInput{
				ImageRecipeArn: aws.String(rs.Primary.ID),
			}

			output, err := conn.GetImageRecipeWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Image Builder Image Recipe (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Image Builder Image Recipe (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckImageRecipeExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn(ctx)

		input := &imagebuilder.GetImageRecipeInput{
			ImageRecipeArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetImageRecipeWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image Recipe (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccImageRecipeBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

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
`, rName)
}

func testAccImageRecipeConfig_blockDeviceMappingDeviceName(rName string, deviceName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    device_name = %[2]q
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, deviceName))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSDeleteOnTermination(rName string, deleteOnTermination bool) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      delete_on_termination = %[2]t
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, deleteOnTermination))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSEncrypted(rName string, encrypted bool) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      encrypted = %[2]t
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, encrypted))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSIOPS(rName string, iops int) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      iops = %[2]d
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, iops))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSKMSKeyID(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      kms_key_id = aws_kms_key.test.arn
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSSnapshotID(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}

resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      snapshot_id = aws_ebs_snapshot.test.id
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSThroughput(rName string, throughput int) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      throughput  = %[2]d
      volume_type = "gp3"
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, throughput))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSVolumeSize(rName string, volumeSize int) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      volume_size = %[2]d
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, volumeSize))
}

func testAccImageRecipeConfig_blockDeviceMappingEBSVolumeType(rName string, volumeType string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    ebs {
      volume_type = %[2]q
    }
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, volumeType))
}

func testAccImageRecipeConfig_blockDeviceMappingNoDevice(rName string, noDevice bool) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    no_device = %[2]t
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, noDevice))
}

func testAccImageRecipeConfig_blockDeviceMappingVirtualName(rName string, virtualName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  block_device_mapping {
    virtual_name = %[2]q
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, virtualName))
}

func testAccImageRecipeConfig_component(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
data "aws_imagebuilder_component" "aws-cli-version-2-linux" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/aws-cli-version-2-linux/x.x.x"
}

data "aws_imagebuilder_component" "update-linux" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
}

resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = data.aws_imagebuilder_component.aws-cli-version-2-linux.arn
  }

  component {
    component_arn = data.aws_imagebuilder_component.update-linux.arn
  }

  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName))
}

func testAccImageRecipeConfig_componentUpdate(rName, command string) string {
	return acctest.ConfigCompose(
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo '%[2]s'"]
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
`, rName, command))
}

func testAccImageRecipeConfig_componentParameter(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

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

resource "aws_imagebuilder_image_recipe" "test" {
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

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName)
}

func testAccImageRecipeConfig_description(rName string, description string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  description  = %[2]q
  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName, description))
}

func testAccImageRecipeConfig_name(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName))
}

func testAccImageRecipeConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccImageRecipeConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccImageRecipeConfig_workingDirectory(rName string, workingDirectory string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name              = %[1]q
  parent_image      = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version           = "1.0.0"
  working_directory = %[2]q
}
`, rName, workingDirectory))
}

func testAccImageRecipePipelineDependencyBaseConfig(rName string) string {
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

resource "aws_imagebuilder_component" "test2" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecuteBash"
        inputs = {
          commands = ["echo 'Hej världen!'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = "%[1]s-2"
  platform = "Linux"
  version  = "1.0.0"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
}

resource "aws_imagebuilder_image_pipeline" "test" {
  description                      = "Världens finaste beskrivning."
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
  name                             = %[1]q
}
`, rName)
}

func testAccImageRecipeConfig_pipelineDependency(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipePipelineDependencyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName))
}

func testAccImageRecipeConfig_pipelineDependencyUpdate(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipePipelineDependencyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  component {
    component_arn = aws_imagebuilder_component.test2.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}
`, rName))
}

func testAccImageRecipeConfig_systemsManagerAgent(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"

  systems_manager_agent {
    uninstall_after_build = true
  }
}
`, rName))
}

func testAccImageRecipeConfig_userDataBase64(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name             = %[1]q
  parent_image     = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version          = "1.0.0"
  user_data_base64 = base64encode("hello world")
}
`, rName))
}

func testAccImageRecipeBaseConfigWindows(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_imagebuilder_component" "test" {
  data = yamlencode({
    phases = [{
      name = "build"
      steps = [{
        action = "ExecutePowerShell"
        inputs = {
          commands = ["Write-Host 'hello world'"]
        }
        name      = "example"
        onFailure = "Continue"
      }]
    }]
    schemaVersion = 1.0
  })
  name     = %[1]q
  platform = "Windows"
  version  = "1.0.0"
}
`, rName)
}

func testAccImageRecipeConfig_windows(rName string) string {
	return acctest.ConfigCompose(
		testAccImageRecipeBaseConfigWindows(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = aws_imagebuilder_component.test.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/windows-server-2022-english-full-base-x86/x.x.x"
  version      = "1.0.0"
}
`, rName))
}
