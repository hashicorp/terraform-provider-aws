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

func TestAccSSOAdminApplicationGrant_authorizationCode(t *testing.T) {
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
				Config: testAccApplicationGrantConfig_authorizationCode(rName, "http://127.0.0.1/auth/callback"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_arn", applicationResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "grant_type", "authorization_code"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.authorization_code.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.authorization_code.0.redirect_uris.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.authorization_code.0.redirect_uris.0", "http://127.0.0.1/auth/callback"),
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

func TestAccSSOAdminApplicationGrant_authorizationCode_update(t *testing.T) {
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
				Config: testAccApplicationGrantConfig_authorizationCode(rName, "http://127.0.0.1/auth/callback"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.0.authorization_code.0.redirect_uris.0", "http://127.0.0.1/auth/callback"),
				),
			},
			{
				Config: testAccApplicationGrantConfig_authorizationCode(rName, "http://127.0.0.1/oauth/callback"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant.0.authorization_code.0.redirect_uris.0", "http://127.0.0.1/oauth/callback"),
				),
			},
		},
	})
}

func TestAccSSOAdminApplicationGrant_tokenExchange(t *testing.T) {
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
				Config: testAccApplicationGrantConfig_tokenExchange(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "grant_type", "urn:ietf:params:oauth:grant-type:token-exchange"),
					resource.TestCheckResourceAttr(resourceName, "grant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "grant.0.token_exchange.#", "1"),
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
				Config: testAccApplicationGrantConfig_authorizationCode(rName, "http://127.0.0.1/auth/callback"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationGrantExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfssoadmin.ResourceApplicationGrant, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

func testAccApplicationGrantConfig_authorizationCode(rName, redirectURI string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_application_grant" "test" {
  application_arn = aws_ssoadmin_application.test.arn
  grant_type      = "authorization_code"

  grant {
    authorization_code {
      redirect_uris = [%[3]q]
    }
  }
}
`, rName, testAccApplicationProviderARN, redirectURI)
}

func testAccApplicationGrantConfig_tokenExchange(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_application_grant" "test" {
  application_arn = aws_ssoadmin_application.test.arn
  grant_type      = "urn:ietf:params:oauth:grant-type:token-exchange"

  grant {
    token_exchange {}
  }
}
`, rName, testAccApplicationProviderARN)
}
