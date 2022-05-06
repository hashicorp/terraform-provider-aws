package efs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEFSFileSystem_basic(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					resource.TestCheckResourceAttr(resourceName, "performance_mode", "generalPurpose"),
					resource.TestCheckResourceAttr(resourceName, "creation_token", rName),
					resource.TestCheckResourceAttr(resourceName, "number_of_mount_targets", "0"),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeBursting),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "size_in_bytes.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes.0.value"),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes.0.value_in_ia"),
					resource.TestCheckResourceAttrSet(resourceName, "size_in_bytes.0.value_in_standard"),
					acctest.MatchResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFileSystemWithPerformanceModeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem("aws_efs_file_system.test2", &desc),
					resource.TestCheckResourceAttr("aws_efs_file_system.test2", "creation_token", "supercalifragilisticexpialidocious"),
					resource.TestCheckResourceAttr("aws_efs_file_system.test2", "performance_mode", "maxIO"),
				),
			},
		},
	})
}

func TestAccEFSFileSystem_availabilityZoneName(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemAvailabilityZoneNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_id", "data.aws_availability_zones.available", "zone_ids.0"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_name", "data.aws_availability_zones.available", "names.0"),
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

func TestAccEFSFileSystem_tags(t *testing.T) {
	var desc efs.FileSystemDescription
	rName := sdkacctest.RandomWithPrefix("tf-acc-tags")
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
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
				Config: testAccFileSystemTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFileSystemTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFileSystemWithMaxTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "50"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Another", "tag"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag45", "TestTagValue"),
				),
			},
		},
	})
}

func TestAccEFSFileSystem_pagedTags(t *testing.T) {
	var desc efs.FileSystemDescription
	rInt := sdkacctest.RandInt()
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPagedTagsConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "10"),
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

func TestAccEFSFileSystem_kmsKey(t *testing.T) {
	var desc efs.FileSystemDescription
	rInt := sdkacctest.RandInt()
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemWithKMSKeyConfig(rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
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

func TestAccEFSFileSystem_kmsWithoutEncryption(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccFileSystemWithKMSKeyConfig(rInt, false),
				ExpectError: regexp.MustCompile(`encrypted must be set to true when kms_key_id is specified`),
			},
		},
	})
}

func TestAccEFSFileSystem_provisionedThroughputInMibps(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_ProvisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				Config: testAccFileSystemConfig_ProvisionedThroughputInMibps(2.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "2"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
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

func TestAccEFSFileSystem_throughputMode(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_ProvisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				Config: testAccFileSystemConfig_ThroughputMode(efs.ThroughputModeBursting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "0"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeBursting),
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

func TestAccEFSFileSystem_lifecyclePolicy(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemWithLifecyclePolicyConfig(
					"transition_to_ia",
					"invalid_value",
				),
				ExpectError: regexp.MustCompile(`got invalid_value`),
			},
			{
				Config: testAccFileSystemWithLifecyclePolicyConfig(
					"transition_to_ia",
					efs.TransitionToIARulesAfter30Days,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_ia", efs.TransitionToIARulesAfter30Days),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFileSystemWithLifecyclePolicyConfig(
					"transition_to_primary_storage_class",
					efs.TransitionToPrimaryStorageClassRulesAfter1Access,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_primary_storage_class", efs.TransitionToPrimaryStorageClassRulesAfter1Access),
				),
			},
			{
				Config: testAccFileSystemRemovedLifecyclePolicyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", "0"),
				),
			},
			{
				Config: testAccFileSystemWithLifecyclePolicyMultiConfig(
					"transition_to_primary_storage_class",
					efs.TransitionToPrimaryStorageClassRulesAfter1Access,
					"transition_to_ia",
					efs.TransitionToIARulesAfter30Days,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.0.transition_to_primary_storage_class", efs.TransitionToPrimaryStorageClassRulesAfter1Access),
					resource.TestCheckResourceAttr(resourceName, "lifecycle_policy.1.transition_to_ia", efs.TransitionToIARulesAfter30Days),
				),
			},
		},
	})
}

func TestAccEFSFileSystem_disappears(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
					acctest.CheckResourceDisappears(acctest.Provider, tfefs.ResourceFileSystem(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfefs.ResourceFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFileSystemDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_file_system" {
			continue
		}

		_, err := tfefs.FindFileSystemByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EFS file system %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFileSystem(resourceID string, fDesc *efs.FileSystemDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EFS file system ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn

		fs, err := tfefs.FindFileSystemByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*fDesc = *fs

		return nil
	}
}

func testAccFileSystemConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %q
}
`, rName)
}

func testAccFileSystemAvailabilityZoneNameConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  creation_token         = %q
  availability_zone_name = data.aws_availability_zones.available.names[0]
}
`, rName)
}

func testAccFileSystemTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFileSystemTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFileSystemPagedTagsConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name           = "test-efs-%[1]d"
    Another        = "tag"
    Test           = "yes"
    User           = "root"
    Page           = "1"
    Environment    = "prod"
    CostCenter     = "terraform"
    AcceptanceTest = "PagedTags"
    CreationToken  = "radek"
    PerfMode       = "max"
  }
}
`, rInt)
}

func testAccFileSystemWithMaxTagsConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name    = %q
    Another = "tag"
    Tag00   = "TestTagValue"
    Tag01   = "TestTagValue"
    Tag02   = "TestTagValue"
    Tag03   = "TestTagValue"
    Tag04   = "TestTagValue"
    Tag05   = "TestTagValue"
    Tag06   = "TestTagValue"
    Tag07   = "TestTagValue"
    Tag08   = "TestTagValue"
    Tag09   = "TestTagValue"
    Tag10   = "TestTagValue"
    Tag11   = "TestTagValue"
    Tag12   = "TestTagValue"
    Tag13   = "TestTagValue"
    Tag14   = "TestTagValue"
    Tag15   = "TestTagValue"
    Tag16   = "TestTagValue"
    Tag17   = "TestTagValue"
    Tag18   = "TestTagValue"
    Tag19   = "TestTagValue"
    Tag20   = "TestTagValue"
    Tag21   = "TestTagValue"
    Tag22   = "TestTagValue"
    Tag23   = "TestTagValue"
    Tag24   = "TestTagValue"
    Tag25   = "TestTagValue"
    Tag26   = "TestTagValue"
    Tag27   = "TestTagValue"
    Tag28   = "TestTagValue"
    Tag29   = "TestTagValue"
    Tag30   = "TestTagValue"
    Tag31   = "TestTagValue"
    Tag32   = "TestTagValue"
    Tag33   = "TestTagValue"
    Tag34   = "TestTagValue"
    Tag35   = "TestTagValue"
    Tag36   = "TestTagValue"
    Tag37   = "TestTagValue"
    Tag38   = "TestTagValue"
    Tag39   = "TestTagValue"
    Tag40   = "TestTagValue"
    Tag41   = "TestTagValue"
    Tag42   = "TestTagValue"
    Tag43   = "TestTagValue"
    Tag44   = "TestTagValue"
    Tag45   = "TestTagValue"
    Tag46   = "TestTagValue"
    Tag47   = "TestTagValue"
  }
}
`, rName)
}

const testAccFileSystemWithPerformanceModeConfig = `
resource "aws_efs_file_system" "test2" {
  creation_token   = "supercalifragilisticexpialidocious"
  performance_mode = "maxIO"
}
`

func testAccFileSystemWithKMSKeyConfig(rInt int, enable bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test %[1]d"
}

resource "aws_efs_file_system" "test" {
  encrypted  = %[2]t
  kms_key_id = aws_kms_key.test.arn
}
`, rInt, enable)
}

func testAccFileSystemConfig_ThroughputMode(throughputMode string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  throughput_mode = %q
}
`, throughputMode)
}

func testAccFileSystemConfig_ProvisionedThroughputInMibps(provisionedThroughputInMibps float64) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  provisioned_throughput_in_mibps = %f
  throughput_mode                 = "provisioned"
}
`, provisionedThroughputInMibps)
}

func testAccFileSystemWithLifecyclePolicyConfig(lpName, lpVal string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  lifecycle_policy {
    %s = %q
  }
}
`, lpName, lpVal)
}

func testAccFileSystemWithLifecyclePolicyMultiConfig(lpName1, lpVal1, lpName2, lpVal2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  lifecycle_policy {
    %[1]s = %[2]q
  }

  lifecycle_policy {
    %[3]s = %[4]q
  }
}
`, lpName1, lpVal1, lpName2, lpVal2)
}

const testAccFileSystemRemovedLifecyclePolicyConfig = `
resource "aws_efs_file_system" "test" {}
`
