package fsx_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFSxWindowsFileSystemDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_fsx_windows_file_system.test"
	datasourceName := "data.aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWindowsFileSystemDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWindowsFileSystemDataSourceConfig_subnetIDs1(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "active_directory_id", resourceName, "active_directory_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "automatic_backup_retention_days", resourceName, "automatic_backup_retention_days"),
					resource.TestCheckResourceAttrPair(datasourceName, "daily_automatic_backup_start_time", resourceName, "daily_automatic_backup_start_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "deployment_type", resourceName, "deployment_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_capacity", resourceName, "storage_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_type", resourceName, "storage_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids", resourceName, "subnet_ids"),
					resource.TestCheckResourceAttrPair(datasourceName, "throughput_capacity", resourceName, "throughput_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccWindowsFileSystemDataSourceBaseConfig() string {
	return `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
}

resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
    vpc_id     = aws_vpc.test.id
  }
}
`
}

func testAccWindowsFileSystemDataSourceConfig_subnetIDs1() string {
	return testAccWindowsFileSystemDataSourceBaseConfig() + `
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}

data "aws_fsx_windows_file_system" "test" {
  id = aws_fsx_windows_file_system.test.id
}
`
}
