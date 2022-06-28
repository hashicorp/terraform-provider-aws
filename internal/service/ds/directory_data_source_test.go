package ds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDSDirectoryDataSource_nonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDirectoryDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func TestAccDSDirectoryDataSource_simpleAD(t *testing.T) {
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test-simple-ad"
	dataSourceName := "data.aws_directory_service_directory.test-simple-ad"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckDirectoryServiceSimpleDirectory(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_simpleAD(rName, alias, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", directoryservice.DirectoryTypeSimpleAd),
					resource.TestCheckResourceAttr(dataSourceName, "size", "Small"),
					resource.TestCheckResourceAttr(dataSourceName, "name", domainName),
					resource.TestCheckResourceAttr(dataSourceName, "description", "tf-testacc SimpleAD"),
					resource.TestCheckResourceAttr(dataSourceName, "short_name", "corp"),
					resource.TestCheckResourceAttr(dataSourceName, "alias", alias),
					resource.TestCheckResourceAttr(dataSourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.vpc_id", resourceName, "vpc_settings.0.vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.subnet_ids", resourceName, "vpc_settings.0.subnet_ids"),
					resource.TestCheckResourceAttr(dataSourceName, "access_url", fmt.Sprintf("%s.awsapps.com", alias)),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_ip_addresses", resourceName, "dns_ip_addresses"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
				),
			},
		},
	})
}

func TestAccDSDirectoryDataSource_microsoftAD(t *testing.T) {
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test-microsoft-ad"
	dataSourceName := "data.aws_directory_service_directory.test-microsoft-ad"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_microsoftAD(rName, alias, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", directoryservice.DirectoryTypeMicrosoftAd),
					resource.TestCheckResourceAttr(dataSourceName, "edition", "Standard"),
					resource.TestCheckResourceAttr(dataSourceName, "name", domainName),
					resource.TestCheckResourceAttr(dataSourceName, "description", "tf-testacc MicrosoftAD"),
					resource.TestCheckResourceAttr(dataSourceName, "short_name", "corp"),
					resource.TestCheckResourceAttr(dataSourceName, "alias", alias),
					resource.TestCheckResourceAttr(dataSourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.vpc_id", resourceName, "vpc_settings.0.vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.subnet_ids", resourceName, "vpc_settings.0.subnet_ids"),
					resource.TestCheckResourceAttr(dataSourceName, "access_url", fmt.Sprintf("%s.awsapps.com", alias)),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_ip_addresses", resourceName, "dns_ip_addresses"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
				),
			},
		},
	})
}

func TestAccDSDirectoryDataSource_connector(t *testing.T) {
	resourceName := "aws_directory_service_directory.test"
	dataSourceName := "data.aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_connector(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "connect_settings.0.connect_ips", resourceName, "connect_settings.0.connect_ips"),
				),
			},
		},
	})
}

const testAccDirectoryDataSourceConfig_nonExistent = `
data "aws_directory_service_directory" "test" {
  directory_id = "d-abc0123456"
}
`

func testAccDirectoryDataSourceConfig_simpleAD(rName, alias, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
data "aws_directory_service_directory" "test-simple-ad" {
  directory_id = aws_directory_service_directory.test-simple-ad.id
}

resource "aws_directory_service_directory" "test-simple-ad" {
  type        = "SimpleAD"
  size        = "Small"
  name        = %[2]q
  description = "tf-testacc SimpleAD"
  short_name  = "corp"
  password    = "#S1ncerely"

  alias      = %[1]q
  enable_sso = false

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, alias, domain))
}

func testAccDirectoryDataSourceConfig_microsoftAD(rName, alias, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
data "aws_directory_service_directory" "test-microsoft-ad" {
  directory_id = aws_directory_service_directory.test-microsoft-ad.id
}

resource "aws_directory_service_directory" "test-microsoft-ad" {
  type        = "MicrosoftAD"
  edition     = "Standard"
  name        = %[2]q
  description = "tf-testacc MicrosoftAD"
  short_name  = "corp"
  password    = "#S1ncerely"

  alias      = %[1]q
  enable_sso = false

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }
}
`, alias, domain))
}

func testAccDirectoryDataSourceConfig_connector(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
data "aws_directory_service_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
}

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
`, domain))
}
