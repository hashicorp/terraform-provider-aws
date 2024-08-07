// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
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

func TestAccSageMakerImage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("image/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
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

func TestAccSageMakerImage_description(t *testing.T) {
	ctx := acctest.Context(t)
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccImageConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
		},
	})
}

func TestAccSageMakerImage_displayName(t *testing.T) {
	ctx := acctest.Context(t)
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_displayName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, ""),
				),
			},
			{
				Config: testAccImageConfig_displayName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
				),
			},
		},
	})
}

func TestAccSageMakerImage_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
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
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccImageConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerImage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(ctx, resourceName, &image),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceImage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckImageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_image" {
				continue
			}

			Image, err := tfsagemaker.FindImageByName(ctx, conn, rs.Primary.ID)

			if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Image (%s): %w", rs.Primary.ID, err)
			}

			if aws.StringValue(Image.ImageName) == rs.Primary.ID {
				return fmt.Errorf("sagemaker Image %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckImageExists(ctx context.Context, n string, image *sagemaker.DescribeImageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Image ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		resp, err := tfsagemaker.FindImageByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*image = *resp

		return nil
	}
}

func testAccImageBaseConfig(rName string) string {
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
`, rName)
}

func testAccImageConfig_basic(rName string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}
`, rName)
}

func testAccImageConfig_description(rName string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name  = %[1]q
  role_arn    = aws_iam_role.test.arn
  description = %[1]q
}
`, rName)
}

func testAccImageConfig_displayName(rName string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name   = %[1]q
  role_arn     = aws_iam_role.test.arn
  display_name = %[1]q
}
`, rName)
}

func testAccImageConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccImageConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
