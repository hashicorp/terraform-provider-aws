// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
)

func TestAccSSOAdminInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var instance ssoadmin.DescribeInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccSSOAdminInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var instance ssoadmin.DescribeInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceInstance, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSSOAdminInstance_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var instance ssoadmin.DescribeInstanceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccInstanceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccInstanceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(ctx, resourceName, &instance),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_instance" {
				continue
			}

			_, err := tfssoadmin.FindInstanceByARN(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameInstance, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckInstanceExists(ctx context.Context, name string, instance *ssoadmin.DescribeInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameInstance, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameInstance, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)
		resp, err := tfssoadmin.FindInstanceByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameInstance, rs.Primary.ID, err)
		}

		*instance = *resp

		return nil
	}
}

func testAccInstanceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssoadmin_instance" "test" {
  name             = %[1]q
}
`, rName)
}

func testAccInstanceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ssoadmin_instance" "test" {
  name             = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccInstanceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ssoadmin_instance" "test" {
  name             = %[1]q

  %[2]q = %[3]q
  %[4]q = %[5]q
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
