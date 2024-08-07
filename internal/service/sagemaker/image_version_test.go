// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerImageVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					resource.TestCheckResourceAttr(resourceName, "base_image", baseImage),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "image_arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image-version/%s/1", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "container_image"),
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

func TestAccSageMakerImageVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceImageVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSageMakerImageVersion_Disappears_image(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE") == "" {
		t.Skip("Environment variable SAGEMAKER_IMAGE_VERSION_BASE_IMAGE is not set")
	}

	var image sagemaker.DescribeImageVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image_version.test"
	baseImage := os.Getenv("SAGEMAKER_IMAGE_VERSION_BASE_IMAGE")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionConfig_basic(rName, baseImage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageVersionExists(ctx, resourceName, &image),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceImage(), "aws_sagemaker_image.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckImageVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_image_version" {
				continue
			}

			imageVersion, err := tfsagemaker.FindImageVersionByName(ctx, conn, rs.Primary.ID)

			if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Image Version (%s): %w", rs.Primary.ID, err)
			}

			if aws.StringValue(imageVersion.ImageVersionArn) == rs.Primary.Attributes[names.AttrARN] {
				return fmt.Errorf("SageMaker Image Version %q still exists", rs.Primary.ID)
			}
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

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Image ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		resp, err := tfsagemaker.FindImageVersionByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*image = *resp

		return nil
	}
}

func testAccImageVersionConfig_basic(rName, baseImage string) string {
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

resource "aws_sagemaker_image_version" "test" {
  image_name = aws_sagemaker_image.test.id
  base_image = %[2]q
}
`, rName, baseImage)
}
