// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// imageVersionBaseImageEnvVar is the environment variable which must be
// set to an ECR image URI for certain acceptance tests to run
//
// Follow this guide to set up a private ECR repository and push a simple
// "hello world" image to it:
// https://docs.aws.amazon.com/AmazonECR/latest/userguide/getting-started-cli.html
const imageVersionBaseImageEnvVar = "SAGEMAKER_IMAGE_VERSION_BASE_IMAGE"

func TestAccSageMakerImageVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
					resource.TestCheckResourceAttr(resourceName, "horovod", acctest.CtFalse),
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

func TestAccSageMakerImageVersion_update(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdate := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_full(rName, baseImage, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
					resource.TestCheckResourceAttr(resourceName, "horovod", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "processor", "CPU"),
					resource.TestCheckResourceAttr(resourceName, "vendor_guidance", "STABLE"),
					resource.TestCheckResourceAttr(resourceName, "release_notes", rName),
					resource.TestCheckResourceAttr(resourceName, "job_type", "TRAINING"),
					resource.TestCheckResourceAttr(resourceName, "ml_framework", "TensorFlow 1.1"),
					resource.TestCheckResourceAttr(resourceName, "programming_lang", "Python 3.8"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageVersionConfig_full(rName, baseImage, rNameUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
					resource.TestCheckResourceAttr(resourceName, "horovod", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "processor", "CPU"),
					resource.TestCheckResourceAttr(resourceName, "vendor_guidance", "STABLE"),
					resource.TestCheckResourceAttr(resourceName, "release_notes", rNameUpdate),
					resource.TestCheckResourceAttr(resourceName, "job_type", "TRAINING"),
					resource.TestCheckResourceAttr(resourceName, "ml_framework", "TensorFlow 1.1"),
					resource.TestCheckResourceAttr(resourceName, "programming_lang", "Python 3.8"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerImageVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceImageVersion(), resourceName),
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

func TestAccSageMakerImageVersion_Disappears_image(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	imageResourceName := "aws_sagemaker_image.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceImage(), imageResourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(imageResourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// TestAccSageMakerImageVersion_multiple verifies multiple image versions by the
// same name can co-exist in the same configuration without overwriting one another
//
// Ref: https://github.com/hashicorp/terraform-provider-aws/issues/40597
func TestAccSageMakerImageVersion_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	var image, imageV2 sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	resourceNameV2 := "aws_sagemaker_image_version.test_v2"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_multiple(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					testAccCheckImageVersionExists(ctx, resourceNameV2, &imageV2),
					resource.TestCheckResourceAttr(resourceNameV2, "image_name", rName),
					resource.TestCheckResourceAttr(resourceNameV2, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceNameV2, names.AttrVersion, "2"),
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

func TestAccSageMakerImageVersion_aliases(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_aliases(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "latest"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "stable"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "0"),
				),
			},
			{
				Config: testAccImageVersionConfig_aliases(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, "aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "latest"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aliases.*", "stable"),
				),
			},
		},
	})
}

func TestAccSageMakerImageVersion_upgrade_V5_98_0(t *testing.T) {
	ctx := acctest.Context(t)

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := acctest.SkipIfEnvVarNotSet(t, imageVersionBaseImageEnvVar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.SageMakerServiceID),
		CheckDestroy: testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// After v5.97.0, id was change to a multi-part key
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.97.0",
					},
				},
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckImageVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_image_version" {
				continue
			}

			name := rs.Primary.Attributes["image_name"]
			version, err := strconv.Atoi(rs.Primary.Attributes[names.AttrVersion])
			if err != nil {
				return fmt.Errorf("reading SageMaker AI Image Version (%s): %w", rs.Primary.ID, err)
			}

			_, err = tfsagemaker.FindImageVersionByTwoPartKey(ctx, conn, name, version)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker AI Image Version (%s): %w", rs.Primary.ID, err)
			}

			return fmt.Errorf("SageMaker AI Image Version %q still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckImageVersionExists(ctx context.Context, n string, image *sagemaker.DescribeImageVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		name := rs.Primary.Attributes["image_name"]
		version, err := strconv.Atoi(rs.Primary.Attributes[names.AttrVersion])
		if err != nil {
			return fmt.Errorf("reading SageMaker AI Image Version (%s): %w", rs.Primary.ID, err)
		}

		if name == "" || version == 0 {
			return fmt.Errorf("image_name or version not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindImageVersionByTwoPartKey(ctx, conn, name, version)
		if err != nil {
			return err
		}

		*image = *resp

		return nil
	}
}

func testAccImageVersionConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFullAccess"
}

resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccImageVersionConfig_basic(rName, baseImage string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[1]q
}
`, baseImage))
}

func testAccImageVersionConfig_full(rName, baseImage, notes string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name       = aws_sagemaker_image.test.id
  base_image       = %[1]q
  job_type         = "TRAINING"
  processor        = "CPU"
  release_notes    = %[2]q
  vendor_guidance  = "STABLE"
  ml_framework     = "TensorFlow 1.1"
  programming_lang = "Python 3.8"
}
`, baseImage, notes))
}

func testAccImageVersionConfig_multiple(rName, baseImage string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[1]q
}

resource "aws_sagemaker_image_version" "test_v2" {
  image_name = aws_sagemaker_image_version.test.image_name
  base_image = %[1]q
}
`, baseImage))
}

func testAccImageVersionConfig_aliases(rName, baseImage string) string {
	return acctest.ConfigCompose(
		testAccImageVersionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[1]q
  aliases    = ["latest", "stable"]
}
`, baseImage))
}
