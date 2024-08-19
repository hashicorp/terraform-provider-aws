// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamDirectoryConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rUserNameUpdated := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPasswordUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	orgUnitDN := orgUnitFromDomain("Test", domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryConfigDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_name", rUserName),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_password", rPassword),
				),
			},
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserNameUpdated, rPasswordUpdated, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, resourceName, &v2),
					testAccCheckDirectoryConfigNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	var v awstypes.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	orgUnitDN := orgUnitFromDomain("Test", domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryConfigDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceDirectoryConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamDirectoryConfig_OrganizationalUnitDistinguishedNames(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 awstypes.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	orgUnitDN1 := orgUnitFromDomain("One", domain)
	orgUnitDN2 := orgUnitFromDomain("Two", domain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryConfigDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN1),
				),
			},
			{
				Config: testAccDirectoryConfigConfig_organizationalUnitDistinguishedNames(rName, domain, rUserName, rPassword, orgUnitDN1, orgUnitDN2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN1),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.1", orgUnitDN2),
				),
			},
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN2),
				),
			},
		},
	})
}

func testAccCheckDirectoryConfigExists(ctx context.Context, resourceName string, appStreamDirectoryConfig *awstypes.DirectoryConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)
		resp, err := conn.DescribeDirectoryConfigs(ctx, &appstream.DescribeDirectoryConfigsInput{DirectoryNames: []string{rs.Primary.ID}})

		if err != nil {
			return err
		}

		if resp == nil || len(resp.DirectoryConfigs) == 0 {
			return fmt.Errorf("AppStream Directory Config %q does not exist", rs.Primary.ID)
		}

		*appStreamDirectoryConfig = resp.DirectoryConfigs[0]

		return nil
	}
}

func testAccCheckDirectoryConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_directory_config" {
				continue
			}

			resp, err := conn.DescribeDirectoryConfigs(ctx, &appstream.DescribeDirectoryConfigsInput{DirectoryNames: []string{rs.Primary.ID}})

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
}

func testAccCheckDirectoryConfigNotRecreated(i, j *awstypes.DirectoryConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreatedTime).Equal(aws.ToTime(j.CreatedTime)) {
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

func testAccDirectoryConfigConfig_basic(rName, domain, userName, password, orgUnitDN string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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

func testAccDirectoryConfigConfig_organizationalUnitDistinguishedNames(rName, domain, userName, password, orgUnitDN1, orgUnitDN2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
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
