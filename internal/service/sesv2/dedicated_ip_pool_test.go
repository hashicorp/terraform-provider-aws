// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2DedicatedIPPool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDedicatedIPPool(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", regexache.MustCompile(`dedicated-ip-pool/.+`)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDedicatedIPPool(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfsesv2.ResourceDedicatedIPPool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSESV2DedicatedIPPool_scalingMode(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sesv2_dedicated_ip_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDedicatedIPPool(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDedicatedIPPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedIPPoolConfig_scalingMode(rName, string(types.ScalingModeManaged)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDedicatedIPPoolExists(ctx, t, resourceName),
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
					testAccCheckDedicatedIPPoolExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_mode", string(types.ScalingModeStandard)),
				),
			},
		},
	})
}

func testAccCheckDedicatedIPPoolDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_dedicated_ip_pool" {
				continue
			}

			_, err := tfsesv2.FindDedicatedIPPoolByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SESv2 Dedicated IP Pool %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDedicatedIPPoolExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

		_, err := tfsesv2.FindDedicatedIPPoolByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccPreCheckDedicatedIPPool(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).SESV2Client(ctx)

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
