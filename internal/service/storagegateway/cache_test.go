// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestCacheParseResourceID(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		Input              string
		ExpectedGatewayARN string
		ExpectedDiskID     string
		ErrCount           int
	}{
		{
			Input:              "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0", //lintignore:AWSAT003,AWSAT005
			ExpectedGatewayARN: "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678",                               //lintignore:AWSAT003,AWSAT005
			ExpectedDiskID:     "pci-0000:03:00.0-scsi-0:0:0:0",
			ErrCount:           0,
		},
		{
			Input:    "sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0",
			ErrCount: 1,
		},
		{
			Input:    "example:pci-0000:03:00.0-scsi-0:0:0:0",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678", //lintignore:AWSAT003,AWSAT005
			ErrCount: 1,
		},
		{
			Input:    "pci-0000:03:00.0-scsi-0:0:0:0",
			ErrCount: 1,
		},
		{
			Input:    "gateway/sgw-12345678",
			ErrCount: 1,
		},
		{
			Input:    "sgw-12345678",
			ErrCount: 1,
		},
	}

	for _, tc := range testCases {
		gatewayARN, diskID, err := tfstoragegateway.CacheParseResourceID(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if gatewayARN != tc.ExpectedGatewayARN {
			t.Fatalf("expected %q to return Gateway ARN %q, received: %s", tc.Input, tc.ExpectedGatewayARN, gatewayARN)
		}
		if diskID != tc.ExpectedDiskID {
			t.Fatalf("expected %q to return Disk ID %q, received: %s", tc.Input, tc.ExpectedDiskID, diskID)
		}
	}
}

func TestAccStorageGatewayCache_fileGateway(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cache.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// Storage Gateway API does not support removing caches.
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCacheConfig_fileGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCacheExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "disk_id"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, names.AttrARN),
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

func TestAccStorageGatewayCache_tapeAndVolumeGateway(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cache.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccCacheConfig_tapeAndVolumeGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCacheExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "disk_id"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, names.AttrARN),
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

func testAccCheckCacheExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		gatewayARN, diskID, err := tfstoragegateway.CacheParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		return tfstoragegateway.FindCacheByTwoPartKey(ctx, conn, gatewayARN, diskID)
	}
}

func testAccCacheConfig_fileGateway(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_typeFileS3(rName, rName), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdb"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.test.id
}

data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_cache" "test" {
  # ACCEPTANCE TESTING WORKAROUND:
  # Data sources are not refreshed before plan after apply in TestStep
  # Step 0 error: After applying this step, the plan was not empty:
  #   disk_id:     "877ee674-99d3-4cd4-99f0-aadae7e3942b" => "/dev/nvme1n1" (forces new resource)
  # We expect this data source value to change due to how Storage Gateway works.

  lifecycle {
    ignore_changes = [disk_id]
  }

  disk_id     = data.aws_storagegateway_local_disk.test.id
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`, rName))
}

func testAccCacheConfig_tapeAndVolumeGateway(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_typeCached(rName), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.test.id
}

data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_cache" "test" {
  # ACCEPTANCE TESTING WORKAROUND:
  # Data sources are not refreshed before plan after apply in TestStep
  # Step 0 error: After applying this step, the plan was not empty:
  #   disk_id:     "0b68f77a-709b-4c79-ad9d-d7728014b291" => "/dev/xvdc" (forces new resource)
  # We expect this data source value to change due to how Storage Gateway works.

  lifecycle {
    ignore_changes = ["disk_id"]
  }

  disk_id     = data.aws_storagegateway_local_disk.test.id
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`, rName))
}
