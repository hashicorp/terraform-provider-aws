// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfverifiedpermissions "github.com/hashicorp/terraform-provider-aws/internal/service/verifiedpermissions"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsPolicyStoreAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var alias verifiedpermissions.GetPolicyStoreAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	aliasName := testAccPolicyStoreAliasName(rName)
	resourceName := "aws_verifiedpermissions_policy_store_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreAliasExists(ctx, t, resourceName, &alias),
					resource.TestCheckResourceAttr(
						resourceName,
						"alias_name",
						aliasName,
					),
					resource.TestCheckResourceAttrSet(
						resourceName,
						"policy_store_id",
					),
					resource.TestCheckResourceAttrSet(
						resourceName,
						names.AttrARN,
					),
					resource.TestCheckResourceAttrSet(
						resourceName,
						"created_at",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						names.AttrState,
						"Active",
					),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionCreate,
						),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(
						resourceName,
						map[string]knownvalue.Check{
							"account_id": knownvalue.NotNull(),
							"alias_name": knownvalue.StringExact(aliasName),
							"region":     knownvalue.StringExact(acctest.Region()),
						},
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New("alias_name"),
						knownvalue.StringExact(aliasName),
					),
					statecheck.ExpectKnownValue(
						resourceName,
						tfjsonpath.New(names.AttrState),
						knownvalue.StringExact("Active"),
					),
				},
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: acctest.AttrImportStateIdFunc(
					resourceName,
					"alias_name",
				),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "alias_name",
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStoreAlias_policyStoreIDReplace(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var alias verifiedpermissions.GetPolicyStoreAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	aliasName := testAccPolicyStoreAliasName(rName)
	resourceName := "aws_verifiedpermissions_policy_store_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreAliasConfig_policyStoreIndex(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreAliasExists(
						ctx,
						t,
						resourceName,
						&alias,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"alias_name",
						aliasName,
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"policy_store_id",
						"aws_verifiedpermissions_policy_store.test.0",
						"policy_store_id",
					),
				),
			},
			{
				Config: testAccPolicyStoreAliasConfig_policyStoreIndex(rName, 1),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreAliasExists(
						ctx,
						t,
						resourceName,
						&alias,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"alias_name",
						aliasName,
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"policy_store_id",
						"aws_verifiedpermissions_policy_store.test.1",
						"policy_store_id",
					),
				),
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStoreAlias_aliasNameReplace(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var alias verifiedpermissions.GetPolicyStoreAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy_store_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreAliasConfig_named(rName, rName+"-first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreAliasExists(
						ctx,
						t,
						resourceName,
						&alias,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"alias_name",
						testAccPolicyStoreAliasName(rName+"-first"),
					),
				),
			},
			{
				Config: testAccPolicyStoreAliasConfig_named(rName, rName+"-second"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreAliasExists(
						ctx,
						t,
						resourceName,
						&alias,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"alias_name",
						testAccPolicyStoreAliasName(rName+"-second"),
					),
				),
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStoreAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var alias verifiedpermissions.GetPolicyStoreAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy_store_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreAliasExists(
						ctx,
						t,
						resourceName,
						&alias,
					),
					acctest.CheckFrameworkResourceDisappears(
						ctx,
						t,
						tfverifiedpermissions.ResourcePolicyStoreAlias,
						resourceName,
					),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionCreate,
						),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionCreate,
						),
					},
				},
			},
		},
	})
}

func testAccCheckPolicyStoreAliasDestroy(
	ctx context.Context,
	t *testing.T,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_policy_store_alias" {
				continue
			}

			aliasName := rs.Primary.Attributes["alias_name"]
			if aliasName == "" {
				continue
			}

			_, err := tfverifiedpermissions.FindPolicyStoreAliasByName(
				ctx,
				conn,
				aliasName,
			)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(
				names.VerifiedPermissions,
				create.ErrActionCheckingDestroyed,
				tfverifiedpermissions.ResNamePolicyStoreAlias,
				aliasName,
				errors.New("not destroyed"),
			)
		}

		return nil
	}
}

func testAccCheckPolicyStoreAliasExists(
	ctx context.Context,
	t *testing.T,
	resourceName string,
	alias *verifiedpermissions.GetPolicyStoreAliasOutput,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return create.Error(
				names.VerifiedPermissions,
				create.ErrActionCheckingExistence,
				tfverifiedpermissions.ResNamePolicyStoreAlias,
				resourceName,
				errors.New("not found"),
			)
		}

		aliasName := rs.Primary.Attributes["alias_name"]
		if aliasName == "" {
			return create.Error(
				names.VerifiedPermissions,
				create.ErrActionCheckingExistence,
				tfverifiedpermissions.ResNamePolicyStoreAlias,
				resourceName,
				errors.New("alias_name not set"),
			)
		}

		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)

		output, err := tfverifiedpermissions.FindPolicyStoreAliasByName(
			ctx,
			conn,
			aliasName,
		)
		if err != nil {
			return create.Error(
				names.VerifiedPermissions,
				create.ErrActionCheckingExistence,
				tfverifiedpermissions.ResNamePolicyStoreAlias,
				aliasName,
				err,
			)
		}

		if alias != nil {
			*alias = *output
		}

		return nil
	}
}

func testAccPolicyStoreAliasName(name string) string {
	return fmt.Sprintf("policy-store-alias/%s", name)
}

func testAccPolicyStoreAliasConfig_basic(rName string) string {
	return testAccPolicyStoreAliasConfig_named(rName, rName)
}

func testAccPolicyStoreAliasConfig_named(
	rName string,
	aliasResourceName string,
) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = %[1]q

  validation_settings {
    mode = "OFF"
  }
}

resource "aws_verifiedpermissions_policy_store_alias" "test" {
  alias_name      = %[2]q
  policy_store_id = aws_verifiedpermissions_policy_store.test.policy_store_id
}
`, rName, testAccPolicyStoreAliasName(aliasResourceName))
}

func testAccPolicyStoreAliasConfig_policyStoreIndex(
	rName string,
	policyStoreIndex int,
) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  count = 2

  description = "%[1]s-${count.index}"

  validation_settings {
    mode = "OFF"
  }
}

resource "aws_verifiedpermissions_policy_store_alias" "test" {
  alias_name      = %[2]q
  policy_store_id = aws_verifiedpermissions_policy_store.test[%[3]d].policy_store_id
}
`, rName, testAccPolicyStoreAliasName(rName), policyStoreIndex)
}
