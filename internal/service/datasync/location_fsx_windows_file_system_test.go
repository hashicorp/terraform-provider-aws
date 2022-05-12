package datasync_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
)

func TestAccDataSyncLocationFSxWindowsFileSystem_basic(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"
	fsResourceName := "aws_fsx_windows_file_system.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(resourceName, &locationFsxWindows1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "fsx_filesystem_arn", fsResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^fsxw://.+/`)),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccWSDataSyncLocationFsxWindowsImportStateIdFunc(resourceName),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDataSyncLocationFSxWindowsFileSystem_disappears(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(resourceName, &locationFsxWindows1),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationFSxWindowsFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationFSxWindowsFileSystem_subdirectory(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsSubdirectoryConfig(domainName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccWSDataSyncLocationFsxWindowsImportStateIdFunc(resourceName),
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccDataSyncLocationFSxWindowsFileSystem_tags(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationFSxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationFSxWindowsTags1Config(domainName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccWSDataSyncLocationFsxWindowsImportStateIdFunc(resourceName),
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccLocationFSxWindowsTags2Config(domainName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationFSxWindowsTags1Config(domainName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationFSxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationFSxWindowsDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_fsx_windows_file_system" {
			continue
		}

		input := &datasync.DescribeLocationFsxWindowsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationFsxWindows(input)

		if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckLocationFSxWindowsExists(resourceName string, locationFsxWindows *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		input := &datasync.DescribeLocationFsxWindowsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationFsxWindows(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationFsxWindows = *output

		return nil
	}
}

func testAccWSDataSyncLocationFsxWindowsImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s#%s", rs.Primary.ID, rs.Primary.Attributes["fsx_filesystem_arn"]), nil
	}
}

func testAccLocationFSxWindowsConfig(domain string) string {
	return acctest.ConfigCompose(testAccFSxWindowsFileSystemSecurityGroupIds1Config(domain), `
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test1.arn]
}
`)
}

func testAccLocationFSxWindowsSubdirectoryConfig(domain, subdirectory string) string {
	return testAccFSxWindowsFileSystemSecurityGroupIds1Config(domain) + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test1.arn]
  subdirectory        = %[1]q
}
`, subdirectory)
}

func testAccLocationFSxWindowsTags1Config(domain, key1, value1 string) string {
	return testAccFSxWindowsFileSystemSecurityGroupIds1Config(domain) + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test1.arn]

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1)
}

func testAccLocationFSxWindowsTags2Config(domain, key1, value1, key2, value2 string) string {
	return testAccFSxWindowsFileSystemSecurityGroupIds1Config(domain) + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test1.arn]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2)
}

func testAccFSxWindowsFileSystemBaseConfig(domain string) string {
	return fmt.Sprintf(`
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
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
    vpc_id     = aws_vpc.test.id
  }
}
`, domain)
}

func testAccFSxWindowsFileSystemSecurityGroupIds1Config(domain string) string {
	return testAccFSxWindowsFileSystemBaseConfig(domain) + `
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test1.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test1.id]
  throughput_capacity = 8
}
`
}
