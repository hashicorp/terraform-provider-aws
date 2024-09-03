// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffinspace "github.com/hashicorp/terraform-provider-aws/internal/service/finspace"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFinSpaceKxUser_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_basic(rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(ctx, resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, userName),
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

func TestAccFinSpaceKxUser_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_basic(rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(ctx, resourceName, &kxuser),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffinspace.ResourceKxUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFinSpaceKxUser_updateRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_basic(rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(ctx, resourceName, &kxuser),
				),
			},
			{
				Config: testAccKxUserConfig_updateRole(rName, "updated"+rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(ctx, resourceName, &kxuser),
				),
			},
		},
	})
}

func TestAccFinSpaceKxUser_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_tags1(rName, userName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(ctx, resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccKxUserConfig_tags2(rName, userName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(ctx, resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccKxUserConfig_tags1(rName, userName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(ctx, resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckKxUserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_finspace_kx_user" {
				continue
			}

			input := &finspace.GetKxUserInput{
				UserName:      aws.String(rs.Primary.Attributes[names.AttrName]),
				EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
			}
			_, err := conn.GetKxUser(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.FinSpace, create.ErrActionCheckingDestroyed, tffinspace.ResNameKxUser, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKxUserExists(ctx context.Context, name string, kxuser *finspace.GetKxUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxUser, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxUser, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)
		resp, err := conn.GetKxUser(ctx, &finspace.GetKxUserInput{
			UserName:      aws.String(rs.Primary.Attributes[names.AttrName]),
			EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
		})

		if err != nil {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxUser, rs.Primary.ID, err)
		}

		*kxuser = *resp

		return nil
	}
}

func testAccKxUserConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}
`, rName)
}

func testAccKxUserConfig_basic(rName, userName string) string {
	return acctest.ConfigCompose(
		testAccKxUserConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_user" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.test.arn
}
`, userName))
}

func testAccKxUserConfig_updateRole(rName, rName2, userName string) string {
	return acctest.ConfigCompose(
		testAccKxUserConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_role" "updated" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_user" "test" {
  name           = %[2]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.updated.arn
}
`, rName2, userName))
}

func testAccKxUserConfig_tags1(rName, userName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccKxUserConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_user" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.test.arn
  tags = {
    %[2]q = %[3]q
  }
}

`, userName, tagKey1, tagValue1))
}

func testAccKxUserConfig_tags2(rName, userName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccKxUserConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_user" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.test.arn
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, userName, tagKey1, tagValue1, tagKey2, tagValue2))
}
