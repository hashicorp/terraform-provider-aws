package datasync_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
)

func TestAccDataSyncLocationEFS_basic(t *testing.T) {
	var locationEfs1 datasync.DescribeLocationEfsOutput
	efsFileSystemResourceName := "aws_efs_file_system.test"
	resourceName := "aws_datasync_location_efs.test"
	subnetResourceName := "aws_subnet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ec2_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ec2_config.0.security_group_arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_config.0.subnet_arn", subnetResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "efs_file_system_arn", efsFileSystemResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^efs://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"efs_file_system_arn"},
			},
		},
	})
}

func TestAccDataSyncLocationEFS_accessPointARN(t *testing.T) {
	var locationEfs1 datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_accessPointARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					resource.TestCheckResourceAttrPair(resourceName, "access_point_arn", "aws_efs_access_point.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "in_transit_encryption", "TLS1_2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"efs_file_system_arn"},
			},
		},
	})
}

func TestAccDataSyncLocationEFS_disappears(t *testing.T) {
	var locationEfs1 datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationEFS(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceLocationEFS(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationEFS_subdirectory(t *testing.T) {
	var locationEfs1 datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_subdirectory(rName, "/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"efs_file_system_arn"},
			},
		},
	})
}

func TestAccDataSyncLocationEFS_tags(t *testing.T) {
	var locationEfs1, locationEfs2, locationEfs3 datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"efs_file_system_arn"},
			},
			{
				Config: testAccLocationEFSConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs2),
					testAccCheckLocationEFSNotRecreated(&locationEfs1, &locationEfs2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationEFSConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs3),
					testAccCheckLocationEFSNotRecreated(&locationEfs2, &locationEfs3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationEFSDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_efs" {
			continue
		}

		input := &datasync.DescribeLocationEfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationEfs(input)

		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckLocationEFSExists(resourceName string, locationEfs *datasync.DescribeLocationEfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		input := &datasync.DescribeLocationEfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationEfs(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationEfs = *output

		return nil
	}
}

func testAccCheckLocationEFSNotRecreated(i, j *datasync.DescribeLocationEfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location EFS was recreated")
		}

		return nil
	}
}

func testAccLocationEFSBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName)
}

func testAccLocationEFSConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLocationEFSBaseConfig(rName), `
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`)
}

func testAccLocationEFSConfig_subdirectory(rName, subdirectory string) string {
	return acctest.ConfigCompose(testAccLocationEFSBaseConfig(rName), fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn
  subdirectory        = %[1]q

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`, subdirectory))
}

func testAccLocationEFSConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationEFSBaseConfig(rName), fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccLocationEFSConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationEFSBaseConfig(rName), fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}

func testAccLocationEFSConfig_accessPointARN(rName string) string {
	return acctest.ConfigCompose(testAccLocationEFSBaseConfig(rName), fmt.Sprintf(`
resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn   = aws_efs_mount_target.test.file_system_arn
  access_point_arn      = aws_efs_access_point.test.arn
  in_transit_encryption = "TLS1_2"

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`, rName))
}
