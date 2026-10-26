// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcloudwatchomni "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkOrganizationDomainARN                = tfknownvalue.RegionalARNRegexp("cloudwatch", regexache.MustCompile(`organization-domain/d-[0-9a-z]{1,25}`))
	checkOrganizationDomainARNAlternateRegion = tfknownvalue.RegionalARNAlternateRegionRegexp("cloudwatch", regexache.MustCompile(`organization-domain/d-[0-9a-z]{1,25}`))
	checkOrganizationDomainID                 = knownvalue.StringRegexp(regexache.MustCompile(`^d-[0-9a-z]{1,25}$`))
	checkOrganizationID                       = knownvalue.StringRegexp(regexache.MustCompile(`^o-[0-9a-z]{10,32}$`))
)

func testAccOrganizationDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationDomain
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_organization_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationDomain(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkOrganizationDomainARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_endpoint_urls"), knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("domain_access_role_arn"), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_endpoint_url"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("domain_id"), checkOrganizationDomainID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_application_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_provider_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_providers"), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.IdentityProviderIam),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("organization_id"), checkOrganizationID),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrOwnerAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.DomainStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "domain_id"),
				ImportStateVerifyIdentifierAttribute: "domain_id",
			},
		},
	})
}

func testAccOrganizationDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationDomain
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_organization_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationDomain(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcloudwatchomni.ResourceOrganizationDomain, resourceName),
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

func testAccOrganizationDomain_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationDomain
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_organization_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationDomain(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName1),
					"domain_name":   config.StringVariable(rName1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName1)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/name/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName1),
					"domain_name":   config.StringVariable(rName2),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName2)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_endpoint_urls"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringRegexp(regexache.MustCompile(`^https://` + rName2 + `\.`)),
					})),
				},
			},
		},
	})
}

func testAccOrganizationDomain_domainAccessRoleARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationDomain
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_organization_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationDomain(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/domain_access_role_arn/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"role_index":    config.IntegerVariable(0),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("domain_access_role_arn"), "aws_iam_role.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/domain_access_role_arn/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"role_index":    config.IntegerVariable(1),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("domain_access_role_arn"), "aws_iam_role.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func testAccOrganizationDomain_identityCenter(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.OrganizationDomain
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatchomni_organization_domain.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckOrganizationDomain(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudWatchOmniServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDomainDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_application_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_provider_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_providers"), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.IdentityProviderIam),
					})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/identity_provider_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_application_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_provider_configuration"), knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New("identity_provider_configuration").AtSliceIndex(0).AtMapKey("identity_center_configuration").AtSliceIndex(0).AtMapKey("identity_center_instance_arn"),
						"data.aws_ssoadmin_instances.test", tfjsonpath.New(names.AttrARNs).AtSliceIndex(0),
						compare.ValuesSame(),
					),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_providers"), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.IdentityProviderIam),
						tfknownvalue.StringExact(awstypes.IdentityProviderIdc),
					})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/identity_provider_configuration/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "domain_id"),
				ImportStateVerifyIdentifierAttribute: "domain_id",
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/OrganizationDomain/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationDomainExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_application_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_provider_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_providers"), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.IdentityProviderIam),
					})),
				},
			},
		},
	})
}

// testAccPreCheckOrganizationDomain skips unless the caller is the management account of an organization
// that has trusted access enabled for CloudWatch Omni and no CloudWatch Omni domain.
func testAccPreCheckOrganizationDomain(ctx context.Context, t *testing.T) {
	acctest.PreCheckOrganizationManagementAccount(ctx, t)
	acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "observabilityadmin.amazonaws.com")
	testAccPreCheck(ctx, t)
}

func testAccCheckOrganizationDomainDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatchomni_organization_domain" {
				continue
			}

			_, err := tfcloudwatchomni.FindOrganizationDomainByID(ctx, conn, rs.Primary.Attributes["domain_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Omni Organization Domain %s still exists", rs.Primary.Attributes["domain_id"])
		}

		return nil
	}
}

func testAccCheckOrganizationDomainExists(ctx context.Context, t *testing.T, n string, v *awstypes.OrganizationDomain) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

		output, err := tfcloudwatchomni.FindOrganizationDomainByID(ctx, conn, rs.Primary.Attributes["domain_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
