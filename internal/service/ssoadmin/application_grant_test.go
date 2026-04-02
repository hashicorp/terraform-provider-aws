// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminApplicationGrant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_grant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationGrantConfig_refreshToken(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant_type", "refresh_token"),
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

func TestAccSSOAdminApplicationGrant_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_grant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationGrantConfig_refreshToken(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfssoadmin.ResourceApplicationGrant, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminApplicationGrant_disappears_Application(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_grant.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationGrantConfig_refreshToken(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfssoadmin.ResourceApplication, applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminApplicationGrant_jwtBearerUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_grant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationGrantConfig_jwtBearer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.0.authorized_audiences.#", "1"),
				),
			},
			{
				Config: testAccApplicationGrantConfig_jwtBearerUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.0.authorized_audiences.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.0.authorized_audiences.0", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.0.authorized_audiences.1", "https://example2.com"),
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

func TestAccSSOAdminApplicationGrant_refreshToken(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_grant.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationGrantConfig_refreshToken(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant_type", "refresh_token"),
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

func TestAccSSOAdminApplicationGrant_jwtBearer(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_grant.test"
	ttiResourceName := "aws_ssoadmin_trusted_token_issuer.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationGrantDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationGrantConfig_jwtBearer(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.0.authorized_audiences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.0.authorized_audiences.0", "https://example.com"),
					resource.TestCheckResourceAttrPair(resourceName, "grant.0.jwt_bearer.0.authorized_token_issuers.0.trusted_token_issuer_arn", ttiResourceName, names.AttrARN),
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

func testAccCheckApplicationGrantDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_application_grant" {
				continue
			}

			_, err := tfssoadmin.FindApplicationGrantByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplicationGrant, rs.Primary.ID, err)
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplicationGrant, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationGrantExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationGrant, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationGrant, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

		_, err := tfssoadmin.FindApplicationGrantByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationGrant, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccApplicationGrantConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName, testAccApplicationProviderARN)
}

func testAccApplicationGrantConfig_refreshToken(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationGrantConfigBase(rName),
		`
resource "aws_ssoadmin_application_grant" "test" {
  application_arn = aws_ssoadmin_application.test.application_arn
  grant_type      = "refresh_token"
}
`)
}

func testAccApplicationGrantConfig_jwtBearer(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationGrantConfigBase(rName),
		fmt.Sprintf(`
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

resource "aws_ssoadmin_application_grant" "test" {
  application_arn = aws_ssoadmin_application.test.application_arn
  grant_type      = "urn:ietf:params:oauth:grant-type:jwt-bearer"

  grant {
    jwt_bearer {
      authorized_token_issuers {
        authorized_audiences      = ["https://example.com"]
        trusted_token_issuer_arn  = aws_ssoadmin_trusted_token_issuer.test.arn
      }
    }
  }
}
`, rName))
}

func testAccApplicationGrantConfig_jwtBearerUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationGrantConfigBase(rName),
		fmt.Sprintf(`
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

resource "aws_ssoadmin_application_grant" "test" {
  application_arn = aws_ssoadmin_application.test.application_arn
  grant_type      = "urn:ietf:params:oauth:grant-type:jwt-bearer"

  grant {
    jwt_bearer {
      authorized_token_issuers {
        authorized_audiences      = ["https://example.com", "https://example2.com"]
        trusted_token_issuer_arn  = aws_ssoadmin_trusted_token_issuer.test.arn
      }
    }
  }
}
`, rName))
}
