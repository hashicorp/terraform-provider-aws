// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAppAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var appauthorization types.AppAuthorization

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, t, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
		},
	})
}

func testAccAppAuthorization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var appauthorization types.AppAuthorization
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appfabric_app_authorization.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, t, resourceName, &appauthorization),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfappfabric.ResourceAppAuthorization, resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAppAuthorization_apiKeyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var appauthorization types.AppAuthorization

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, t, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
			{
				Config: testAccAppAuthorizationConfig_updatedAPIkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, t, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "updatedApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
		},
	})
}

func testAccAppAuthorization_oath2Update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	var appauthorization types.AppAuthorization

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_oath2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, t, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "DROPBOX"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "oauth2"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_id", "ClinentID"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_secret", "SecretforOath2"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
			{
				Config: testAccAppAuthorizationConfig_updatedOath2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, t, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "DROPBOX"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "oauth2"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_id", "newClinentID"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_secret", "newSecretforOath2"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
		},
	})
}

func testAccCheckAppAuthorizationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_app_authorization" {
				continue
			}

			_, err := tfappfabric.FindAppAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["app_bundle_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Fabric App Authorization %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppAuthorizationExists(ctx context.Context, t *testing.T, n string, v *types.AppAuthorization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppFabricClient(ctx)

		output, err := tfappfabric.FindAppAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["app_bundle_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppAuthorizationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  app            = "TERRAFORMCLOUD"
  auth_type      = "apiKey"

  credential {
    api_key_credential {
      api_key = "ApiExampleKey"
    }
  }
  tenant {
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }
}
`, rName)
}

func testAccAppAuthorizationConfig_updatedAPIkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  app            = "TERRAFORMCLOUD"
  auth_type      = "apiKey"

  credential {
    api_key_credential {
      api_key = "updatedApiExampleKey"
    }
  }
  tenant {
    tenant_display_name = "updated"
    tenant_identifier   = "test"
  }
}
`, rName)
}

func testAccAppAuthorizationConfig_oath2(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  app            = "DROPBOX"
  auth_type      = "oauth2"

  credential {
    oauth2_credential {
      client_id     = "ClinentID"
      client_secret = "SecretforOath2"
    }
  }
  tenant {
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }
}
`, rName)
}

func testAccAppAuthorizationConfig_updatedOath2(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_app_bundle" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_app_authorization" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  app            = "DROPBOX"
  auth_type      = "oauth2"

  credential {
    oauth2_credential {
      client_id     = "newClinentID"
      client_secret = "newSecretforOath2"
    }
  }
  tenant {
    tenant_display_name = "updated"
    tenant_identifier   = "test"
  }
}
`, rName)
}
