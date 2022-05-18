package storagegateway_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/internal/service/storagegateway"
)

func TestParseVolumeGatewayARNAndTargetNameFromARN(t *testing.T) {
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
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachediSCSIVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "gateway_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "target_arn", "storagegateway", regexp.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-.+$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "volume_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.`)),
					resource.TestCheckResourceAttr(resourceName, "volume_size_in_bytes", "5368709120"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
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
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"
	keyResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachediSCSIVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeKMSEncryptedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &cachedIscsiVolume),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key", keyResourceName, "arn"),
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
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachediSCSIVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCachediSCSIVolumeTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCachediSCSIVolumeTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccStorageGatewayCachediSCSIVolume_snapshotID(t *testing.T) {
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachediSCSIVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_SnapshotID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "chap_enabled", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "gateway_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "lun_number", "0"),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "snapshot_id", regexp.MustCompile(`^snap-.+$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "target_arn", "storagegateway", regexp.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-.+$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "volume_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.`)),
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
	var cachedIscsiVolume storagegateway.CachediSCSIVolume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachediSCSIVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_SourceVolumeARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &cachedIscsiVolume),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "gateway_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_id", regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "network_interface_port", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "snapshot_id", ""),
					acctest.MatchResourceAttrRegionalARN(resourceName, "source_volume_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "target_arn", "storagegateway", regexp.MustCompile(fmt.Sprintf(`gateway/sgw-.+/target/iqn.1997-05.com.amazon:%s`, rName))),
					resource.TestCheckResourceAttr(resourceName, "target_name", rName),
					resource.TestMatchResourceAttr(resourceName, "volume_id", regexp.MustCompile(`^vol-.+$`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "volume_arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+/volume/vol-.`)),
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
	var storedIscsiVolume storagegateway.CachediSCSIVolume
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_storagegateway_cached_iscsi_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, storagegateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckCachediSCSIVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCachediSCSIVolumeConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCachediSCSIVolumeExists(resourceName, &storedIscsiVolume),
					acctest.CheckResourceDisappears(acctest.Provider, tfstoragegateway.ResourceCachediSCSIVolume(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCachediSCSIVolumeExists(resourceName string, cachedIscsiVolume *storagegateway.CachediSCSIVolume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

		input := &storagegateway.DescribeCachediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeCachediSCSIVolumes(input)

		if err != nil {
			return fmt.Errorf("error reading Storage Gateway cached iSCSI volume: %s", err)
		}

		if output == nil || len(output.CachediSCSIVolumes) == 0 || output.CachediSCSIVolumes[0] == nil || aws.StringValue(output.CachediSCSIVolumes[0].VolumeARN) != rs.Primary.ID {
			return fmt.Errorf("Storage Gateway cached iSCSI volume %q not found", rs.Primary.ID)
		}

		*cachedIscsiVolume = *output.CachediSCSIVolumes[0]

		return nil
	}
}

func testAccCheckCachediSCSIVolumeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).StorageGatewayConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_cached_iscsi_volume" {
			continue
		}

		input := &storagegateway.DescribeCachediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeCachediSCSIVolumes(input)

		if err != nil {
			if tfstoragegateway.IsErrGatewayNotFound(err) {
				return nil
			}
			if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) {
				return nil
			}
			if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
				return nil
			}
			return err
		}

		if output != nil && len(output.CachediSCSIVolumes) > 0 && output.CachediSCSIVolumes[0] != nil && aws.StringValue(output.CachediSCSIVolumes[0].VolumeARN) == rs.Primary.ID {
			return fmt.Errorf("Storage Gateway cached iSCSI volume %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCachediSCSIVolumeBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccGatewayConfig_GatewayType_Cached(rName),
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

func testAccCachediSCSIVolumeConfig_Basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_storagegateway_cached_iscsi_volume" "test" {
  gateway_arn          = aws_storagegateway_cache.test.gateway_arn
  network_interface_id = aws_instance.test.private_ip
  target_name          = %[1]q
  volume_size_in_bytes = 5368709120
}
`, rName))
}

func testAccCachediSCSIVolumeKMSEncryptedConfig(rName string) string {
	return testAccCachediSCSIVolumeBaseConfig(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test %[1]s"
  policy      = <<POLICY
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
`, rName)
}

func testAccCachediSCSIVolumeTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeBaseConfig(rName),
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

func testAccCachediSCSIVolumeTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeBaseConfig(rName),
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

func testAccCachediSCSIVolumeConfig_SnapshotID(rName string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeBaseConfig(rName),
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

func testAccCachediSCSIVolumeConfig_SourceVolumeARN(rName string) string {
	return acctest.ConfigCompose(
		testAccCachediSCSIVolumeBaseConfig(rName),
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
