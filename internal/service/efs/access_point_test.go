package efs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
)

func TestAccEFSAccessPoint_basic(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"
	fsResourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_arn", fsResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_id", fsResourceName, "id"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticfilesystem", regexp.MustCompile(`access-point/fsap-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/"),
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

func TestAccEFSAccessPoint_Root_directory(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_rootDirectory(rName, "/home/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/home/test"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.#", "0"),
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

func TestAccEFSAccessPoint_RootDirectoryCreation_info(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_rootDirectoryCreationInfo(rName, "/home/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/home/test"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.owner_gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.owner_uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.permissions", "755"),
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

func TestAccEFSAccessPoint_POSIX_user(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_posixUser(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.secondary_gids.%", "0"),
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

func TestAccEFSAccessPoint_POSIXUserSecondary_gids(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_posixUserSecondaryGids(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.secondary_gids.#", "1")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEFSAccessPoint_tags(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
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
				Config: testAccAccessPointConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAccessPointConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEFSAccessPoint_disappears(t *testing.T) {
	var ap efs.AccessPointDescription
	resourceName := "aws_efs_access_point.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPointExists(resourceName, &ap),
					acctest.CheckResourceDisappears(acctest.Provider, tfefs.ResourceAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessPointDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_access_point" {
			continue
		}

		resp, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
			AccessPointId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if tfawserr.ErrCodeEquals(err, efs.ErrCodeAccessPointNotFound) {
				continue
			}
			return fmt.Errorf("Error describing EFS access point in tests: %s", err)
		}
		if len(resp.AccessPoints) > 0 {
			return fmt.Errorf("EFS access point %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAccessPointExists(resourceID string, mount *efs.AccessPointDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		fs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn
		mt, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
			AccessPointId: aws.String(fs.Primary.ID),
		})
		if err != nil {
			return err
		}

		apId := aws.StringValue(mt.AccessPoints[0].AccessPointId)
		if apId != fs.Primary.ID {
			return fmt.Errorf("access point ID mismatch: %q != %q", apId, fs.Primary.ID)
		}

		*mount = *mt.AccessPoints[0]

		return nil
	}
}

func testAccAccessPointConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = "%s"
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}
`, rName)
}

func testAccAccessPointConfig_rootDirectory(rName, dir string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  root_directory {
    path = %[2]q
  }
}
`, rName, dir)
}

func testAccAccessPointConfig_rootDirectoryCreationInfo(rName, dir string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  root_directory {
    path = %[2]q
    creation_info {
      owner_gid   = 1001
      owner_uid   = 1001
      permissions = "755"
    }
  }
}
`, rName, dir)
}

func testAccAccessPointConfig_posixUser(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = "%s"
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  posix_user {
    gid = 1001
    uid = 1001
  }
}
`, rName)
}

func testAccAccessPointConfig_posixUserSecondaryGids(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = "%s"
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  posix_user {
    gid            = 1001
    uid            = 1001
    secondary_gids = [1002]
  }
}
`, rName)
}

func testAccAccessPointConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAccessPointConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
