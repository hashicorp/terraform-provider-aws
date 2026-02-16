// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcognitoidp "github.com/hashicorp/terraform-provider-aws/internal/service/cognitoidp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPResourceServer_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceServer awstypes.ResourceServerType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identifier := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_resource_server.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceServerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerConfig_basic(identifier, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceServerExists(ctx, t, resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, names.AttrIdentifier, identifier),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "0"),
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

func TestAccCognitoIDPResourceServer_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceServer awstypes.ResourceServerType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identifier := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_resource_server.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceServerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerConfig_basic(identifier, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceServerExists(ctx, t, resourceName, &resourceServer),
					acctest.CheckSDKResourceDisappears(ctx, t, tfcognitoidp.ResourceResourceServer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCognitoIDPResourceServer_scope(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceServer awstypes.ResourceServerType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identifier := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_resource_server.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceServerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerConfig_scope(identifier, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceServerExists(ctx, t, resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "2"),
				),
			},
			{
				Config: testAccResourceServerConfig_scopeUpdate(identifier, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceServerExists(ctx, t, resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure we can remove scope completely
			{
				Config: testAccResourceServerConfig_basic(identifier, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceServerExists(ctx, t, resourceName, &resourceServer),
					resource.TestCheckResourceAttr(resourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope_identifiers.#", "0"),
				),
			},
		},
	})
}

func TestAccCognitoIDPResourceServer_nameChange(t *testing.T) {
	ctx := acctest.Context(t)
	var resourceServer awstypes.ResourceServerType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identifier := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cognito_resource_server.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIdentityProvider(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourceServerDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServerConfig_basic(identifier, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceServerExists(ctx, t, resourceName, &resourceServer),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
				},
			},
			{
				Config: testAccResourceServerConfig_nameUpdate(identifier, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceServerExists(ctx, t, resourceName, &resourceServer),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName+" updated")),
				},
			},
		},
	})
}

func testAccCheckResourceServerExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResourceServerType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		output, err := tfcognitoidp.FindResourceServerByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrIdentifier])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckResourceServerDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CognitoIDPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cognito_resource_server" {
				continue
			}

			_, err := tfcognitoidp.FindResourceServerByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrUserPoolID], rs.Primary.Attributes[names.AttrIdentifier])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Cognito Resource Server %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceServerConfig_basic(identifier, rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "test" {
  identifier   = %[1]q
  name         = %[2]q
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[2]q
}
`, identifier, rName)
}

func testAccResourceServerConfig_nameUpdate(identifier, rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "test" {
  identifier   = %[1]q
  name         = "%[2]s updated"
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[2]q
}
`, identifier, rName)
}

func testAccResourceServerConfig_scope(identifier, rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "test" {
  identifier = %[1]q
  name       = %[2]q

  scope {
    scope_name        = "scope_1_name"
    scope_description = "scope_1_description"
  }

  scope {
    scope_name        = "scope_2_name"
    scope_description = "scope_2_description"
  }

  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[2]q
}
`, identifier, rName)
}

func testAccResourceServerConfig_scopeUpdate(identifier, rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_resource_server" "test" {
  identifier = %[1]q
  name       = %[2]q

  scope {
    scope_name        = "scope_1_name_updated"
    scope_description = "scope_1_description"
  }

  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[2]q
}
`, identifier, rName)
}
