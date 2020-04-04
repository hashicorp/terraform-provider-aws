package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_fsx_windows", &resource.Sweeper{
		Name: "aws_datasync_location_fsx_windows",
		F:    testSweepDataSyncLocationFsxWindows,
	})
}

func testSweepDataSyncLocationFsxWindows(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).datasyncconn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location FSX Windows sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location FSX Windows: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location FSX Windowss to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "FsxWindows://") {
				log.Printf("[INFO] Skipping DataSync Location FSX Windows: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location FSX Windows: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if isAWSErr(err, "InvalidRequestException", "not found") {
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^FsxWindows://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationFsxWindows_disappears(t *testing.T) {
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					testAccCheckAWSDataSyncLocationFsxWindowsDisappears(&locationFsxWindows1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationFsxWindows_Subdirectory(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	var locationFsxWindows1 datasync.DescribeLocationFsxWindowsOutput
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigSubdirectory(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationFsxWindows_Tags(t *testing.T) {
	var locationFsxWindows1, locationFsxWindows2, locationFsxWindows3 datasync.DescribeLocationFsxWindowsOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_location_fsx_windows.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncLocationFsxWindowsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags1(rName, "key1", "value1"),
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
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows2),
					testAccCheckAWSDataSyncLocationFsxWindowsNotRecreated(&locationFsxWindows1, &locationFsxWindows2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncLocationFsxWindowsConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncLocationFsxWindowsExists(resourceName, &locationFsxWindows3),
					testAccCheckAWSDataSyncLocationFsxWindowsNotRecreated(&locationFsxWindows2, &locationFsxWindows3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAWSDataSyncLocationFsxWindowsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_fsx_windows" {
			continue
		}

		input := &datasync.DescribeLocationFsxWindowsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationFsxWindows(input)

		if isAWSErr(err, datasync.ErrCodeInvalidRequestException, "not found") {
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

		conn := testAccProvider.Meta().(*AWSClient).datasyncconn
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

func testAccCheckAWSDataSyncLocationFsxWindowsDisappears(location *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckAWSDataSyncLocationFsxWindowsNotRecreated(i, j *datasync.DescribeLocationFsxWindowsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) != aws.TimeValue(j.CreationTime) {
			return errors.New("DataSync Location FSX Windows was recreated")
		}

		return nil
	}
}

func testAccAWSDataSyncLocationFsxWindowsConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "aws-thinstaller" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-location-FsxWindows"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-FsxWindows"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-FsxWindows"
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }

  tags = {
    Name = "tf-acc-test-datasync-location-FsxWindows"
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = "${aws_subnet.test.id}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "tf-acc-test-datasync-location-FsxWindows"
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-thinstaller.id}"
  associate_public_ip_address = true

  # Default instance type from sync.sh
  instance_type          = "c5.2xlarge"
  vpc_security_group_ids = ["${aws_security_group.test.id}"]
  subnet_id              = "${aws_subnet.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-location-FsxWindows"
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = "${aws_instance.test.public_ip}"
  name       = %q
}
`, rName)
}

func testAccAWSDataSyncLocationFsxWindowsConfig(rName string) string {
	return testAccAWSDataSyncLocationFsxWindowsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
}
`)
}

func testAccAWSDataSyncLocationFsxWindowsConfigSubdirectory(rName, subdirectory string) string {
	return testAccAWSDataSyncLocationFsxWindowsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
  server_hostname = "example.com"
  subdirectory    = %q
}
`, subdirectory)
}

func testAccAWSDataSyncLocationFsxWindowsConfigTags1(rName, key1, value1 string) string {
	return testAccAWSDataSyncLocationFsxWindowsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1)
}

func testAccAWSDataSyncLocationFsxWindowsConfigTags2(rName, key1, value1, key2, value2 string) string {
	return testAccAWSDataSyncLocationFsxWindowsConfigBase(rName) + fmt.Sprintf(`
resource "aws_datasync_location_fsx_windows" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2)
}
