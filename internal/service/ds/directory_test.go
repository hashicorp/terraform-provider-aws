// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDSDirectory_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAlias),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "Small"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SimpleAD"),
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
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccDSDirectory_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_basic(rName, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					acctest.CheckSDKResourceDisappears(ctx, t, tfds.ResourceDirectory(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDSDirectory_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_tags1(rName, domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrPassword,
				},
			},
			{
				Config: testAccDirectoryConfig_tags2(rName, domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDirectoryConfig_tags1(rName, domainName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDSDirectory_microsoft(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_microsoft(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAlias),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", "Enterprise"),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "Large"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MicrosoftAD"),
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
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccDSDirectory_microsoftStandard(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_microsoftStandard(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAlias),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", "Standard"),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "Small"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MicrosoftAD"),
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
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccDSDirectory_connector(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_connector(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAlias),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "1"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "connect_settings.0.customer_dns_ips.#", 0),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.0.customer_username", "Administrator"),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainName),
					resource.TestCheckResourceAttr(resourceName, "radius_settings.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "Small"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "ADConnector"),
					resource.TestCheckResourceAttr(resourceName, "vpc_settings.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrPassword,
				},
			},
		},
	})
}

func TestAccDSDirectory_withAliasAndSSO(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	alias := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckDirectoryService(ctx, t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_alias(rName, domainName, alias),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, alias),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "0"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "Small"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SimpleAD"),
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
					names.AttrPassword,
				},
			},
			{
				Config: testAccDirectoryConfig_sso(rName, domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtTrue),
				),
			},
			{
				Config: testAccDirectoryConfig_ssoModified(rName, domainName, alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccDSDirectory_desiredNumberOfDomainControllers(t *testing.T) {
	ctx := acctest.Context(t)
	var ds awstypes.DirectoryDescription
	resourceName := "aws_directory_service_directory.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckDirectoryService(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfig_desiredNumberOfDomainControllers(rName, domainName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttrSet(resourceName, "access_url"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAlias),
					resource.TestCheckResourceAttr(resourceName, "connect_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "2"),
					acctest.CheckResourceAttrGreaterThanValue(resourceName, "dns_ip_addresses.#", 0),
					resource.TestCheckResourceAttr(resourceName, "edition", "Enterprise"),
					resource.TestCheckResourceAttr(resourceName, "enable_sso", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, domainName),
					resource.TestCheckResourceAttrSet(resourceName, "security_group_id"),
					resource.TestCheckResourceAttrSet(resourceName, "short_name"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSize, "Large"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MicrosoftAD"),
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
					names.AttrPassword,
				},
			},
			{
				Config: testAccDirectoryConfig_desiredNumberOfDomainControllers(rName, domainName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "4"),
				),
			},
			{
				Config: testAccDirectoryConfig_desiredNumberOfDomainControllers(rName, domainName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryExists(ctx, t, resourceName, &ds),
					resource.TestCheckResourceAttr(resourceName, "desired_number_of_domain_controllers", "3"),
				),
			},
		},
	})
}

func testAccCheckDirectoryDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_directory_service_directory" {
				continue
			}

			_, err := tfds.FindDirectoryByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckDirectoryExists(ctx context.Context, t *testing.T, n string, v *awstypes.DirectoryDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DSClient(ctx)

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
