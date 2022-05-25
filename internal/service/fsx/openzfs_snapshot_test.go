package fsx_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffsx "github.com/hashicorp/terraform-provider-aws/internal/service/fsx"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccFSxOpenzfsSnapshot_basic(t *testing.T) {
	var snapshot fsx.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`snapshot/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "volume_id"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccFSxOpenzfsSnapshot_disappears(t *testing.T) {
	var snapshot fsx.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot),
					acctest.CheckResourceDisappears(acctest.Provider, tffsx.ResourceOpenzfsSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFSxOpenzfsSnapshot_tags(t *testing.T) {
	var snapshot fsx.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot),
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
				Config: testAccOpenZFSSnapshotConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccOpenZFSSnapshotConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsSnapshot_name(t *testing.T) {
	var snapshot1, snapshot2 fsx.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSSnapshotConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot2),
					testAccCheckOpenzfsSnapshotNotRecreated(&snapshot1, &snapshot2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func TestAccFSxOpenzfsSnapshot_childVolume(t *testing.T) {
	var snapshot fsx.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_childVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`snapshot/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccFSxOpenzfsSnapshot_volumeId(t *testing.T) {
	var snapshot1, snapshot2 fsx.Snapshot
	resourceName := "aws_fsx_openzfs_snapshot.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(fsx.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, fsx.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckOpenzfsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOpenZFSSnapshotConfig_volumeID1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOpenZFSSnapshotConfig_volumeID2(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOpenzfsSnapshotExists(resourceName, &snapshot2),
					testAccCheckOpenzfsSnapshotRecreated(&snapshot1, &snapshot2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
		},
	})
}

func testAccCheckOpenzfsSnapshotExists(resourceName string, fs *fsx.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

		output, err := tffsx.FindSnapshotByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("FSx OpenZFS Snapshot (%s) not found", rs.Primary.ID)
		}

		*fs = *output

		return nil
	}
}

func testAccCheckOpenzfsSnapshotDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FSxConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_openzfs_snapshot" {
			continue
		}

		_, err := tffsx.FindSnapshotByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("FSx OpenZFS snapshot %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckOpenzfsSnapshotNotRecreated(i, j *fsx.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.SnapshotId) != aws.StringValue(j.SnapshotId) {
			return fmt.Errorf("FSx OpenZFS Snapshot (%s) recreated", aws.StringValue(i.SnapshotId))
		}

		return nil
	}
}

func testAccCheckOpenzfsSnapshotRecreated(i, j *fsx.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.SnapshotId) == aws.StringValue(j.SnapshotId) {
			return fmt.Errorf("FSx OpenZFS Snapshot (%s) not recreated", aws.StringValue(i.SnapshotId))
		}

		return nil
	}
}

func testAccOpenzfsSnapshotBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_fsx_openzfs_file_system" "test" {
  storage_capacity    = 64
  subnet_ids          = [aws_subnet.test1.id]
  deployment_type     = "SINGLE_AZ_1"
  throughput_capacity = 64


  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_tags1(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccOpenZFSSnapshotConfig_tags2(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_file_system.test.root_volume_id


  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccOpenZFSSnapshotConfig_childVolume(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_volume.test.id
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_volumeID1(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test1" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_volume.test1.id
}
`, rName))
}

func testAccOpenZFSSnapshotConfig_volumeID2(rName string) string {
	return acctest.ConfigCompose(testAccOpenzfsSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_fsx_openzfs_volume" "test2" {
  name             = %[1]q
  parent_volume_id = aws_fsx_openzfs_file_system.test.root_volume_id
}

resource "aws_fsx_openzfs_snapshot" "test" {
  name      = %[1]q
  volume_id = aws_fsx_openzfs_volume.test2.id
}
`, rName))
}
