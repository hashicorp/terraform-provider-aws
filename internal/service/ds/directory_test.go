// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDSDirectory_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, "alias"),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, "size", "Small"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "SimpleAD"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.availability_zones.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.subnet_ids.#", "2"),
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

func TestAccDSDirectory_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfds.ResourceDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDSDirectory_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_tags1(rName, domainName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
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
				Config: testAccDirectoryConfig_tags2(rName, domainName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDirectoryConfig_tags1(rName, domainName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDSDirectory_microsoft(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_microsoft(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, "alias"),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", "Enterprise"),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, "size", "Large"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MicrosoftAD"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.availability_zones.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.subnet_ids.#", "2"),
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

func TestAccDSDirectory_microsoftStandard(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_microsoftStandard(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, "alias"),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, "size", "Small"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MicrosoftAD"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.availability_zones.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.subnet_ids.#", "2"),
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

func TestAccDSDirectory_connector(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_connector(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, "alias"),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "1"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "connect_settings.0.customer_dns_ips.#", 0),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.0.customer_username", "Administrator"),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttr(resourceName, "radius_settings.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, "size", "Small"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "ADConnector"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.#", "0"),
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

func TestAccDSDirectory_withAliasAndSSO(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_alias(rName, domainName, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttr(resourceName, "alias", alias),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, "size", "Small"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "SimpleAD"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.availability_zones.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.subnet_ids.#", "2"),
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
				Config: testAccDirectoryConfig_sso(rName, domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "true"),
				),
			},
			{
				Config: testAccDirectoryConfig_ssoModified(rName, domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "false"),
				),
			},
		},
	})
}

func TestAccDSDirectory_desiredNumberOfDomainControllers(t *testing.T) {
	ctx := acctest.Context(t)
	var ds directoryservice.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_desiredNumberOfDomainControllers(rName, domainName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, "alias"),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", "Enterprise"),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, "size", "Large"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MicrosoftAD"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.availability_zones.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.0.subnet_ids.#", "2"),
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
				Config: testAccDirectoryConfig_desiredNumberOfDomainControllers(rName, domainName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "4"),
				),
			},
			{
				Config: testAccDirectoryConfig_desiredNumberOfDomainControllers(rName, domainName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "3"),
				),
			},
		},
	})
}

func testAccCheckDirectoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_directory" {
				continue
			}

			_, err := tfds.FindDirectoryByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Directory Service Directory %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDirectoryExists(ctx context.Context, n string, v *directoryservice.DirectoryDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Directory Service Directory ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn(ctx)

		output, err := tfds.FindDirectoryByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDirectoryConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfig_tags1(rName, domain, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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
    %[2]q = %[3]q
  }
}
`, domain, tagKey1, tagValue1))
}

func testAccDirectoryConfig_tags2(rName, domain, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, domain, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDirectoryConfig_connector(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfig_microsoft(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfig_microsoftStandard(rName, domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfig_alias(rName, domain, alias string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfig_sso(rName, domain, alias string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfig_ssoModified(rName, domain, alias string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfig_desiredNumberOfDomainControllers(rName, domain string, desiredNumber int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  desired_number_of_domain_controllers = %[2]d
}
`, domain, desiredNumber),
	)
}
