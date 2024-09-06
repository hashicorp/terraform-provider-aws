// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRAMResourceShare_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceShare awstypes.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ram", regexache.MustCompile(`resource-share/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "permission_arns.#", acctest.Ct0),
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

func TestAccRAMResourceShare_permission(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceShare awstypes.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareConfig_namePermission(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ram", regexache.MustCompile(`resource-share/.+`)),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "permission_arns.#", acctest.Ct1),
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

func TestAccRAMResourceShare_allowExternalPrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceShare1, resourceShare2 awstypes.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareConfig_allowExternalPrincipals(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceShareConfig_allowExternalPrincipals(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, "allow_external_principals", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccRAMResourceShare_name(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceShare1, resourceShare2 awstypes.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccResourceShareConfig_name(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			},
		},
	})
}

func TestAccRAMResourceShare_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceShare1, resourceShare2, resourceShare3 awstypes.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare1),
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
				Config: testAccResourceShareConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccResourceShareConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccRAMResourceShare_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceShare awstypes.ResourceShare
	resourceName := "aws_ram_resource_share.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceShareDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceShareConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceShareExists(ctx, resourceName, &resourceShare),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfram.ResourceResourceShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckResourceShareExists(ctx context.Context, n string, v *awstypes.ResourceShare) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		output, err := tfram.FindResourceShareOwnerSelfByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckResourceShareDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ram_resource_share" {
				continue
			}

			_, err := tfram.FindResourceShareOwnerSelfByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("RAM Resource Share %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceShareConfig_allowExternalPrincipals(rName string, allowExternalPrincipals bool) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  allow_external_principals = %[1]t
  name                      = %[2]q
}
`, allowExternalPrincipals, rName)
}

func testAccResourceShareConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q
}
`, rName)
}

func testAccResourceShareConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccResourceShareConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ram_resource_share" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccResourceShareConfig_namePermission(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_ram_resource_share" "test" {
  name            = %[1]q
  permission_arns = ["arn:${data.aws_partition.current.partition}:ram::aws:permission/AWSRAMBlankEndEntityCertificateAPICSRPassthroughIssuanceCertificateAuthority"]
}
`, rName)
}
