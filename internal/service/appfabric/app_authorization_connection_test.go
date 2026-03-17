// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAppAuthorizationConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization_connection.test"
	appBudleResourceName := "aws_appfabric_app_bundle.test"
	appAuthorization := "aws_appfabric_app_authorization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	// See https://docs.aws.amazon.com/appfabric/latest/adminguide/terraform.html#terraform-appfabric-connecting.
	tenantID := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_TENANT_ID")
	serviceAccountToken := acctest.SkipIfEnvVarNotSet(t, "AWS_APPFABRIC_TERRAFORMCLOUD_SERVICE_ACCOUNT_TOKEN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConnectionConfig_basic(rName, tenantID, serviceAccountToken),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppAuthorizationConnectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "app"),
					resource.TestCheckResourceAttrPair(resourceName, "app_bundle_arn", appBudleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "app_authorization_arn", appAuthorization, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "auth_request.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tenant.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tenant.0.tenant_identifier", tenantID),
				),
			},
		},
	})
}
func testAccAppAuthorizationConnection_OAuth2(t *testing.T) {
	acctest.Skip(t, "Currently not able to test")

	ctx := acctest.Context(t)
	resourceName := "aws_appfabric_app_authorization_connection.test"
	appBudleResourceName := "aws_appfabric_app_bundle.test"
	appAuthorization := "aws_appfabric_app_authorization.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.ApNortheast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAppAuthorizationConnectionConfig_OAuth2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppAuthorizationConnectionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "app_bundle_arn", appBudleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "app_authorization_arn", appAuthorization, names.AttrARN),
				),
			},
		},
	})
}

func testAccCheckAppAuthorizationConnectionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppFabricClient(ctx)

		_, err := tfappfabric.FindAppAuthorizationConnectionByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_authorization_arn"], rs.Primary.Attributes["app_bundle_arn"])

		return err
	}
}

func testAccAppAuthorizationConnectionConfig_basic(rName, tenantID, serviceAccountToken string) string {
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
      api_key = %[3]q
    }
  }

  tenant {
    tenant_display_name = %[1]q
    tenant_identifier   = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_appfabric_app_authorization_connection" "test" {
  app_bundle_arn        = aws_appfabric_app_bundle.test.arn
  app_authorization_arn = aws_appfabric_app_authorization.test.arn
}
`, rName, tenantID, serviceAccountToken)
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
