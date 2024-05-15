// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSDirectoryDataSource_simpleAD(t *testing.T) {
	ctx := acctest.Context(t)
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test"
	dataSourceName := "data.aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_simpleAD(rName, alias, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "access_url", dataSourceName, "access_url"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAlias, dataSourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.#", dataSourceName, "connect_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_ip_addresses.#", dataSourceName, "dns_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(resourceName, "edition", dataSourceName, "edition"),
					resource.TestCheckResourceAttrPair(resourceName, "enable_sso", dataSourceName, "enable_sso"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "radius_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", dataSourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "short_name", dataSourceName, "short_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSize, dataSourceName, names.AttrSize),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.#", dataSourceName, "vpc_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.0.availability_zones.#", dataSourceName, "vpc_settings.0.availability_zones.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.0.subnet_ids.#", dataSourceName, "vpc_settings.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.0.vpc_id", dataSourceName, "vpc_settings.0.vpc_id"),
				),
			},
		},
	})
}

func TestAccDSDirectoryDataSource_microsoftAD(t *testing.T) {
	ctx := acctest.Context(t)
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test"
	dataSourceName := "data.aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_microsoftAD(rName, alias, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "access_url", dataSourceName, "access_url"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAlias, dataSourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.#", dataSourceName, "connect_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_ip_addresses.#", dataSourceName, "dns_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(resourceName, "edition", dataSourceName, "edition"),
					resource.TestCheckResourceAttrPair(resourceName, "enable_sso", dataSourceName, "enable_sso"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "radius_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", dataSourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "short_name", dataSourceName, "short_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSize, dataSourceName, names.AttrSize),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.#", dataSourceName, "vpc_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.0.availability_zones.#", dataSourceName, "vpc_settings.0.availability_zones.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.0.subnet_ids.#", dataSourceName, "vpc_settings.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.0.vpc_id", dataSourceName, "vpc_settings.0.vpc_id"),
				),
			},
		},
	})
}

func TestAccDSDirectoryDataSource_connector(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_directory.test"
	dataSourceName := "data.aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_connector(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "access_url", dataSourceName, "access_url"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAlias, dataSourceName, names.AttrAlias),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.#", dataSourceName, "connect_settings.#"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.0.availability_zones.#", dataSourceName, "connect_settings.0.availability_zones.#"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.0.connect_ips.#", dataSourceName, "connect_settings.0.connect_ips.#"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.0.customer_dns_ips.#", dataSourceName, "connect_settings.0.customer_dns_ips.#"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.0.customer_username", dataSourceName, "connect_settings.0.customer_username"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.0.subnet_ids.#", dataSourceName, "connect_settings.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "connect_settings.0.vpc_id", dataSourceName, "connect_settings.0.vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_ip_addresses.#", dataSourceName, "dns_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(resourceName, "edition", dataSourceName, "edition"),
					resource.TestCheckResourceAttrPair(resourceName, "enable_sso", dataSourceName, "enable_sso"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "radius_settings.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "security_group_id", dataSourceName, "security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "short_name", dataSourceName, "short_name"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrSize, dataSourceName, names.AttrSize),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_settings.#", dataSourceName, "vpc_settings.#"),
				),
			},
		},
	})
}

func TestAccDSDirectoryDataSource_sharedMicrosoftAD(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_directory_service_directory.test"
	dataSourceName := "data.aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_sharedMicrosoftAD(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "dns_ip_addresses.#", dataSourceName, "dns_ip_addresses.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrType, "SharedMicrosoftAD"),
				),
			},
		},
	})
}

func testAccDirectoryDataSourceConfig_simpleAD(rName, alias, domain string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_directory_service_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
}

resource "aws_directory_service_directory" "test" {
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
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
data "aws_directory_service_directory" "test" {
  directory_id = aws_directory_service_directory.test.id
}

resource "aws_directory_service_directory" "test" {
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
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
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

func testAccDirectoryDataSourceConfig_sharedMicrosoftAD(rName, domain string) string {
	return acctest.ConfigCompose(testAccSharedDirectoryConfig_basic(rName, domain), `
resource "aws_directory_service_shared_directory_accepter" "test" {
  provider = "awsalternate"

  shared_directory_id = aws_directory_service_shared_directory.test.shared_directory_id
}

data "aws_directory_service_directory" "test" {
  provider = "awsalternate"

  directory_id = aws_directory_service_shared_directory_accepter.test.shared_directory_id
}
`)
}
