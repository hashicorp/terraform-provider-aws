package appstream_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
)

func TestAccAppStreamDirectoryConfig_basic(t *testing.T) {
	var v1, v2 appstream.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserNameUpdated := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPasswordUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	orgUnitDN := orgUnitFromDomain("Test", domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDirectoryConfigDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig(rName, domain, rUserName, rPassword, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_name", rUserName),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_password", rPassword),
				),
			},
			{
				Config: testAccDirectoryConfigConfig(rName, domain, rUserNameUpdated, rPasswordUpdated, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &v2),
					testAccCheckDirectoryConfigNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_name", rUserNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_password", rPasswordUpdated),
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
	var v appstream.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	orgUnitDN := orgUnitFromDomain("Test", domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDirectoryConfigDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig(rName, domain, rUserName, rPassword, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfappstream.ResourceDirectoryConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamDirectoryConfig_OrganizationalUnitDistinguishedNames(t *testing.T) {
	var v1, v2, v3 appstream.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	orgUnitDN1 := orgUnitFromDomain("One", domain)
	orgUnitDN2 := orgUnitFromDomain("Two", domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDirectoryConfigDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, appstream.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig(rName, domain, rUserName, rPassword, orgUnitDN1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN1),
				),
			},
			{
				Config: testAccDirectoryConfig_OrganizationalUnitDistinguishedNamesConfig(rName, domain, rUserName, rPassword, orgUnitDN1, orgUnitDN2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN1),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.1", orgUnitDN2),
				),
			},
			{
				Config: testAccDirectoryConfigConfig(rName, domain, rUserName, rPassword, orgUnitDN2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN2),
				),
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
			return fmt.Errorf("AppStream Directory Config %q does not exist", rs.Primary.ID)
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
			return fmt.Errorf("AppStream Directory Config %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDirectoryConfigNotRecreated(i, j *appstream.DirectoryConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreatedTime).Equal(aws.TimeValue(j.CreatedTime)) {
			return fmt.Errorf("AppStream Directory Config recreated")
		}

		return nil
	}
}

func orgUnitFromDomain(orgUnit, domainName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("OU=%s", orgUnit))
	for _, dc := range strings.Split(domainName, ".") {
		sb.WriteString(fmt.Sprintf(" DC=%s", dc))
	}
	return sb.String()
}

func testAccDirectoryConfigConfig(rName, domain, userName, password, orgUnitDN string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_appstream_directory_config" "test" {
  directory_name                          = %[1]q
  organizational_unit_distinguished_names = [%[4]q]

  service_account_credentials {
    account_name     = %[2]q
    account_password = %[3]q
  }

  depends_on = [
    aws_directory_service_directory.test
  ]
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = %[3]q
  edition  = "Standard"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain, userName, password, orgUnitDN))
}

func testAccDirectoryConfig_OrganizationalUnitDistinguishedNamesConfig(rName, domain, userName, password, orgUnitDN1, orgUnitDN2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_appstream_directory_config" "test" {
  directory_name                          = %[1]q
  organizational_unit_distinguished_names = [%[4]q, %[5]q]

  service_account_credentials {
    account_name     = %[2]q
    account_password = %[3]q
  }

  depends_on = [
    aws_directory_service_directory.test
  ]
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = %[3]q
  edition  = "Standard"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain, userName, password, orgUnitDN1, orgUnitDN2))
}
