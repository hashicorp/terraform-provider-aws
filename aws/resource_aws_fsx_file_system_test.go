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
		CheckDestroy:  testAccCheckFsxFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemLustreBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "LUSTRE"),
					resource.TestCheckResourceAttr(resourceName, "capacity", "3600"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeout", "security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxFileSystem_lustreConfig(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemLustreConfigOpts(),
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
				ImportStateVerifyIgnore: []string{"timeout", "security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxFileSystem_lustreUpdate(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemLustreConfigOpts(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "lustre_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "lustre_configuration.0.weekly_maintenance_start_time"),
				),
			},
			{
				Config: testAccAwsFsxFileSystemLustreUpdateOpts(),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemWindowsConfigOpts(),
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
				ImportStateVerifyIgnore: []string{"timeout", "security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxFileSystem_windowsUpdate(t *testing.T) {
	var v fsx.FileSystem
	resourceName := "aws_fsx_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxFileSystemWindowsConfigOpts(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFileSystemExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "windows_configuration.0.backup_retention", "3"),
				),
			},
			{
				Config: testAccAwsFsxFileSystemWindowsUpdateOpts(),
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

func testAccCheckFsxFileSystemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fsxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_file_system" {
			continue
		}

		filesystem, err := describeFsxFileSystem(conn, rs.Primary.ID)

		if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
			continue
		}

		if err != nil {
			return err
		}

		if filesystem != nil {
			return fmt.Errorf("FSx File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

const testAccAwsFsxFileSystemBaseConfig = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "test2" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    from_port   = 988
    to_port     = 988
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 135
    to_port     = 135
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 445
    to_port     = 445
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  ingress {
    from_port   = 55555
    to_port     = 55555
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks     = ["0.0.0.0/0"]
  }
}
`

const testAccAwsFsxFileSystemBaseWindowsConfig = `
resource "aws_kms_key" "test" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  
  vpc_settings {
    vpc_id     = "${aws_vpc.test.id}"
	subnet_ids = ["${aws_subnet.test1.id}", "${aws_subnet.test2.id}"]
  }
}
`

func testAccAwsFsxFileSystemLustreBasic() string {
	return testAccAwsFsxFileSystemBaseConfig + fmt.Sprintf(`
resource "aws_fsx_file_system" "test" {
  type       = "LUSTRE"
  capacity   = 3600
  subnet_ids = ["${aws_subnet.test1.id}"]
  security_group_ids = ["${aws_security_group.test1.id}"]
}
`)
}

func testAccAwsFsxFileSystemLustreConfigOpts() string {
	return testAccAwsFsxFileSystemBaseConfig + fmt.Sprintf(`
resource "aws_fsx_file_system" "test" {
  type       = "LUSTRE"
  capacity   = 3600
  subnet_ids = ["${aws_subnet.test1.id}"]

  lustre_configuration {
    import_path = "s3://nasanex"
    chunk_size 	= 2048
  }
}
`)
}

func testAccAwsFsxFileSystemLustreUpdateOpts() string {
	return testAccAwsFsxFileSystemBaseConfig + fmt.Sprintf(`
resource "aws_fsx_file_system" "test" {
  type       = "LUSTRE"
  capacity   = 3600
  subnet_ids = ["${aws_subnet.test1.id}"]

  lustre_configuration {
    import_path = "s3://nasanex"
    chunk_size 	= 2048
    weekly_maintenance_start_time = "5:05:50"
  }
}
`)
}

func testAccAwsFsxFileSystemWindowsConfigOpts() string {
	return testAccAwsFsxFileSystemBaseConfig + testAccAwsFsxFileSystemBaseWindowsConfig + fmt.Sprintf(`
resource "aws_fsx_file_system" "test" {
  type          = "WINDOWS"
  capacity 		= 300
  kms_key_id 	= "${aws_kms_key.test.arn}"
  subnet_ids 	= ["${aws_subnet.test1.id}"]

  windows_configuration {
    active_directory_id  = "${aws_directory_service_directory.test.id}"
    backup_retention     = 3
    copy_tags_to_backups = true
    throughput_capacity  = 1024
  }
}
`)
}

func testAccAwsFsxFileSystemWindowsUpdateOpts() string {
	return testAccAwsFsxFileSystemBaseConfig + testAccAwsFsxFileSystemBaseWindowsConfig + fmt.Sprintf(`
resource "aws_fsx_file_system" "test" {
  type         = "WINDOWS"
  capacity     = 300
  kms_key_id = "${aws_kms_key.test.arn}"
  subnet_ids = ["${aws_subnet.test1.id}"]

  windows_configuration {
    active_directory_id  = "${aws_directory_service_directory.test.id}"
	backup_retention     = 30
	copy_tags_to_backups = true
    throughput_capacity  = 1024
  }
}
`)
}
