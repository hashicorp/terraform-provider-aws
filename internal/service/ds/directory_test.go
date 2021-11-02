package ds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
)

func TestAccDirectoryServiceDirectory_basic(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccDirectoryServiceDirectory_tags(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryTagsConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.project", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			{
				Config: testAccDirectoryServiceDirectoryUpdateTagsConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.project", "test2"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryRemoveTagsConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "test"),
				),
			},
		},
	})
}

func TestAccDirectoryServiceDirectory_microsoft(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckDirectoryService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoft(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "edition", directoryservice.DirectoryEditionEnterprise),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccDirectoryServiceDirectory_microsoftStandard(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckDirectoryService(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoftStandard(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "edition", directoryservice.DirectoryEditionStandard),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccDirectoryServiceDirectory_connector(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_connector(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "connect_settings.0.connect_ips.#"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccDirectoryServiceDirectory_withAliasAndSSO(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_withAlias(domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					testAccCheckServiceDirectoryAlias(resourceName, alias),
					testAccCheckServiceDirectorySso(resourceName, false),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso(domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					testAccCheckServiceDirectoryAlias(resourceName, alias),
					testAccCheckServiceDirectorySso(resourceName, true),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso_modified(domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					testAccCheckServiceDirectoryAlias(resourceName, alias),
					testAccCheckServiceDirectorySso(resourceName, false),
				),
			},
		},
	})
}

func testAccCheckDirectoryServiceDirectoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_directory" {
			continue
		}

		input := directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		}
		out, err := conn.DescribeDirectories(&input)

		if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil && len(out.DirectoryDescriptions) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Directory to be gone, but was still found")
		}
	}

	return nil
}

func TestAccDirectoryServiceDirectory_disappears(t *testing.T) {
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists(resourceName, &ds),
					acctest.CheckResourceDisappears(acctest.Provider, tfds.ResourceDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckServiceDirectoryExists(name string, ds *directoryservice.DirectoryDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn
		out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(out.DirectoryDescriptions) < 1 {
			return fmt.Errorf("No DS directory found")
		}

		if *out.DirectoryDescriptions[0].DirectoryId != rs.Primary.ID {
			return fmt.Errorf("DS directory ID mismatch - existing: %q, state: %q",
				*out.DirectoryDescriptions[0].DirectoryId, rs.Primary.ID)
		}

		*ds = *out.DirectoryDescriptions[0]

		return nil
	}
}

func testAccCheckServiceDirectoryAlias(name, alias string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn
		out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if *out.DirectoryDescriptions[0].Alias != alias {
			return fmt.Errorf("DS directory Alias mismatch - actual: %q, expected: %q",
				*out.DirectoryDescriptions[0].Alias, alias)
		}

		return nil
	}
}

func testAccCheckServiceDirectorySso(name string, ssoEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn
		out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if *out.DirectoryDescriptions[0].SsoEnabled != ssoEnabled {
			return fmt.Errorf("DS directory SSO mismatch - actual: %t, expected: %t",
				*out.DirectoryDescriptions[0].SsoEnabled, ssoEnabled)
		}

		return nil
	}
}

func testAccDirectoryServiceDirectoryConfig(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain),
	)
}

func testAccDirectoryServiceDirectoryTagsConfig(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    foo     = "test"
    project = "test"
  }
}
`, domain),
	)
}

func testAccDirectoryServiceDirectoryUpdateTagsConfig(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    foo     = "test"
    project = "test2"
    fizz    = "buzz"
  }
}
`, domain),
	)
}

func testAccDirectoryServiceDirectoryRemoveTagsConfig(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    foo = "test"
  }
}
`, domain),
	)
}

func testAccDirectoryServiceDirectoryConfig_connector(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"
  type     = "ADConnector"

  connect_settings {
    customer_dns_ips  = aws_directory_service_directory.base.dns_ip_addresses
    customer_username = "Administrator"
    vpc_id            = aws_vpc.test.id
    subnet_ids        = aws_subnet.test[*].id
  }
}

resource "aws_directory_service_directory" "base" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain),
	)
}

func testAccDirectoryServiceDirectoryConfig_microsoft(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain),
	)
}

func testAccDirectoryServiceDirectoryConfig_microsoftStandard(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain),
	)
}

func testAccDirectoryServiceDirectoryConfig_withAlias(domain, alias string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"
  alias    = %[2]q

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain, alias),
	)
}

func testAccDirectoryServiceDirectoryConfig_withSso(domain, alias string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name       = %[1]q
  password   = "SuperSecretPassw0rd"
  size       = "Small"
  alias      = %[2]q
  enable_sso = true

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain, alias),
	)
}

func testAccDirectoryServiceDirectoryConfig_withSso_modified(domain, alias string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVpcWithSubnets(2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name       = %[1]q
  password   = "SuperSecretPassw0rd"
  size       = "Small"
  alias      = %[2]q
  enable_sso = false

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, domain, alias),
	)
}
