// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPManagedLoginBranding_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ManagedLoginBrandingType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBranding_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("managed_login_branding_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("settings_all"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "managed_login_branding_id",
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "user_pool_id", "managed_login_branding_id"),
			},
		},
	})
}

func TestAccCognitoIDPManagedLoginBranding_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var client awstypes.ManagedLoginBrandingType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cognito_managed_login_branding.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckManagedLoginBrandingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccManagedLoginBranding_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckManagedLoginBrandingExists(ctx, resourceName, &client),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcognitoidp.ResourceManagedLoginBranding, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckManagedLoginBrandingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_managed_login_branding" {
				continue
			}

			_, err := tfcognitoidp.FindManagedLoginBrandingByThreePartKey(ctx, conn, rs.Primary.Attributes["user_pool_id"], rs.Primary.Attributes["managed_login_branding_id"], false)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito Managed Login Branding %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckManagedLoginBrandingExists(ctx context.Context, n string, v *awstypes.ManagedLoginBrandingType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindManagedLoginBrandingByThreePartKey(ctx, conn, rs.Primary.Attributes["user_pool_id"], rs.Primary.Attributes["managed_login_branding_id"], false)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccManagedLoginBranding_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

resource "aws_cognito_user_pool_client" "test" {
  name                = %[1]q
  user_pool_id        = aws_cognito_user_pool.test.id
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}

resource "aws_cognito_managed_login_branding" "test" {
  client_id    = aws_cognito_user_pool_client.test.id
  user_pool_id = aws_cognito_user_pool.test.id

  use_cognito_provided_values = true
}
`, rName)
}
