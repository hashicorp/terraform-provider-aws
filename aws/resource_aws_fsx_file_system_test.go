package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSFsxFileSystem_lustreBasic(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemLustreBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "LUSTRE"),
					resource.TestCheckResourceAttr(resourceName, "capacity", "3600"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kms_key_id"},
			},
		},
	})
}

func TestAccAWSFsxFileSystem_lustreConfig(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemLustreConfigOpts,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.import_path", "s3://nasanex"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.chunk_size", "2048"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kms_key_id"},
			},
		},
	})
}

func TestAccAWSFsxFileSystem_lustreUpdate(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemLustreConfigOpts,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.weekly_maintenance_start_time", "3:03:30"),
				),
			},
			{
				Config: testAccAwsFsxFileSystemLustreUpdateOpts,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.0.weekly_maintenance_start_time", "5:05:50"),
				),
			},
		},
	})
}

func TestAccAWSFsxFileSystem_windowsConfig(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemWindowsConfigOpts,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.0.backup_retention", "3"),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.0.copy_tags_to_backups", "true"),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.0.throughput_capacity", "1024"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kms_key_id"},
			},
		},
	})
}

func TestAccAWSFsxFileSystem_windowsUpdate(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemWindowsConfigOpts,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.0.backup_retention", "3"),
				),
			},
			{
				Config: testAccAwsFsxFileSystemWindowsUpdateOpts,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.0.backup_retention", "30"),
				),
			},
		},
	})
}

func testAccCheckFileSystemExists(n string, v *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).fsxconn

		request := &fsx.DescribeFileSystemsInput{
			FileSystemIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeFileSystems(request)
		if err == nil {
			if response.FileSystems != nil && len(response.FileSystems) > 0 {
				*v = *response.FileSystems[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding FSx filesystem %s", rs.Primary.ID)
	}
}

const testAccAwsFsxFileSystemLustreBasic = `
resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     				= "${aws_vpc.test.id}"
	cidr_block 				= "10.0.1.0/24"
	availability_zone = "us-east-1a"
}

resource "aws_kms_key" "test" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_file_system" "test" {
	depends_on 	= ["aws_subnet.test", "aws_kms_key.test"]
  type 				= "LUSTRE"
  capacity 		= 3600
	kms_key_id 	= "${aws_kms_key.test.key_id}"
	subnet_ids 	= ["${aws_subnet.test.id}"]
}
`

const testAccAwsFsxFileSystemLustreConfigOpts = `
resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     				= "${aws_vpc.test.id}"
	cidr_block 				= "10.0.1.0/24"
	availability_zone = "us-east-1a"
}

resource "aws_kms_key" "test" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_file_system" "test" {
	depends_on 	= ["aws_subnet.test", "aws_kms_key.test"]
  type 				= "LUSTRE"
  capacity 		= 3600
	kms_key_id 	= "${aws_kms_key.test.key_id}"
	subnet_ids 	= ["${aws_subnet.test.id}"]

	lustre_configuration {
		import_path = "s3://nasanex"
		chunk_size 	= 2048
	}
}
`

const testAccAwsFsxFileSystemLustreUpdateOpts = `
resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     				= "${aws_vpc.test.id}"
	cidr_block 				= "10.0.1.0/24"
	availability_zone = "us-east-1a"
}

resource "aws_kms_key" "test" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_file_system" "test" {
	depends_on 	= ["aws_subnet.test", "aws_kms_key.test"]
  type 				= "LUSTRE"
  capacity 		= 3600
	kms_key_id 	= "${aws_kms_key.test.key_id}"
	subnet_ids 	= ["${aws_subnet.test.id}"]

	lustre_configuration {
		import_path = "s3://nasanex"
		chunk_size 	= 2048
		weekly_maintenance_start_time = "5:05:50"
	}
}
`

const testAccAwsFsxFileSystemWindowsConfigOpts = `
resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id     				= "${aws_vpc.test.id}"
	cidr_block 				= "10.0.1.0/24"
	availability_zone = "us-east-1a"
}

resource "aws_subnet" "test2" {
  vpc_id     				= "${aws_vpc.test.id}"
	cidr_block 				= "10.0.2.0/24"
	availability_zone = "us-east-1d"
}

resource "aws_kms_key" "test" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type = "MicrosoftAD"

  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.test1.id}", "${aws_subnet.test2.id}"]
  }
}

resource "aws_fsx_file_system" "test" {
	depends_on 	= ["aws_subnet.test1", "aws_kms_key.test", "aws_directory_service_directory.test"]
  type 				= "WINDOWS"
  capacity 		= 300
	kms_key_id 	= "${aws_kms_key.test.arn}"
	subnet_ids 	= ["${aws_subnet.test1.id}"]

	windows_configuration {
		active_directory_id		= "${aws_directory_service_directory.test.id}"
		backup_retention 			= 3
		copy_tags_to_backups 	= true
		throughput_capacity 	= 1024
	}
}
`

const testAccAwsFsxFileSystemWindowsUpdateOpts = `
resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id     				= "${aws_vpc.test.id}"
	cidr_block 				= "10.0.1.0/24"
	availability_zone = "us-east-1a"
}

resource "aws_subnet" "test2" {
  vpc_id     				= "${aws_vpc.test.id}"
	cidr_block 				= "10.0.2.0/24"
	availability_zone = "us-east-1d"
}

resource "aws_kms_key" "test" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type = "MicrosoftAD"

  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.test1.id}", "${aws_subnet.test2.id}"]
  }
}

resource "aws_fsx_file_system" "test" {
	depends_on 	= ["aws_subnet.test1", "aws_kms_key.test", "aws_directory_service_directory.test"]
  type 				= "WINDOWS"
  capacity 		= 300
	kms_key_id 	= "${aws_kms_key.test.arn}"
	subnet_ids 	= ["${aws_subnet.test1.id}"]

	windows_configuration {
		active_directory_id		= "${aws_directory_service_directory.test.id}"
		backup_retention 			= 30
		copy_tags_to_backups 	= true
		throughput_capacity 	= 1024
	}
}
`
