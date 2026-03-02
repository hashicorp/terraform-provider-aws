// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccStorageGatewayTapePool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var TapePool awstypes.PoolInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_tape_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTapePoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTapePoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTapePoolExists(ctx, t, resourceName, &TapePool),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`tapepool/pool-.+`)),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "GLACIER"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_time_in_days", "0"),
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

func TestAccStorageGatewayTapePool_retention(t *testing.T) {
	ctx := acctest.Context(t)
	var TapePool awstypes.PoolInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_tape_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTapePoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTapePoolConfig_retention(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTapePoolExists(ctx, t, resourceName, &TapePool),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`tapepool/pool-.+`)),
					resource.TestCheckResourceAttr(resourceName, "pool_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStorageClass, "GLACIER"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_type", "GOVERNANCE"),
					resource.TestCheckResourceAttr(resourceName, "retention_lock_time_in_days", "1"),
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

func TestAccStorageGatewayTapePool_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var TapePool awstypes.PoolInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_tape_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTapePoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTapePoolConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTapePoolExists(ctx, t, resourceName, &TapePool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTapePoolConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTapePoolExists(ctx, t, resourceName, &TapePool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTapePoolConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTapePoolExists(ctx, t, resourceName, &TapePool),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccStorageGatewayTapePool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var storedIscsiVolume awstypes.PoolInfo
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_tape_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTapePoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTapePoolConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTapePoolExists(ctx, t, resourceName, &storedIscsiVolume),
					acctest.CheckSDKResourceDisappears(ctx, t, tfstoragegateway.ResourceTapePool(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTapePoolExists(ctx context.Context, t *testing.T, n string, v *awstypes.PoolInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		output, err := tfstoragegateway.FindTapePoolByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTapePoolDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_storagegateway_tape_pool" {
				continue
			}

			_, err := tfstoragegateway.FindTapePoolByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Storage Gateway Tape Pool %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTapePoolConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name     = %[1]q
  storage_class = "GLACIER"
}
`, rName)
}

func testAccTapePoolConfig_retention(rName string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name                   = %[1]q
  storage_class               = "GLACIER"
  retention_lock_type         = "GOVERNANCE"
  retention_lock_time_in_days = 1
}
`, rName)
}

func testAccTapePoolConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name     = %[1]q
  storage_class = "GLACIER"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccTapePoolConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_storagegateway_tape_pool" "test" {
  pool_name     = %[1]q
  storage_class = "GLACIER"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
