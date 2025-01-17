// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53domains "github.com/hashicorp/terraform-provider-aws/internal/service/route53domains"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDomain_basic(t *testing.T) {
	acctest.Skip(t, "Route 53 domain registration acceptance test skipped")

	ctx := acctest.Context(t)
	resourceName := "aws_route53domains_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := rName + ".click"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccDomainImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrDomainName,
				ImportStateVerifyIgnore:              []string{"duration_in_years"},
			},
		},
	})
}

func testAccDomain_disappears(t *testing.T) {
	acctest.Skip(t, "Route 53 domain registration acceptance test skipped")

	ctx := acctest.Context(t)
	resourceName := "aws_route53domains_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := rName + ".click"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfroute53domains.ResourceDomain, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDomain_tags(t *testing.T) {
	acctest.Skip(t, "Route 53 domain registration acceptance test skipped")

	ctx := acctest.Context(t)
	resourceName := "aws_route53domains_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := rName + ".click"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53DomainsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainConfig_tags1(domainName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccDomainImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrDomainName,
				ImportStateVerifyIgnore:              []string{"duration_in_years"},
			},
			{
				Config: testAccDomainConfig_tags2(domainName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccDomainConfig_tags1(domainName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func testAccCheckDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53domains_domain" {
				continue
			}

			_, err := tfroute53domains.FindDomainDetailByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route 53 Domains Domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53DomainsClient(ctx)

		_, err := tfroute53domains.FindDomainDetailByName(ctx, conn, rs.Primary.Attributes[names.AttrDomainName])

		return err
	}
}

func testAccDomainImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes[names.AttrDomainName], nil
	}
}

func testAccDomainConfig_basic(domainName string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name = %[1]q

  admin_contact {
    address_line_1    = "101 2nd St #700"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551230"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551240"
    state             = "CA"
    zip_code          = "94105"
  }

  registrant_contact {
    address_line_1    = "101 2nd St #702"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551232"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551242"
    state             = "CA"
    zip_code          = "94105"
  }

  tech_contact {
    address_line_1    = "101 2nd St #703"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551233"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551243"
    state             = "CA"
    zip_code          = "94105"
  }
}
`, domainName, acctest.DefaultEmailAddress)
}

func testAccDomainConfig_tags1(domainName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name   = %[1]q
  transfer_lock = false

  admin_contact {
    address_line_1    = "101 2nd St #700"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551230"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551240"
    state             = "CA"
    zip_code          = "94105"
  }

  registrant_contact {
    address_line_1    = "101 2nd St #702"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551232"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551242"
    state             = "CA"
    zip_code          = "94105"
  }

  tech_contact {
    address_line_1    = "101 2nd St #703"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551233"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551243"
    state             = "CA"
    zip_code          = "94105"
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, domainName, acctest.DefaultEmailAddress, tag1Key, tag1Value)
}

func testAccDomainConfig_tags2(domainName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_route53domains_domain" "test" {
  domain_name   = %[1]q
  transfer_lock = false

  admin_contact {
    address_line_1    = "101 2nd St #700"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551230"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551240"
    state             = "CA"
    zip_code          = "94105"
  }

  registrant_contact {
    address_line_1    = "101 2nd St #702"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551232"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551242"
    state             = "CA"
    zip_code          = "94105"
  }

  tech_contact {
    address_line_1    = "101 2nd St #703"
    city              = "San Francisco"
    contact_type      = "COMPANY"
    country_code      = "US"
    email             = %[2]q
    fax               = "+1.4155551233"
    first_name        = "Terraform"
    last_name         = "Team"
    organization_name = "HashiCorp"
    phone_number      = "+1.4155551243"
    state             = "CA"
    zip_code          = "94105"
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, domainName, acctest.DefaultEmailAddress, tag1Key, tag1Value, tag2Key, tag2Value)
}
