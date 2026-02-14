// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamDirectoryConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rUserNameUpdated := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPasswordUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	orgUnitDN := orgUnitFromDomain("Test", domain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryConfigDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_name", rUserName),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_password", rPassword),
				),
			},
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserNameUpdated, rPasswordUpdated, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v2),
					testAccCheckDirectoryConfigNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
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
	ctx := acctest.Context(t)
	var v awstypes.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	orgUnitDN := orgUnitFromDomain("Test", domain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryConfigDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappstream.ResourceDirectoryConfig(), resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	orgUnitDN1 := orgUnitFromDomain("One", domain)
	orgUnitDN2 := orgUnitFromDomain("Two", domain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryConfigDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN1),
				),
			},
			{
				Config: testAccDirectoryConfigConfig_organizationalUnitDistinguishedNames(rName, domain, rUserName, rPassword, orgUnitDN1, orgUnitDN2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN1),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.1", orgUnitDN2),
				),
			},
			{
				Config: testAccDirectoryConfigConfig_basic(rName, domain, rUserName, rPassword, orgUnitDN2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN2),
				),
			},
		},
	})
}

func TestAccAppStreamDirectoryConfig_CertificateBasedAuthParameters(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.DirectoryConfig
	resourceName := "aws_appstream_directory_config.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	rUserName := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPassword := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rUserNameUpdated := fmt.Sprintf("%s\\%s", domain, sdkacctest.RandString(10))
	rPasswordUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	orgUnitDN := orgUnitFromDomain("Test", domain)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDirectoryConfigDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryConfigConfig_certifcateBasedAuthParameters(rName, domain, rUserName, rPassword, orgUnitDN, string(awstypes.CertificateBasedAuthStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_name", rUserName),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_password", rPassword),
					resource.TestCheckResourceAttr(resourceName, "certificate_based_auth_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_based_auth_properties.0.status", string(awstypes.CertificateBasedAuthStatusEnabled)),
				),
			},
			{
				Config: testAccDirectoryConfigConfig_certifcateBasedAuthParameters(rName, domain, rUserNameUpdated, rPasswordUpdated, orgUnitDN, string(awstypes.CertificateBasedAuthStatusEnabledNoDirectoryLoginFallback)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDirectoryConfigExists(ctx, t, resourceName, &v2),
					testAccCheckDirectoryConfigNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "directory_name", domain),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_distinguished_names.0", orgUnitDN),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_name", rUserNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "service_account_credentials.0.account_password", rPasswordUpdated),
					resource.TestCheckResourceAttr(resourceName, "certificate_based_auth_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_based_auth_properties.0.status", string(awstypes.CertificateBasedAuthStatusEnabledNoDirectoryLoginFallback)),
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

func testAccCheckDirectoryConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_directory_config" {
				continue
			}

			_, err := tfappstream.FindDirectoryConfigByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream Directory Config %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDirectoryConfigExists(ctx context.Context, t *testing.T, n string, v *awstypes.DirectoryConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		output, err := tfappstream.FindDirectoryConfigByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

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
	fmt.Fprintf(&sb, "OU=%s", orgUnit)
	for dc := range strings.SplitSeq(domainName, ".") {
		fmt.Fprintf(&sb, " DC=%s", dc)
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
    aws_directory_service_directory.test,
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

func testAccDirectoryConfigConfig_certifcateBasedAuthParameters(rName, domain, userName, password, orgUnitDN, status string) string {
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

  certificate_based_auth_properties {
    certificate_authority_arn = aws_acmpca_certificate_authority.test_ca.arn
    status                    = %[5]q
  }

  depends_on = [
    aws_directory_service_directory.test,
    aws_acmpca_certificate_authority.test_ca
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

resource "aws_acmpca_certificate_authority" "test_ca" {
  type       = "ROOT"
  usage_mode = "SHORT_LIVED_CERTIFICATE"

  certificate_authority_configuration {
    key_algorithm     = "RSA_2048"
    signing_algorithm = "SHA256WITHRSA"

    subject {
      common_name = "example.com"
    }
  }
}
`, domain, userName, password, orgUnitDN, status))
}
