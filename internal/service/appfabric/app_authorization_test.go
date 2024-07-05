// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAppAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var appauthorization types.AppAuthorization

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_app_authorization.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceAppAuthorization, resourceName),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var appauthorization types.AppAuthorization

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
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
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "updatedApiExampleKey"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var appauthorization types.AppAuthorization

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_oath2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "DROPBOX"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "oauth2"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_id", "ClinentID"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_secret", "SecretforOath2"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
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
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
					resource.TestCheckResourceAttr(resourceName, "app", "DROPBOX"),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "oauth2"),
					resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_id", "newClinentID"),
					resource.TestCheckResourceAttr(resourceName, "credential.0.oauth2_credential.0.client_secret", "newSecretforOath2"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
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

func testAccAppAuthorization_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var appauthorization types.AppAuthorization

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
			{
				Config: testAccAppAuthorizationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credential"},
			},
			{
				Config: testAccAppAuthorizationConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
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

func testAccCheckAppAuthorizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_app_authorization" {
				continue
			}

			_, err := tfappfabric.FindAppAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["app_bundle_arn"])

			if tfresource.NotFound(err) {
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

func testAccCheckAppAuthorizationExists(ctx context.Context, n string, v *types.AppAuthorization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

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

func testAccAppAuthorizationConfig_tags1(rName, tagKey1, tagValue1 string) string {
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
      api_key = "apiexamplekeytest"
    }
  }
  tenant {
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAppAuthorizationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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
      api_key = "apiexamplekeytest"
    }
  }
  tenant {
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
