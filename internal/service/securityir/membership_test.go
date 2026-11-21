// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityir_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/securityir"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityir/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityir "github.com/hashicorp/terraform-provider-aws/internal/service/securityir"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var checkMembershipARN = tfknownvalue.RegionalARNRegexp("security-ir", regexache.MustCompile(`membership/m-[0-9a-z]+`))

func testAccMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v securityir.GetMembershipOutput
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName(t)
	email1, email2 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_basic(rName, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkMembershipARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cover_entire_organization"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("customer_type"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.StringRegexp(regexache.MustCompile(`^m-[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("incident_response_team"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"communication_preferences": knownvalue.Null(),
							names.AttrEmail:             knownvalue.StringExact(email1),
							"job_title":                 knownvalue.StringExact("Security Engineer"),
							names.AttrName:              knownvalue.StringExact("Responder One"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"communication_preferences": knownvalue.Null(),
							names.AttrEmail:             knownvalue.StringExact(email2),
							"job_title":                 knownvalue.StringExact("Security Manager"),
							names.AttrName:              knownvalue.StringExact("Responder Two"),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("membership_activation_timestamp"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("membership_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("membership_status"), tfknownvalue.StringExact(awstypes.MembershipStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("number_of_accounts_covered"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("organizational_units"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v securityir.GetMembershipOutput
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName(t)
	email1, email2 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_basic(rName, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsecurityir.ResourceMembership, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccMembership_membershipName(t *testing.T) {
	ctx := acctest.Context(t)
	var v securityir.GetMembershipOutput
	resourceName := "aws_securityir_membership.test"
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName(t)
	email1, email2 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_basic(rName1, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("membership_name"), knownvalue.StringExact(rName1)),
				},
			},
			{
				Config: testAccMembershipConfig_basic(rName2, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("membership_name"), knownvalue.StringExact(rName2)),
				},
			},
		},
	})
}

func testAccMembership_incidentResponseTeam(t *testing.T) {
	ctx := acctest.Context(t)
	var v securityir.GetMembershipOutput
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName(t)
	email1, email2 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)
	email3, email4 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_basic(rName, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("incident_response_team"), knownvalue.ListSizeExact(2)),
				},
			},
			{
				Config: testAccMembershipConfig_incidentResponseTeam(rName, email1, email3, email4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("incident_response_team"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"communication_preferences": knownvalue.SetExact([]knownvalue.Check{
								tfknownvalue.StringExact(awstypes.CommunicationTypeCaseCreated),
								tfknownvalue.StringExact(awstypes.CommunicationTypeCaseClosed),
							}),
							names.AttrEmail: knownvalue.StringExact(email1),
							"job_title":     knownvalue.StringExact("Security Lead"),
							names.AttrName:  knownvalue.StringExact("Responder One"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"communication_preferences": knownvalue.Null(),
							names.AttrEmail:             knownvalue.StringExact(email3),
							"job_title":                 knownvalue.StringExact("Security Analyst"),
							names.AttrName:              knownvalue.StringExact("Responder Three"),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"communication_preferences": knownvalue.Null(),
							names.AttrEmail:             knownvalue.StringExact(email4),
							"job_title":                 knownvalue.StringExact("Incident Commander"),
							names.AttrName:              knownvalue.StringExact("Responder Four"),
						}),
					})),
				},
			},
		},
	})
}

func testAccMembership_optInFeature(t *testing.T) {
	ctx := acctest.Context(t)
	var v securityir.GetMembershipOutput
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName(t)
	email1, email2 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_basic(rName, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Triage is enabled by default but is not tracked unless configured.
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_in_feature"), knownvalue.SetSizeExact(0)),
				},
			},
			{
				Config:      testAccMembershipConfig_optInFeature(rName, email1, email2, false),
				ExpectError: regexache.MustCompile(`The Triage opt-in feature cannot be disabled`),
			},
			{
				Config: testAccMembershipConfig_optInFeature(rName, email1, email2, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_in_feature"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"feature_name": tfknownvalue.StringExact(awstypes.OptInFeatureNameTriage),
							"is_enabled":   knownvalue.Bool(true),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Only configured opt-in features are tracked, and there is no configuration on import.
				ImportStateVerifyIgnore: []string{"opt_in_feature"},
			},
			{
				Config: testAccMembershipConfig_basic(rName, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("opt_in_feature"), knownvalue.SetSizeExact(0)),
				},
			},
		},
	})
}

func testAccMembership_coverEntireOrganization(t *testing.T) {
	ctx := acctest.Context(t)
	var v securityir.GetMembershipOutput
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName(t)
	email1, email2 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_coverEntireOrganization(rName, email1, email2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cover_entire_organization"), knownvalue.Bool(false)),
				},
			},
			{
				Config: testAccMembershipConfig_coverEntireOrganization(rName, email1, email2, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cover_entire_organization"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("customer_type"), tfknownvalue.StringExact(awstypes.CustomerTypeOrganization)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMembershipConfig_coverEntireOrganization(rName, email1, email2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cover_entire_organization"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

func testAccMembership_organizationalUnits(t *testing.T) {
	ctx := acctest.Context(t)
	var v securityir.GetMembershipOutput
	resourceName := "aws_securityir_membership.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domain := acctest.RandomDomainName(t)
	email1, email2 := acctest.RandomEmailAddress(domain), acctest.RandomEmailAddress(domain)

	ou := func(i int) string {
		return fmt.Sprintf("aws_organizations_organizational_unit.test[%d].id", i)
	}
	ouChecks := func(is ...int) resource.TestCheckFunc {
		var fs []resource.TestCheckFunc
		for _, i := range is {
			fs = append(fs, resource.TestCheckTypeSetElemAttrPair(resourceName, "organizational_units.*", fmt.Sprintf("aws_organizations_organizational_unit.test.%d", i), names.AttrID))
		}
		return resource.ComposeAggregateTestCheckFunc(fs...)
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityIRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMembershipConfig_organizationalUnits(rName, email1, email2, ou(0)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
					ouChecks(0),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("organizational_units"), knownvalue.SetSizeExact(1)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMembershipConfig_organizationalUnits(rName, email1, email2, ou(1), ou(2)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
					ouChecks(1, 2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("organizational_units"), knownvalue.SetSizeExact(2)),
				},
			},
			{
				// Adding more OUs than UpdateMembership accepts in one call.
				Config: testAccMembershipConfig_organizationalUnits(rName, email1, email2, ou(0), ou(1), ou(2), ou(3), ou(4), ou(5), ou(6)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
					ouChecks(0, 1, 2, 3, 4, 5, 6),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("organizational_units"), knownvalue.SetSizeExact(7)),
				},
			},
			{
				Config: testAccMembershipConfig_organizationalUnits(rName, email1, email2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMembershipExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("organizational_units"), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccCheckMembershipDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityIRClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityir_membership" {
				continue
			}

			_, err := tfsecurityir.FindMembershipByID(ctx, conn, rs.Primary.Attributes[names.AttrID])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Security Incident Response Membership %s still exists", rs.Primary.Attributes[names.AttrID])
		}

		return nil
	}
}

func testAccCheckMembershipExists(ctx context.Context, t *testing.T, n string, v *securityir.GetMembershipOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityIRClient(ctx)

		output, err := tfsecurityir.FindMembershipByID(ctx, conn, rs.Primary.Attributes[names.AttrID])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccMembershipConfig_basic(rName, email1, email2 string) string {
	return fmt.Sprintf(`
resource "aws_securityir_membership" "test" {
  membership_name = %[1]q

  incident_response_team {
    name      = "Responder One"
    job_title = "Security Engineer"
    email     = %[2]q
  }

  incident_response_team {
    name      = "Responder Two"
    job_title = "Security Manager"
    email     = %[3]q
  }
}
`, rName, email1, email2)
}

func testAccMembershipConfig_incidentResponseTeam(rName, email1, email3, email4 string) string {
	return fmt.Sprintf(`
resource "aws_securityir_membership" "test" {
  membership_name = %[1]q

  incident_response_team {
    name      = "Responder One"
    job_title = "Security Lead"
    email     = %[2]q

    communication_preferences = ["Case Created", "Case Closed"]
  }

  incident_response_team {
    name      = "Responder Three"
    job_title = "Security Analyst"
    email     = %[3]q
  }

  incident_response_team {
    name      = "Responder Four"
    job_title = "Incident Commander"
    email     = %[4]q
  }
}
`, rName, email1, email3, email4)
}

func testAccMembershipConfig_optInFeature(rName, email1, email2 string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_securityir_membership" "test" {
  membership_name = %[1]q

  incident_response_team {
    name      = "Responder One"
    job_title = "Security Engineer"
    email     = %[2]q
  }

  incident_response_team {
    name      = "Responder Two"
    job_title = "Security Manager"
    email     = %[3]q
  }

  opt_in_feature {
    feature_name = "Triage"
    is_enabled   = %[4]t
  }
}
`, rName, email1, email2, enabled)
}

func testAccMembershipConfig_coverEntireOrganization(rName, email1, email2 string, cover bool) string {
	return fmt.Sprintf(`
resource "aws_securityir_membership" "test" {
  membership_name           = %[1]q
  cover_entire_organization = %[4]t

  incident_response_team {
    name      = "Responder One"
    job_title = "Security Engineer"
    email     = %[2]q
  }

  incident_response_team {
    name      = "Responder Two"
    job_title = "Security Manager"
    email     = %[3]q
  }
}
`, rName, email1, email2, cover)
}

func testAccMembershipConfig_organizationalUnits(rName, email1, email2 string, ous ...string) string {
	var organizationalUnits string
	if len(ous) > 0 {
		organizationalUnits = fmt.Sprintf("organizational_units = [%s]", strings.Join(ous, ", "))
	}

	return fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

resource "aws_organizations_organizational_unit" "test" {
  count = 7

  name      = "%[1]s-${count.index}"
  parent_id = data.aws_organizations_organization.current.roots[0].id
}

resource "aws_securityir_membership" "test" {
  membership_name = %[1]q
  %[4]s

  incident_response_team {
    name      = "Responder One"
    job_title = "Security Engineer"
    email     = %[2]q
  }

  incident_response_team {
    name      = "Responder Two"
    job_title = "Security Manager"
    email     = %[3]q
  }
}
`, rName, email1, email2, organizationalUnits)
}
