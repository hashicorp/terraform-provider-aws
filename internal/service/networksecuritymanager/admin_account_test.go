// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
			testAccPreCheckAdminAccount(ctx, t)
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), tfknownvalue.AccountID()),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAccountID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
			},
			// priority is updated in place through PutAdminAccount; 1 and 10 are the bounds.
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
			// Delete then set the same administrator account again immediately: the API
			// rejects PutAdminAccount for a while after a removal, which Create retries.
			{
				Taint:  []string{resourceName},
				Config: testAccAdminAccountConfig_basic(5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(5)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.AdminAccountStatusOnboarded)),
				},
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
			testAccPreCheckAdminAccount(ctx, t)
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
			testAccPreCheckAdminAccount(ctx, t)
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAccountID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
				ImportStateVerifyIgnore:              []string{names.AttrTimeouts},
			},
		},
	})
}

// Every scope_filter union member and firewall_type_scope combination, and the
// transitions between them. One test case, so the administrator account is
// created and removed only once.
func testAccAdminAccount_adminScope(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

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
			testAccPreCheckAdminAccount(ctx, t)
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAccountID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAccountID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
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
			// firewall_type_scope added
			{
				Config: testAccAdminAccountConfig_firewallTypes(`"WAF"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAdminAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: updateCheck,
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAccountID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
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

// Plan-time validation: nothing is created, so unlike the other tests this one
// does not have to be serialized.
func TestAccNetworkSecurityManagerAdminAccount_validation(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckAdminAccount(ctx, t)
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
			{
				Config:      testAccAdminAccountConfig_firewallTypes(""),
				ExpectError: regexache.MustCompile(`firewall_types set must\s+contain at least 1 elements`),
			},
			{
				Config:      testAccAdminAccountConfig_scopeFilterIncludeOnlyEmptyAccounts(),
				ExpectError: regexache.MustCompile(`accounts set must\s+contain at least 1 elements`),
			},
		},
	})
}

const resNameAdminAccount = "Admin Account"

func testAccCheckAdminAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_admin_account" {
				continue
			}

			id := rs.Primary.Attributes[names.AttrAccountID]
			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)
			_, err := tfnetworksecuritymanager.FindAdminAccountByID(ctx, conn, id)
			if retry.NotFound(err) || tfnetworksecuritymanager.IsAccountNotOnboardedError(err) {
				continue
			}
			if err != nil {
				return create.Error(names.NetworkSecurityManager, create.ErrActionCheckingDestroyed, resNameAdminAccount, id, err)
			}

			return create.Error(names.NetworkSecurityManager, create.ErrActionCheckingDestroyed, resNameAdminAccount, id, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAdminAccountExists(ctx context.Context, t *testing.T, n string, v *awstypes.AdminAccountDetails) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.NetworkSecurityManager, create.ErrActionCheckingExistence, resNameAdminAccount, n, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.NetworkSecurityManager, create.ErrActionCheckingExistence, resNameAdminAccount, n, errors.New("not set"))
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)
		id := rs.Primary.Attributes[names.AttrAccountID]
		output, err := tfnetworksecuritymanager.FindAdminAccountByID(ctx, conn, id)
		if err != nil {
			return create.Error(names.NetworkSecurityManager, create.ErrActionCheckingExistence, resNameAdminAccount, id, err)
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckAdminAccount(ctx context.Context, t *testing.T) {
	acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "network-security-manager.amazonaws.com")
}

func testAccAdminAccountConfig_base() string {
	return `data "aws_caller_identity" "current" {}`
}

func testAccAdminAccountConfig_basic(priority int) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = %[1]d
}
`, priority), testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_accountID(accountID string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = %[1]q
  priority   = 1
}
`, accountID), testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_full(rName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 5

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

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}
`, rName), testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterIncludeAll() string {
	return acctest.ConfigCompose(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {
      include_all = true
    }
  }
}
`, testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterIncludeOnlyAccounts() string {
	return acctest.ConfigCompose(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {
      include_only {
        accounts = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}
`, testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterIncludeOnlyOUs(rName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {
      include_only {
        organizational_units = [aws_organizations_organizational_unit.test.id]
      }
    }
  }
}

data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = data.aws_organizations_organization.current.roots[0].id
}
`, rName), testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterExcludeOnlyAccounts() string {
	return acctest.ConfigCompose(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {
      exclude_only {
        accounts = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}
`, testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterTwoMembers() string {
	return acctest.ConfigCompose(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {
      include_all = true

      include_only {
        accounts = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}
`, testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterEmpty() string {
	return acctest.ConfigCompose(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {}
  }
}
`, testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterIncludeAllFalse() string {
	return acctest.ConfigCompose(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {
      include_all = false
    }
  }
}
`, testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_scopeFilterIncludeOnlyEmptyAccounts() string {
	return acctest.ConfigCompose(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    scope_filter {
      include_only {
        accounts = []
      }
    }
  }
}
`, testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_firewallTypes(firewallTypes string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    firewall_type_scope {
      firewall_types = [%[1]s]
    }

    scope_filter {
      include_all = true
    }
  }
}
`, firewallTypes), testAccAdminAccountConfig_base())
}

func testAccAdminAccountConfig_allFirewallTypesEnabled(enabled bool) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
resource "aws_networksecuritymanager_admin_account" "test" {
  account_id = data.aws_caller_identity.current.account_id
  priority   = 1

  admin_scope {
    firewall_type_scope {
      all_firewall_types_enabled = %[1]t
    }

    scope_filter {
      include_all = true
    }
  }
}
`, enabled), testAccAdminAccountConfig_base())
}
