// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodedeploy "github.com/hashicorp/terraform-provider-aws/internal/service/deploy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDeployApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application1 types.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application1),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codedeploy", fmt.Sprintf(`application:%s`, rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "linked_to_github", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrApplicationID),
				),
			},
			// Import by ID
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Import by name
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployApp_computePlatform(t *testing.T) {
	ctx := acctest.Context(t)
	var application1, application2 types.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_computePlatform(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
				),
			},
			{
				Config: testAccAppConfig_computePlatform(rName, "Server"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application2),
					testAccCheckAppRecreated(&application1, &application2),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
				),
			},
		},
	})
}

func TestAccDeployApp_ComputePlatform_ecs(t *testing.T) {
	ctx := acctest.Context(t)
	var application1 types.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_computePlatform(rName, "ECS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "ECS"),
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

func TestAccDeployApp_ComputePlatform_lambda(t *testing.T) {
	ctx := acctest.Context(t)
	var application1 types.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_computePlatform(rName, "Lambda"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Lambda"),
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

func TestAccDeployApp_name(t *testing.T) {
	ctx := acctest.Context(t)
	var application1, application2 types.ApplicationInfo
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccAppConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
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

func TestAccDeployApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var application types.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
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
				Config: testAccAppConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAppConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDeployApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application1 types.ApplicationInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codedeploy_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodedeploy.ResourceApp(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAppDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DeployClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codedeploy_app" {
				continue
			}

			_, err := tfcodedeploy.FindApplicationByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeDeploy Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppExists(ctx context.Context, n string, v *types.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeployClient(ctx)

		output, err := tfcodedeploy.FindApplicationByName(ctx, conn, rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckAppRecreated(i, j *types.ApplicationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.CreateTime).Equal(aws.ToTime(j.CreateTime)) {
			return errors.New("CodeDeploy Application was not recreated")
		}

		return nil
	}
}

func testAccAppConfig_computePlatform(rName string, computePlatform string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  compute_platform = %[1]q
  name             = %[2]q
}
`, computePlatform, rName)
}

func testAccAppConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAppConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
