// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2DedicatedIPPool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDedicatedIPPool(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ses", regexache.MustCompile(`dedicated-ip-pool/.+`)),
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

func TestAccSESV2DedicatedIPPool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDedicatedIPPool(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsesv2.ResourceDedicatedIPPool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2DedicatedIPPool_scalingMode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDedicatedIPPool(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolConfig_scalingMode(rName, string(types.ScalingModeManaged)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_mode", string(types.ScalingModeManaged)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDedicatedIPPoolConfig_scalingMode(rName, string(types.ScalingModeStandard)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_mode", string(types.ScalingModeStandard)),
				),
			},
		},
	})
}

func TestAccSESV2DedicatedIPPool_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDedicatedIPPool(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
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
				Config: testAccDedicatedIPPoolConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDedicatedIPPoolConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckDedicatedIPPoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_dedicated_ip_pool" {
				continue
			}

			_, err := tfsesv2.FindDedicatedIPPoolByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.NotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameDedicatedIPPool, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDedicatedIPPoolExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameDedicatedIPPool, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameDedicatedIPPool, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindDedicatedIPPoolByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameDedicatedIPPool, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheckDedicatedIPPool(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

	_, err := conn.ListDedicatedIpPools(ctx, &sesv2.ListDedicatedIpPoolsInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccDedicatedIPPoolConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_dedicated_ip_pool" "test" {
  pool_name = %[1]q
}
`, rName)
}

func testAccDedicatedIPPoolConfig_scalingMode(rName, scalingMode string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_dedicated_ip_pool" "test" {
  pool_name    = %[1]q
  scaling_mode = %[2]q
}
`, rName, scalingMode)
}

func testAccDedicatedIPPoolConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_dedicated_ip_pool" "test" {
  pool_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDedicatedIPPoolConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_dedicated_ip_pool" "test" {
  pool_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
