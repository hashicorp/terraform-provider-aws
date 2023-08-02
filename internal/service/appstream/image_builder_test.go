// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/appstream"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_basic(instanceType, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
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
	imageName := "AppStream-WinServer2019-07-12-2022"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_withIAMRole(rName, imageName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
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
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_complete(rName, description, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceType),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "state", appstream.ImageBuilderStateRunning),
					resource.TestCheckResourceAttr(resourceName, "instance_type", instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, "description", descriptionUpdated),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
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
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccImageBuilderConfig_tags1(instanceType, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"image_name"},
			},
			{
				Config: testAccImageBuilderConfig_tags2(instanceType, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccImageBuilderConfig_tags1(instanceType, rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageBuilderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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
	imageName := "AppStream-WinServer2019-07-12-2022"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageBuilderDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, appstream.EndpointsID),
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn(ctx)

		_, err := tfappstream.FindImageBuilderByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckImageBuilderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn(ctx)

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
  image_name    = "AppStream-WinServer2019-07-12-2022"
  instance_type = %[1]q
  name          = %[2]q
}
`, instanceType, rName)
}

func testAccImageBuilderConfig_complete(rName, description, instanceType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_appstream_image_builder" "test" {
  image_name                     = "AppStream-WinServer2019-07-12-2022"
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
  image_name    = "AppStream-WinServer2019-07-12-2022"
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
  image_name    = "AppStream-WinServer2019-07-12-2022"
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
