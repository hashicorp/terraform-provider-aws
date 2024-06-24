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

func testAccAppAuthorizationConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization_connection.test"
	appBudleResourceName := "aws_appfabric_app_bundle.test"
	appAuthorization := "aws_appfabric_app_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "app_bundle_arn", appBudleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "app_authorization_arn", appAuthorization, names.AttrARN),
				),
			},
		},
	})
}
func testAccAppAuthorizationConnection_OAuth2(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization_connection.test"
	appBudleResourceName := "aws_appfabric_app_bundle.test"
	appAuthorization := "aws_appfabric_app_authorization.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, names.USEast1RegionID, names.APNortheast1RegionID, names.EUWest1RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConnectionConfig_OAuth2(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "app_bundle_arn", appBudleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "app_authorization_arn", appAuthorization, names.AttrARN),
				),
			},
		},
	})
}

func testAccAppAuthorizationConnectionConfig_basic(rName string) string {
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
      api_key = "ApiKeyTest"
    }
  }
  tenant {
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }
}

resource "aws_appfabric_app_authorization_connection" "test" {
  app_bundle_arn        = aws_appfabric_app_bundle.test.arn
  app_authorization_arn = aws_appfabric_app_authorization.test.arn
}
`, rName)
}

func testAccAppAuthorizationConnectionConfig_OAuth2(rName string) string {
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
    tenant_display_name = "test"
    tenant_identifier   = "test"
  }
}

resource "aws_appfabric_app_authorization_connection" "test" {
  app_bundle_arn        = aws_appfabric_app_bundle.test.arn
  app_authorization_arn = aws_appfabric_app_authorization.test.arn
  auth_request {
    code         = "testcode"
    redirect_uri = aws_appfabric_app_authorization.test.auth_url
  }

}
`, rName)
}
