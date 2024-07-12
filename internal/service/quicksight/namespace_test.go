// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightNamespace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var namespace quicksight.NamespaceInfoV2
	resourceName := "aws_quicksight_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName, &namespace),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, rName),
					resource.TestCheckResourceAttr(resourceName, "identity_store", quicksight.IdentityStoreQuicksight),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "quicksight", fmt.Sprintf("namespace/%[1]s", rName)),
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

func TestAccQuickSightNamespace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var namespace quicksight.NamespaceInfoV2
	resourceName := "aws_quicksight_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName, &namespace),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceNamespace, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightNamespace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var namespace quicksight.NamespaceInfoV2
	resourceName := "aws_quicksight_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNamespaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName, &namespace),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, rName),
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
				Config: testAccNamespaceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName, &namespace),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccNamespaceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNamespaceExists(ctx, resourceName, &namespace),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckNamespaceExists(ctx context.Context, resourceName string, namespace *quicksight.NamespaceInfoV2) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		output, err := tfquicksight.FindNamespaceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameNamespace, rs.Primary.ID, err)
		}

		*namespace = *output

		return nil
	}
}

func testAccCheckNamespaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_namespace" {
				continue
			}

			output, err := tfquicksight.FindNamespaceByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, quicksight.ErrCodeResourceNotFoundException) {
					return nil
				}
				return err
			}

			if output != nil {
				return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameNamespace, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccNamespaceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_namespace" "test" {
  namespace = %[1]q
}
`, rName)
}

func testAccNamespaceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_namespace" "test" {
  namespace = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccNamespaceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_namespace" "test" {
  namespace = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
