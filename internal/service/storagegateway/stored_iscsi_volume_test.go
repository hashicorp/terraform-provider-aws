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

func TestAccStorageGatewayStorediSCSIVolume_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var storedIscsiVolume awstypes.StorediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorediSCSIVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					resource.TestCheckResourceAttr(resourceName, "preserve_existing_data", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "disk_id", "data.aws_storagegateway_local_disk.test", names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", "aws_storagegateway_gateway.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, "aws_instance.test", "private_ip"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrTargetARN, "storagegateway", regexache.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexache.MustCompile(`^vol-+`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "10737418240"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtFalse),
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

func TestAccStorageGatewayStorediSCSIVolume_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var storedIscsiVolume awstypes.StorediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"
	keyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorediSCSIVolumeConfig_kmsEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKey, keyResourceName, names.AttrARN),
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

func TestAccStorageGatewayStorediSCSIVolume_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var storedIscsiVolume awstypes.StorediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorediSCSIVolumeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
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
				Config: testAccStorediSCSIVolumeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccStorediSCSIVolumeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccStorageGatewayStorediSCSIVolume_snapshotID(t *testing.T) {
	ctx := acctest.Context(t)
	var storedIscsiVolume awstypes.StorediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorediSCSIVolumeConfig_snapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", "aws_storagegateway_gateway.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, "aws_instance.test", "private_ip"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSnapshotID, "aws_ebs_snapshot.test", names.AttrID),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrTargetARN, "storagegateway", regexache.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexache.MustCompile(`^vol-+`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "10737418240"),
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

func TestAccStorageGatewayStorediSCSIVolume_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var storedIscsiVolume awstypes.StorediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_stored_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStorediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccStorediSCSIVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckStorediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					acctest.CheckSDKResourceDisappears(ctx, t, tfstoragegateway.ResourceStorediSCSIVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckStorediSCSIVolumeExists(ctx context.Context, t *testing.T, n string, v *awstypes.StorediSCSIVolume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		output, err := tfstoragegateway.FindStorediSCSIVolumeByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStorediSCSIVolumeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_storagegateway_stored_iscsi_volume" {
				continue
			}

			_, err := tfstoragegateway.FindStorediSCSIVolumeByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Storage Gateway Stored iSCSI Volume %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStorediSCSIVolumeConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_typeStored(rName), fmt.Sprintf(`
resource "aws_ebs_volume" "buffer" {
  availability_zone = aws_instance.test.availability_zone
  size              = 10
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "buffer" {
  device_name  = "/dev/xvdc"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.buffer.id
}

data "aws_storagegateway_local_disk" "buffer" {
  disk_node   = aws_volume_attachment.buffer.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_storagegateway_working_storage" "buffer" {
  disk_id     = data.aws_storagegateway_local_disk.buffer.id
  gateway_arn = aws_storagegateway_gateway.test.arn
}

resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = 10
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
`, rName))
}

func testAccStorediSCSIVolumeConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStorediSCSIVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.disk_id

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName))
}

func testAccStorediSCSIVolumeConfig_kmsEncrypted(rName string) string {
	return acctest.ConfigCompose(testAccStorediSCSIVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test %[1]s"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id
  kms_encrypted          = true
  kms_key                = aws_kms_key.test.arn

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName))
}

func testAccStorediSCSIVolumeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccStorediSCSIVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName, tagKey1, tagValue1))
}

func testAccStorediSCSIVolumeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccStorediSCSIVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccStorediSCSIVolumeConfig_snapshotID(rName string) string {
	return acctest.ConfigCompose(testAccStorediSCSIVolumeConfig_base(rName), fmt.Sprintf(`
resource "aws_ebs_volume" "snapvolume" {
  availability_zone = aws_instance.test.availability_zone
  size              = 5
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.snapvolume.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_storagegateway_stored_iscsi_volume" "test" {
  gateway_arn            = data.aws_storagegateway_local_disk.test.gateway_arn
  network_interface_id   = aws_instance.test.private_ip
  snapshot_id            = aws_ebs_snapshot.test.id
  target_name            = %[1]q
  preserve_existing_data = false
  disk_id                = data.aws_storagegateway_local_disk.test.id

  depends_on = [aws_storagegateway_working_storage.buffer]
}
`, rName))
}
