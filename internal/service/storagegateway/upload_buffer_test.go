// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package storagegateway_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestUploadBufferParseResourceID(t *testing.T) {
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
		gatewayARN, diskID, err := tfstoragegateway.UploadBufferParseResourceID(tc.Input)
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

func TestAccStorageGatewayUploadBuffer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_upload_buffer.test"
	localDiskDataSourceName := "data.aws_storagegateway_local_disk.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// Storage Gateway API does not support removing upload buffers.
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccUploadBufferConfig_diskID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadBufferExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "disk_id", localDiskDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "disk_path", localDiskDataSourceName, "disk_path"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17809.
func TestAccStorageGatewayUploadBuffer_diskPath(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_upload_buffer.test"
	localDiskDataSourceName := "data.aws_storagegateway_local_disk.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccUploadBufferConfig_diskPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadBufferExists(ctx, t, resourceName),
					resource.TestMatchResourceAttr(resourceName, "disk_id", regexache.MustCompile(`.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "disk_path", localDiskDataSourceName, "disk_path"),
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

func testAccCheckUploadBufferExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		gatewayARN, diskID, err := tfstoragegateway.UploadBufferParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = tfstoragegateway.FindUploadBufferDiskIDByTwoPartKey(ctx, conn, gatewayARN, diskID)

		return err
	}
}

func testAccUploadBufferConfig_diskID(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_typeStored(rName), fmt.Sprintf(`
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

resource "aws_storagegateway_upload_buffer" "test" {
  disk_id     = data.aws_storagegateway_local_disk.test.id
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`, rName))
}

func testAccUploadBufferConfig_diskPath(rName string) string {
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

resource "aws_storagegateway_upload_buffer" "test" {
  disk_path   = data.aws_storagegateway_local_disk.test.disk_path
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`, rName))
}
