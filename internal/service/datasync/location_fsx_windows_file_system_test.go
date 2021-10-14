package datasync_test

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_fsx_windows_file_system", &resource.Sweeper{
		Name: "aws_datasync_location_fsx_windows_file_system",
		F:    testSweepDataSyncLocationFsxWindows,
	})
}

func testSweepDataSyncLocationFsxWindows(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location FSX Windows sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving DataSync Location FSX Windows: %w", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location FSX Windows File System to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "fsxw://") {
				log.Printf("[INFO] Skipping DataSync Location FSX Windows File System: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location FSX Windows File System: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location FSX Windows (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncLocationFsxWindows_basic(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"
	fsResourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheckAWSDataSync(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
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

func TestAccAWSDataSyncLocationFsxWindows_disappears(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheckAWSDataSync(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceLocationFSxWindowsFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationFsxWindows_subdirectory(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheckAWSDataSync(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigSubdirectory("/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
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

func TestAccAWSDataSyncLocationFsxWindows_tags(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(fsx.EndpointsID, t)
			testAccPreCheckAWSDataSync(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
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
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSDataSyncLocationFsxWindowsDestroy(s *terraform.State) error {
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

func testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName string, locationFsxWindows *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
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

func testAccAWSDataSyncLocationFsxWindowsConfig() string {
	return acctest.ConfigCompose(testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1(), `
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test1.arn]
}
`)
}

func testAccAWSDataSyncLocationFsxWindowsConfigSubdirectory(subdirectory string) string {
	return testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test1.arn]
  subdirectory        = %[1]q
}
`, subdirectory)
}

func testAccAWSDataSyncLocationFsxWindowsConfigTags1(key1, value1 string) string {
	return testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() + fmt.Sprintf(`
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

func testAccAWSDataSyncLocationFsxWindowsConfigTags2(key1, value1, key2, value2 string) string {
	return testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() + fmt.Sprintf(`
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

func testAccAwsFsxWindowsFileSystemConfigBase() string {
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

func testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + `
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
