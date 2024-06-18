// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminTrustedTokenIssuer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeTrustedTokenIssuerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_trusted_token_issuer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustedTokenIssuerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustedTokenIssuerConfigBase_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustedTokenIssuerExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_type", string(types.TrustedTokenIssuerTypeOidcJwt)),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.claim_attribute_path", names.AttrEmail),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.identity_store_attribute_path", "emails.value"),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.issuer_url", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.jwks_retrieval_option", string(types.JwksRetrievalOptionOpenIdDiscovery)),
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

func TestAccSSOAdminTrustedTokenIssuer_update(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeTrustedTokenIssuerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_trusted_token_issuer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustedTokenIssuerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustedTokenIssuerConfigBase_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustedTokenIssuerExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.claim_attribute_path", names.AttrEmail),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.identity_store_attribute_path", "emails.value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrustedTokenIssuerConfigBase_basicUpdated(rNameUpdated, names.AttrName, "userName"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustedTokenIssuerExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.claim_attribute_path", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "trusted_token_issuer_configuration.0.oidc_jwt_configuration.0.identity_store_attribute_path", "userName"),
				),
			},
		},
	})
}

func TestAccSSOAdminTrustedTokenIssuer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeTrustedTokenIssuerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_trusted_token_issuer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustedTokenIssuerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustedTokenIssuerConfigBase_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustedTokenIssuerExists(ctx, resourceName, &application),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceTrustedTokenIssuer, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminTrustedTokenIssuer_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var application ssoadmin.DescribeTrustedTokenIssuerOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_trusted_token_issuer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckSSOAdminInstances(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrustedTokenIssuerDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTrustedTokenIssuerConfigBase_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustedTokenIssuerExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTrustedTokenIssuerConfigBase_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustedTokenIssuerExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTrustedTokenIssuerConfigBase_tags(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrustedTokenIssuerExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckTrustedTokenIssuerExists(ctx context.Context, name string, trustedTokenIssuer *ssoadmin.DescribeTrustedTokenIssuerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameTrustedTokenIssuer, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameTrustedTokenIssuer, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)
		resp, err := tfssoadmin.FindTrustedTokenIssuerByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameTrustedTokenIssuer, rs.Primary.ID, err)
		}

		*trustedTokenIssuer = *resp

		return nil
	}
}

func testAccCheckTrustedTokenIssuerDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_trusted_token_issuer" {
				continue
			}

			_, err := tfssoadmin.FindTrustedTokenIssuerByARN(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameTrustedTokenIssuer, rs.Primary.ID, err)
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameTrustedTokenIssuer, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccTrustedTokenIssuerConfigBase_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_trusted_token_issuer" "test" {
  name                      = %[1]q
  instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = "email"
      identity_store_attribute_path = "emails.value"
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }
}
`, rName)
}

func testAccTrustedTokenIssuerConfigBase_basicUpdated(rNameUpdated, claimAttributePath, identityStoreAttributePath string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_trusted_token_issuer" "test" {
  name                      = %[1]q
  instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = %[2]q
      identity_store_attribute_path = %[3]q
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }
}
`, rNameUpdated, claimAttributePath, identityStoreAttributePath)
}

func testAccTrustedTokenIssuerConfigBase_tags(rName, tagKey, tagValue string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_trusted_token_issuer" "test" {
  name                      = %[1]q
  instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = "email"
      identity_store_attribute_path = "emails.value"
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey, tagValue)
}

func testAccTrustedTokenIssuerConfigBase_tags2(rName, tagKey, tagValue, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_trusted_token_issuer" "test" {
  name                      = %[1]q
  instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = "email"
      identity_store_attribute_path = "emails.value"
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey, tagValue, tagKey2, tagValue2)
}
