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
)

func TestAccEFSFileSystemDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_efs_file_system.test"
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemDataSourceConfig_id,
				Check: resource.ComposeTestCheckFunc(
					testAccFileSystemCheckDataSource(dataSourceName, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "performance_mode", resourceName, "performance_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_token", resourceName, "creation_token"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypted", resourceName, "encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "provisioned_throughput_in_mibps", resourceName, "provisioned_throughput_in_mibps"),
					resource.TestCheckResourceAttrPair(dataSourceName, "throughput_mode", resourceName, "throughput_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lifecycle_policy", resourceName, "lifecycle_policy"),
					resource.TestMatchResourceAttr(dataSourceName, "size_in_bytes", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccEFSFileSystemDataSource_tags(t *testing.T) {
	dataSourceName := "data.aws_efs_file_system.test"
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemDataSourceConfig_tags,
				Check: resource.ComposeTestCheckFunc(
					testAccFileSystemCheckDataSource(dataSourceName, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "performance_mode", resourceName, "performance_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_token", resourceName, "creation_token"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypted", resourceName, "encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "provisioned_throughput_in_mibps", resourceName, "provisioned_throughput_in_mibps"),
					resource.TestCheckResourceAttrPair(dataSourceName, "throughput_mode", resourceName, "throughput_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lifecycle_policy", resourceName, "lifecycle_policy"),
					resource.TestMatchResourceAttr(dataSourceName, "size_in_bytes", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccEFSFileSystemDataSource_name(t *testing.T) {
	dataSourceName := "data.aws_efs_file_system.test"
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemDataSourceConfig_name,
				Check: resource.ComposeTestCheckFunc(
					testAccFileSystemCheckDataSource(dataSourceName, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "performance_mode", resourceName, "performance_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "creation_token", resourceName, "creation_token"),
					resource.TestCheckResourceAttrPair(dataSourceName, "encrypted", resourceName, "encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "provisioned_throughput_in_mibps", resourceName, "provisioned_throughput_in_mibps"),
					resource.TestCheckResourceAttrPair(dataSourceName, "throughput_mode", resourceName, "throughput_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "lifecycle_policy", resourceName, "lifecycle_policy"),
					resource.TestMatchResourceAttr(dataSourceName, "size_in_bytes", regexp.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccEFSFileSystemDataSource_availabilityZone(t *testing.T) {
	dataSourceName := "data.aws_efs_file_system.test"
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemDataSourceConfig_availabilityZone,
				Check: resource.ComposeTestCheckFunc(
					testAccFileSystemCheckDataSource(dataSourceName, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
				),
			},
		},
	})
}

func TestAccEFSFileSystemDataSource_nonExistent_fileSystemID(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccFileSystemDataSourceConfig_idNonExistent,
				ExpectError: regexp.MustCompile(`error reading EFS FileSystem`),
			},
		},
	})
}

func TestAccEFSFileSystemDataSource_nonExistent_tags(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemConfig_dataSourceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystem(resourceName, &desc),
				),
			},
			{
				Config:      testAccFileSystemDataSourceConfig_tagsNonExistent(rName),
				ExpectError: regexp.MustCompile(`Search returned 0 results`),
			},
		},
	})
}

func testAccFileSystemCheckDataSource(dName, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", dName)
		}

		efsRs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("can't find aws_efs_file_system.test in state")
		}

		attr := rs.Primary.Attributes

		if attr["creation_token"] != efsRs.Primary.Attributes["creation_token"] {
			return fmt.Errorf(
				"creation_token is %s; want %s",
				attr["creation_token"],
				efsRs.Primary.Attributes["creation_token"],
			)
		}

		if attr["id"] != efsRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"file_system_id is %s; want %s",
				attr["id"],
				efsRs.Primary.Attributes["id"],
			)
		}

		return nil
	}
}

const testAccFileSystemDataSourceConfig_idNonExistent = `
data "aws_efs_file_system" "test" {
  file_system_id = "fs-nonexistent"
}
`

func testAccFileSystemDataSourceConfig_tagsNonExistent(rName string) string {
	return acctest.ConfigCompose(
		testAccFileSystemConfig_dataSourceBasic(rName),
		`
data "aws_efs_file_system" "test" {
  tags = {
    Name = "Does_Not_Exist"
  }
}
`)
}

const testAccFileSystemDataSourceConfig_name = `
resource "aws_efs_file_system" "test" {}

data "aws_efs_file_system" "test" {
  creation_token = aws_efs_file_system.test.creation_token
}
`

const testAccFileSystemDataSourceConfig_id = `
resource "aws_efs_file_system" "test" {}

data "aws_efs_file_system" "test" {
  file_system_id = aws_efs_file_system.test.id
}
`

const testAccFileSystemDataSourceConfig_tags = `
resource "aws_efs_file_system" "test" {
  tags = {
    Name        = "default-efs"
    Environment = "dev"
  }
}

resource "aws_efs_file_system" "wrong-env" {
  tags = {
    Environment = "test"
  }
}

resource "aws_efs_file_system" "no-tags" {}

data "aws_efs_file_system" "test" {
  tags = aws_efs_file_system.test.tags
}
`

const testAccFileSystemDataSourceConfig_availabilityZone = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  availability_zone_name = data.aws_availability_zones.available.names[0]
}

data "aws_efs_file_system" "test" {
  file_system_id = aws_efs_file_system.test.id
}
`
