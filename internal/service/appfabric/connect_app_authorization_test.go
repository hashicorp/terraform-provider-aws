// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccConnectAppAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_connect_app_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectAppAuthorizationConfig_basic(rName),
				Check:  resource.ComposeTestCheckFunc(
				// testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
				// resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
				// resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
				// resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
				// resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", acctest.Ct1),
				// resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
				// resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
				// resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "test"),
				// resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             false,
				ImportStateVerify:       false,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

// func testConnectAppAuthorization_disappears(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	var appauthorization types.AppAuthorization
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_appfabric_app_authorization.test"

// 	resource.Test(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		CheckDestroy:             testAccCheckAppAuthorizationDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAppAuthorizationConfig_basic(rName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckAppAuthorizationExists(ctx, resourceName, &appauthorization),
// 					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceAppAuthorization, resourceName),
// 					resource.TestCheckResourceAttr(resourceName, "app", "TERRAFORMCLOUD"),
// 					resource.TestCheckResourceAttr(resourceName, "auth_type", "apiKey"),
// 					resource.TestCheckResourceAttr(resourceName, "credential.#", acctest.Ct1),
// 					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.#", acctest.Ct1),
// 					resource.TestCheckResourceAttr(resourceName, "credential.0.api_key_credential.0.api_key", "ApiExampleKey"),
// 					resource.TestCheckResourceAttr(resourceName, "tenant.#", acctest.Ct1),
// 					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", "test"),
// 					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", "test"),
// 				),
// 				ExpectNonEmptyPlan: true,
// 			},
// 		},
// 	})
// }

// func testAccCheckConnectAppAuthorizationDestroy(ctx context.Context) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

// 		for _, rs := range s.RootModule().Resources {
// 			if rs.Type != "aws_appfabric_app_authorization" {
// 				continue
// 			}

// 			_, err := tfappfabric.FindAppAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["app_bundle_arn"])

// 			if tfresource.NotFound(err) {
// 				continue
// 			}

// 			if err != nil {
// 				return err
// 			}

// 			return fmt.Errorf("App Fabric App Authorization %s still exists", rs.Primary.ID)
// 		}

// 		return nil
// 	}
// }

// func testAccCheckConnectAppAuthorizationExists(ctx context.Context, n string, v *types.AppAuthorization) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[n]
// 		if !ok {
// 			return fmt.Errorf("Not found: %s", n)
// 		}

// 		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

// 		output, err := tfappfabric.FindAppAuthorizationByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrARN], rs.Primary.Attributes["app_bundle_arn"])

// 		if err != nil {
// 			return err
// 		}

// 		*v = *output

// 		return nil
// 	}
// }

func testAccConnectAppAuthorizationConfig_basic(rName string) string {
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
      api_key = "TestAPiKey"
    }
  }
  tenant {
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }
}

resource "aws_appfabric_connect_app_authorization" "test" {
  app_bundle_arn = aws_appfabric_app_bundle.test.arn
  app_authorization_arn = aws_appfabric_app_authorization.test.arn
}

`, rName)
}
