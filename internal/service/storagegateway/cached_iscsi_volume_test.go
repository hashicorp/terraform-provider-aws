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

func TestParseVolumeGatewayARNAndTargetNameFromARN(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		Input              string
		ExpectedGatewayARN string
		ExpectedTargetName string
		ErrCount           int
	}{
		{
			Input:              "arn:aws:storagegateway:us-east-2:111122223333:gateway/sgw-12A3456B/target/iqn.1997-05.com.amazon:TargetName", //lintignore:AWSAT003,AWSAT005
			ExpectedGatewayARN: "arn:aws:storagegateway:us-east-2:111122223333:gateway/sgw-12A3456B",                                          //lintignore:AWSAT003,AWSAT005
			ExpectedTargetName: "TargetName",
			ErrCount:           0,
		},
		{
			Input:    "gateway/sgw-12A3456B/target/iqn.1997-05.com.amazon:TargetName",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:storagegateway:us-east-2:111122223333:target/iqn.1997-05.com.amazon:TargetName", //lintignore:AWSAT003,AWSAT005
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678", //lintignore:AWSAT003,AWSAT005
			ErrCount: 1,
		},
		{
			Input:    "TargetName",
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
		gatewayARN, targetName, err := tfstoragegateway.ParseVolumeGatewayARNAndTargetNameFromARN(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if gatewayARN != tc.ExpectedGatewayARN {
			t.Fatalf("expected %q to return Gateway ARN %q, received: %s", tc.Input, tc.ExpectedGatewayARN, gatewayARN)
		}
		if targetName != tc.ExpectedTargetName {
			t.Fatalf("expected %q to return Disk ID %q, received: %s", tc.Input, tc.ExpectedTargetName, targetName)
		}
	}
}

func TestAccStorageGatewayCachediSCSIVolume_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cachedIscsiVolume awstypes.CachediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "gateway_arn", "storagegateway", regexache.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestMatchResourceAttr(resourceName, names.AttrNetworkInterfaceID, regexache.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrTargetARN, "storagegateway", regexache.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexache.MustCompile(`^vol-.+$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "volume_arn", "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "5368709120"),
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

func TestAccStorageGatewayCachediSCSIVolume_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var cachedIscsiVolume awstypes.CachediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"
	keyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_kmsEncrypted(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &cachedIscsiVolume),
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

func TestAccStorageGatewayCachediSCSIVolume_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var cachedIscsiVolume awstypes.CachediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &cachedIscsiVolume),
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
				Config: testAccCachediSCSIVolumeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCachediSCSIVolumeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccStorageGatewayCachediSCSIVolume_snapshotID(t *testing.T) {
	ctx := acctest.Context(t)
	var cachedIscsiVolume awstypes.CachediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_snapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", acctest.CtFalse),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "gateway_arn", "storagegateway", regexache.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestMatchResourceAttr(resourceName, names.AttrNetworkInterfaceID, regexache.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexache.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrSnapshotID, regexache.MustCompile(`^snap-.+$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrTargetARN, "storagegateway", regexache.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexache.MustCompile(`^vol-.+$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "volume_arn", "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "5368709120"),
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

func TestAccStorageGatewayCachediSCSIVolume_sourceVolumeARN(t *testing.T) {
	acctest.Skip(t, "This test can cause Storage Gateway 2.0.10.0 to enter an irrecoverable state during volume deletion.")
	ctx := acctest.Context(t)
	var cachedIscsiVolume awstypes.CachediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_sourceARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "gateway_arn", "storagegateway", regexache.MustCompile(`gateway/sgw-.+`)),
					resource.TestMatchResourceAttr(resourceName, names.AttrNetworkInterfaceID, regexache.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexache.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrSnapshotID, ""),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "source_volume_arn", "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrTargetARN, "storagegateway", regexache.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexache.MustCompile(`^vol-.+$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "volume_arn", "storagegateway", regexache.MustCompile(`gateway/sgw-.+/volume/vol-.`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "1073741824"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"source_volume_arn"},
			},
		},
	})
}

func TestAccStorageGatewayCachediSCSIVolume_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var storedIscsiVolume awstypes.CachediSCSIVolume
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.StorageGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCachediSCSIVolumeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(ctx, t, resourceName, &storedIscsiVolume),
					acctest.CheckSDKResourceDisappears(ctx, t, tfstoragegateway.ResourceCachediSCSIVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCachediSCSIVolumeExists(ctx context.Context, t *testing.T, n string, v *awstypes.CachediSCSIVolume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		output, err := tfstoragegateway.FindCachediSCSIVolumeByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCachediSCSIVolumeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).StorageGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_storagegateway_cached_iscsi_volume" {
				continue
			}

			_, err := tfstoragegateway.FindCachediSCSIVolumeByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Storage Gateway Cached iSCSI Volume %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCachediSCSIVolumeConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccGatewayConfig_typeCached(rName),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = 10
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

func testAccCachediSCSIVolumeConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeConfig_base(rName),
		fmt.Sprintf(`
resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  target_name          = %[1]q
  volume_size_in_bytes = 5368709120
}
`, rName))
}

func testAccCachediSCSIVolumeConfig_kmsEncrypted(rName string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeConfig_base(rName),
		fmt.Sprintf(`
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

resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  target_name          = %[1]q
  volume_size_in_bytes = 5368709120
  kms_encrypted        = true
  kms_key              = aws_kms_key.test.arn
}
`, rName))
}

func testAccCachediSCSIVolumeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeConfig_base(rName),
		fmt.Sprintf(`
resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  target_name          = %[1]q
  volume_size_in_bytes = 5368709120

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccCachediSCSIVolumeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeConfig_base(rName),
		fmt.Sprintf(`
resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  target_name          = %[1]q
  volume_size_in_bytes = 5368709120

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccCachediSCSIVolumeConfig_snapshotID(rName string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeConfig_base(rName),
		fmt.Sprintf(`
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

resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  snapshot_id          = aws_ebs_snapshot.test.id
  target_name          = %[1]q
  volume_size_in_bytes = aws_ebs_snapshot.test.volume_size * 1024 * 1024 * 1024
}
`, rName))
}

func testAccCachediSCSIVolumeConfig_sourceARN(rName string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeConfig_base(rName),
		fmt.Sprintf(`
resource "aws_storagegateway_cached_iscsi_volume" "source" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  target_name          = "%[1]s-source"
  volume_size_in_bytes = 1073741824
}

resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  source_volume_arn    = aws_storagegateway_cached_iscsi_volume.source.arn
  target_name          = %[1]q
  volume_size_in_bytes = aws_storagegateway_cached_iscsi_volume.source.volume_size_in_bytes
}
`, rName))
}
