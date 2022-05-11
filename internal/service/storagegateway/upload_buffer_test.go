package storagegateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
)

func TestDecodeStorageGatewayUploadBufferID(t *testing.T) {
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
		gatewayARN, diskID, err := tfstoragegateway.DecodeUploadBufferID(tc.Input)
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_upload_buffer.test"
	localDiskDataSourceName := "data.aws_storagegateway_local_disk.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// Storage Gateway API does not support removing upload buffers,
		// but we want to ensure other resources are removed.
		CheckDestroy: testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUploadBufferDiskIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadBufferExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "disk_id", localDiskDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "disk_path", localDiskDataSourceName, "disk_path"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17809
func TestAccStorageGatewayUploadBuffer_diskPath(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_upload_buffer.test"
	localDiskDataSourceName := "data.aws_storagegateway_local_disk.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		// Storage Gateway API does not support removing upload buffers,
		// but we want to ensure other resources are removed.
		CheckDestroy: testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUploadBufferDiskPathConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUploadBufferExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "disk_id", regexp.MustCompile(`.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "disk_path", localDiskDataSourceName, "disk_path"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
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

func testAccCheckUploadBufferExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

		gatewayARN, diskID, err := tfstoragegateway.DecodeUploadBufferID(rs.Primary.ID)
		if err != nil {
			return err
		}

		foundDiskID, err := tfstoragegateway.FindUploadBufferDisk(conn, gatewayARN, diskID)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway Upload Buffer (%s): %w", rs.Primary.ID, err)
		}

		if foundDiskID == nil {
			return fmt.Errorf("Storage Gateway Upload Buffer (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccUploadBufferDiskIDConfig(rName string) string {
	return testAccGatewayConfig_GatewayType_Stored(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccUploadBufferDiskPathConfig(rName string) string {
	return testAccGatewayConfig_GatewayType_Cached(rName) + fmt.Sprintf(`
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
`, rName)
}
