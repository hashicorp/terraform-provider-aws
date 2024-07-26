// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamImageBuilder_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_basic(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.ImageBuilderStateRunning)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
		},
	})
}

func TestAccAppStreamImageBuilder_withIAMRole(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.medium"
	imageName := "AppStream-WinServer2022-06-17-2024"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_withIAMRole(rName, imageName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccAppStreamImageBuilder_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_image_builder.test"
	instanceType := "stream.standard.medium"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_basic(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceImageBuilder(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamImageBuilder_complete(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_image_builder.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_complete(rName, description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.ImageBuilderStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
			{
				Config: testAccImageBuilderConfig_complete(rName, descriptionUpdated, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.ImageBuilderStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
		},
	})
}

func TestAccAppStreamImageBuilder_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_image_builder.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_tags1(instanceType, rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
			{
				Config: testAccImageBuilderConfig_tags2(instanceType, rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccImageBuilderConfig_tags1(instanceType, rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccAppStreamImageBuilder_imageARN(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_image_builder.test"
	// imageName selected from the available AWS Managed AppStream 2.0 Base Images
	// Reference: https://docs.aws.amazon.com/appstream2/latest/developerguide/base-image-version-history.html
	imageName := "AppStream-WinServer2022-06-17-2024"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_byARN(rName, imageName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARNNoAccount(resourceName, "image_arn", "appstream", fmt.Sprintf("image/%s", imageName)),
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

func testAccCheckImageBuilderExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AppStream ImageBuilder ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		_, err := tfappstream.FindImageBuilderByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckImageBuilderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_image_builder" {
				continue
			}

			_, err := tfappstream.FindImageBuilderByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream ImageBuilder %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccImageBuilderConfig_basic(instanceType, rName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2022-06-17-2024"
  instance_type = %[1]q
  name          = %[2]q
}
`, instanceType, rName)
}

func testAccImageBuilderConfig_complete(rName, description, instanceType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name                     = "AppStream-WinServer2022-06-17-2024"
  name                           = %[1]q
  description                    = %[2]q
  enable_default_internet_access = false
  instance_type                  = %[3]q
  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }
}
`, rName, description, instanceType))
}

func testAccImageBuilderConfig_tags1(instanceType, rName, key, value string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2022-06-17-2024"
  instance_type = %[1]q
  name          = %[2]q

  tags = {
    %[3]q = %[4]q
  }
}
`, instanceType, rName, key, value)
}

func testAccImageBuilderConfig_tags2(instanceType, rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name    = "AppStream-WinServer2022-06-17-2024"
  instance_type = %[1]q
  name          = %[2]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, instanceType, rName, key1, value1, key2, value2)
}

func testAccImageBuilderConfig_byARN(rName, imageName, instanceType string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_appstream_image_builder" "test" {
  image_arn     = "arn:${data.aws_partition.current.partition}:appstream:%[1]s::image/%[2]s"
  instance_type = %[3]q
  name          = %[4]q
}
`, acctest.Region(), imageName, instanceType, rName)
}

func testAccImageBuilderConfig_withIAMRole(rName, imageName, instanceType string) string {
	return fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  name          = %[1]q
  instance_type = %[2]q
  iam_role_arn  = aws_iam_role.test.arn
  image_name    = %[3]q
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"
    principals {
      type        = "Service"
      identifiers = ["appstream.amazonaws.com"]
    }
  }
}
`, rName, instanceType, imageName)
}
