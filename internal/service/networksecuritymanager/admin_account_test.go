// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworksecuritymanager "github.com/hashicorp/terraform-provider-aws/internal/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAdminAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AdminAccountDetails
	resourceName := "aws_networksecuritymanager_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_account_id"), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.AdminAccountStatusOnboarded)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_scope"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						"admin_account_id":  tfknownvalue.AccountID(),
					}),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "admin_account_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "admin_account_id",
			},
		},
	})
}

func testAccAdminAccount_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AdminAccountDetails
	resourceName := "aws_networksecuritymanager_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceAdminAccount, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// Every optional argument set on creation.
func testAccAdminAccount_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AdminAccountDetails
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(5)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.AdminAccountStatusOnboarded)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_scope"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"firewall_type_scope": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"all_firewall_types_enabled": knownvalue.Bool(false),
									"firewall_types": knownvalue.SetExact([]knownvalue.Check{
										tfknownvalue.StringExact(awstypes.PolicyFirewallTypeWaf),
									}),
								}),
							}),
							"scope_filter": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"include_all": knownvalue.Null(),
									"include_only": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"accounts": knownvalue.SetExact([]knownvalue.Check{
												tfknownvalue.AccountID(),
											}),
											"organizational_units": knownvalue.SetSizeExact(1),
										}),
									}),
									"exclude_only": knownvalue.ListSizeExact(0),
								}),
							}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "admin_account_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "admin_account_id",
				ImportStateVerifyIgnore:              []string{names.AttrTimeouts},
			},
		},
	})
}

// priority is updated in place through PutAdminAccount; 1 and 10 are the bounds.
func testAccAdminAccount_priority(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AdminAccountDetails
	resourceName := "aws_networksecuritymanager_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(1)),
				},
			},
			{
				Config: testAccAdminAccountConfig_basic(10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(10)),
				},
			},
			{
				Config: testAccAdminAccountConfig_basic(5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(5)),
				},
			},
		},
	})
}

// Every scope_filter union member, and the transitions between them.
func testAccAdminAccount_scopeFilter(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AdminAccountDetails
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_admin_account.test"

	scopeFilterCheck := func(includeAll knownvalue.Check, includeOnly, excludeOnly knownvalue.Check) statecheck.StateCheck {
		return statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_scope").AtSliceIndex(0).AtMapKey("scope_filter"), knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"include_all":  includeAll,
				"include_only": includeOnly,
				"exclude_only": excludeOnly,
			}),
		}))
	}
	accountsSelection := knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"accounts":             knownvalue.SetExact([]knownvalue.Check{tfknownvalue.AccountID()}),
			"organizational_units": knownvalue.Null(),
		}),
	})
	ousSelection := knownvalue.ListExact([]knownvalue.Check{
		knownvalue.ObjectExact(map[string]knownvalue.Check{
			"accounts":             knownvalue.Null(),
			"organizational_units": knownvalue.SetSizeExact(1),
		}),
	})
	updateCheck := resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
		},
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_scopeFilterIncludeAll(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					scopeFilterCheck(knownvalue.Bool(true), knownvalue.ListSizeExact(0), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "admin_account_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "admin_account_id",
			},
			// include_all -> include_only (accounts)
			{
				Config: testAccAdminAccountConfig_scopeFilterIncludeOnlyAccounts(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					scopeFilterCheck(knownvalue.Null(), accountsSelection, knownvalue.ListSizeExact(0)),
				},
			},
			// include_only (accounts) -> include_only (organizational units)
			{
				Config: testAccAdminAccountConfig_scopeFilterIncludeOnlyOUs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					scopeFilterCheck(knownvalue.Null(), ousSelection, knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "admin_account_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "admin_account_id",
			},
			// include_only -> exclude_only (accounts)
			{
				Config: testAccAdminAccountConfig_scopeFilterExcludeOnlyAccounts(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					scopeFilterCheck(knownvalue.Null(), knownvalue.ListSizeExact(0), accountsSelection),
				},
			},
			// exclude_only -> include_all
			{
				Config: testAccAdminAccountConfig_scopeFilterIncludeAll(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					scopeFilterCheck(knownvalue.Bool(true), knownvalue.ListSizeExact(0), knownvalue.ListSizeExact(0)),
				},
			},
			// admin_scope removed entirely
			{
				Config: testAccAdminAccountConfig_basic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_scope"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

// Every firewall_type_scope combination.
func testAccAdminAccount_firewallTypeScope(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AdminAccountDetails
	resourceName := "aws_networksecuritymanager_admin_account.test"

	firewallTypeScopeCheck := func(allEnabled knownvalue.Check, firewallTypes knownvalue.Check) statecheck.StateCheck {
		return statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_scope").AtSliceIndex(0).AtMapKey("firewall_type_scope"), knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"all_firewall_types_enabled": allEnabled,
				"firewall_types":             firewallTypes,
			}),
		}))
	}
	updateCheck := resource.ConfigPlanChecks{
		PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
		},
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_firewallTypes(`"WAF"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					firewallTypeScopeCheck(knownvalue.Null(), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.PolicyFirewallTypeWaf),
					})),
				},
			},
			{
				Config: testAccAdminAccountConfig_firewallTypes(`"SHIELD_ADVANCED"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					firewallTypeScopeCheck(knownvalue.Null(), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.PolicyFirewallTypeShieldAdvanced),
					})),
				},
			},
			{
				Config: testAccAdminAccountConfig_firewallTypes(`"WAF", "SHIELD_ADVANCED"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					firewallTypeScopeCheck(knownvalue.Null(), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.PolicyFirewallTypeWaf),
						tfknownvalue.StringExact(awstypes.PolicyFirewallTypeShieldAdvanced),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "admin_account_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "admin_account_id",
			},
			{
				Config: testAccAdminAccountConfig_allFirewallTypesEnabled(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					firewallTypeScopeCheck(knownvalue.Bool(true), knownvalue.Null()),
				},
			},
			{
				Config: testAccAdminAccountConfig_allFirewallTypesEnabled(false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
				ConfigStateChecks: []statecheck.StateCheck{
					firewallTypeScopeCheck(knownvalue.Bool(false), knownvalue.Null()),
				},
			},
		},
	})
}

// Delete then set the same administrator account again immediately: the API
// rejects PutAdminAccount for a while after a removal, which Create retries.
func testAccAdminAccount_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.AdminAccountDetails
	resourceName := "aws_networksecuritymanager_admin_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAdminAccountConfig_basic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccAdminAccountConfig_basic(1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.AdminAccountStatusOnboarded)),
				},
			},
		},
	})
}

// Plan-time validation: nothing is created.
func testAccAdminAccount_validation(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccAdminAccountConfig_basic(0),
				ExpectError: regexache.MustCompile(`Attribute priority value must be between 1 and 10`),
			},
			{
				Config:      testAccAdminAccountConfig_basic(11),
				ExpectError: regexache.MustCompile(`Attribute priority value must be between 1 and 10`),
			},
			{
				Config:      testAccAdminAccountConfig_accountID("12345678901"),
				ExpectError: regexache.MustCompile(`value must be a valid AWS account ID`),
			},
			{
				Config:      testAccAdminAccountConfig_accountID("12345678901a"),
				ExpectError: regexache.MustCompile(`value must be a valid AWS account ID`),
			},
			{
				Config:      testAccAdminAccountConfig_scopeFilterTwoMembers(),
				ExpectError: regexache.MustCompile(`one \(and only one\) of`),
			},
			{
				Config:      testAccAdminAccountConfig_scopeFilterEmpty(),
				ExpectError: regexache.MustCompile(`one \(and only one\) of`),
			},
			{
				Config:      testAccAdminAccountConfig_scopeFilterIncludeAllFalse(),
				ExpectError: regexache.MustCompile(`include_all Value must be "true"`),
			},
		},
	})
}

func testAccCheckAdminAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_admin_account" {
				continue
			}

			_, err := tfnetworksecuritymanager.FindAdminAccountByID(ctx, conn, rs.Primary.Attributes["admin_account_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Admin Account %s still exists", rs.Primary.Attributes["admin_account_id"])
		}

		return nil
	}
}

func testAccCheckAdminAccountExists(ctx context.Context, t *testing.T, n string, v *awstypes.AdminAccountDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		output, err := tfnetworksecuritymanager.FindAdminAccountByID(ctx, conn, rs.Primary.Attributes["admin_account_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckOrganizationManagementAccount(ctx, t)

	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.ListAdminAccountsInput{}
	_, err := conn.ListAdminAccounts(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAdminAccountConfig_base() string {
	return `
data "aws_caller_identity" "current" {}

resource "aws_organizations_aws_service_access" "test" {
  service_principal = "network-security-manager.amazonaws.com"
}
`
}

func testAccAdminAccountConfig_basic(priority int) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = %[1]d
}
`, priority))
}

func testAccAdminAccountConfig_accountID(accountID string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = %[1]q
  priority         = 1
}
`, accountID))
}

func testAccAdminAccountConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 5

  admin_scope {
    firewall_type_scope {
      all_firewall_types_enabled = false
      firewall_types             = ["WAF"]
    }

    scope_filter {
      include_only {
        accounts             = [data.aws_caller_identity.current.account_id]
        organizational_units = [aws_organizations_organizational_unit.test.id]
      }
    }
  }

  timeouts {
    create = "20m"
    update = "20m"
    delete = "20m"
  }
}
`, rName))
}

func testAccAdminAccountConfig_scopeFilterIncludeAll() string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), `
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    scope_filter {
      include_all = true
    }
  }
}
`)
}

func testAccAdminAccountConfig_scopeFilterIncludeOnlyAccounts() string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), `
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    scope_filter {
      include_only {
        accounts = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}
`)
}

func testAccAdminAccountConfig_scopeFilterIncludeOnlyOUs(rName string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    scope_filter {
      include_only {
        organizational_units = [aws_organizations_organizational_unit.test.id]
      }
    }
  }
}
`, rName))
}

func testAccAdminAccountConfig_scopeFilterExcludeOnlyAccounts() string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), `
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    scope_filter {
      exclude_only {
        accounts = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}
`)
}

func testAccAdminAccountConfig_scopeFilterTwoMembers() string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), `
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    scope_filter {
      include_all = true

      include_only {
        accounts = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}
`)
}

func testAccAdminAccountConfig_scopeFilterEmpty() string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), `
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    scope_filter {}
  }
}
`)
}

func testAccAdminAccountConfig_scopeFilterIncludeAllFalse() string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), `
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    scope_filter {
      include_all = false
    }
  }
}
`)
}

func testAccAdminAccountConfig_firewallTypes(firewallTypes string) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    firewall_type_scope {
      firewall_types = [%[1]s]
    }

    scope_filter {
      include_all = true
    }
  }
}
`, firewallTypes))
}

func testAccAdminAccountConfig_allFirewallTypesEnabled(enabled bool) string {
	return acctest.ConfigCompose(testAccAdminAccountConfig_base(), fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  depends_on = [aws_organizations_aws_service_access.test]

  admin_account_id = data.aws_caller_identity.current.account_id
  priority         = 1

  admin_scope {
    firewall_type_scope {
      all_firewall_types_enabled = %[1]t
    }

    scope_filter {
      include_all = true
    }
  }
}
`, enabled))
}
