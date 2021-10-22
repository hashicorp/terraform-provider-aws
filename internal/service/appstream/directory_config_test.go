package appstream_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamDirectoryConfig_basic(t *testing.T) {
	var directoryOutput appstream.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", rName, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserNameUpdated := fmt.Sprintf("%s\\%s", rName, sdkacctest.RandString(10))
	rPasswordUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDirectoryConfigDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig(rName, rUserName, rPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &directoryOutput),
					resource.TestCheckResourceAttr(resourceName, "directory_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				Config: testAccDirectoryConfigConfig(rName, rUserNameUpdated, rPasswordUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &directoryOutput),
					resource.TestCheckResourceAttr(resourceName, "directory_name", rName),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_account_credentials.0.account_password"},
			},
		},
	})
}

func TestAccAppStreamDirectoryConfig_disappears(t *testing.T) {
	var directoryOutput appstream.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", rName, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckHasIAMRole(t, "AmazonAppStreamServiceAccess")
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDirectoryConfigDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig(rName, rUserName, rPassword),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &directoryOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceDirectoryConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDirectoryConfigExists(resourceName string, appStreamDirectoryConfig *appstream.DirectoryConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn
		resp, err := conn.DescribeDirectoryConfigs(&appstream.DescribeDirectoryConfigsInput{DirectoryNames: []*string{aws.String(rs.Primary.ID)}})

		if err != nil {
			return err
		}

		if resp == nil && len(resp.DirectoryConfigs) == 0 {
			return fmt.Errorf("appstream directory config %q does not exist", rs.Primary.ID)
		}

		*appStreamDirectoryConfig = *resp.DirectoryConfigs[0]

		return nil
	}
}

func testAccCheckDirectoryConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_appstream_directory_config" {
			continue
		}

		resp, err := conn.DescribeDirectoryConfigs(&appstream.DescribeDirectoryConfigsInput{DirectoryNames: []*string{aws.String(rs.Primary.ID)}})

		if tfawserr.ErrCodeEquals(err, appstream.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && len(resp.DirectoryConfigs) > 0 {
			return fmt.Errorf("appstream directory config %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccDirectoryConfigConfig(name, userName, password string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

data "aws_organizations_organizational_units" "test" {
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_appstream_directory_config" "test" {
  directory_name                          = %[1]q
  organizational_unit_distinguished_names = data.aws_organizations_organizational_units.test.children.*.id
  
  service_account_credentials{
    account_name     = %[2]q
    account_password = %[3]q
  }
}
`, name, userName, password)
}
