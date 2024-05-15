// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglobalaccelerator "github.com/hashicorp/terraform-provider-aws/internal/service/globalaccelerator"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlobalAcceleratorCrossAccountAttachment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v awstypes.Attachment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrSet(resourceName, "last_modified_time"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
					resource.TestCheckResourceAttr(resourceName, "principals.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "resource.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}
func TestAccGlobalAcceleratorCrossAccountAttachment_principals(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rAccountID1 := sdkacctest.RandStringFromCharSet(12, "012346789")
	rAccountID2 := sdkacctest.RandStringFromCharSet(12, "012346789")
	var v awstypes.Attachment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_principals(rName, rAccountID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "principals.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "principals.*", rAccountID1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrossAccountAttachmentConfig_principals(rName, rAccountID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "principals.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "principals.*", rAccountID2),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorCrossAccountAttachment_resources(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_resources(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "resource.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCrossAccountAttachmentConfig_resourcesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "resource.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccGlobalAcceleratorCrossAccountAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v awstypes.Attachment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCrossAccountAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfglobalaccelerator.ResourceCrossAccountAttachment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlobalAcceleratorCrossAccountAttachment_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_cross_account_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var v awstypes.Attachment

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAcceleratorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCrossAccountAttachmentConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
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
				Config: testAccCrossAccountAttachmentConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCrossAccountAttachmentConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCrossAccountAttachmentExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckCrossAccountAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_globalaccelerator_cross_account_attachment" {
				continue
			}

			_, err := tfglobalaccelerator.FindCrossAccountAttachmentByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Global Accelerator Cross-account Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCrossAccountAttachmentExists(ctx context.Context, n string, v *awstypes.Attachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlobalAcceleratorClient(ctx)

		output, err := tfglobalaccelerator.FindCrossAccountAttachmentByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCrossAccountAttachmentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name = %[1]q
}
`, rName)
}

func testAccCrossAccountAttachmentConfig_principals(rName, accountID string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name       = %[1]q
  principals = [%[2]q]
}
`, rName, accountID)
}

func testAccCrossAccountAttachmentConfig_resources(rName string) string {
	return acctest.ConfigCompose(testAccEndpointGroupConfig_baseALB(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name = %[1]q

  resource {
    endpoint_id = aws_lb.test.id
  }
}
`, rName))
}

func testAccCrossAccountAttachmentConfig_resourcesUpdated(rName string) string {
	return acctest.ConfigCompose(testAccEndpointGroupConfig_baseALB(rName), fmt.Sprintf(`
resource "aws_eip" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name = %[1]q

  resource {
    endpoint_id = aws_eip.test.arn
  }
}
`, rName))
}

func testAccCrossAccountAttachmentConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCrossAccountAttachmentConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_cross_account_attachment" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
