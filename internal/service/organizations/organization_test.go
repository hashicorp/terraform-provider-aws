// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccOrganization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "accounts.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.arn", resourceName, "master_account_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.email", resourceName, "master_account_email"),
					resource.TestCheckResourceAttrPair(resourceName, "accounts.0.id", resourceName, "master_account_id"),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "organizations", regexache.MustCompile(`organization/o-.+`)),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "feature_set", string(awstypes.OrganizationFeatureSetAll)),
					acctest.MatchResourceAttrGlobalARN(resourceName, "master_account_arn", "organizations", regexache.MustCompile(`account/o-.+/.+`)),
					resource.TestMatchResourceAttr(resourceName, "master_account_email", regexache.MustCompile(`.+@.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "master_account_id"),
					resource.TestCheckResourceAttr(resourceName, "non_master_accounts.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "roots.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "roots.0.id", regexache.MustCompile(`r-[0-9a-z]{4,32}`)),
					resource.TestCheckResourceAttrSet(resourceName, "roots.0.name"),
					resource.TestCheckResourceAttrSet(resourceName, "roots.0.arn"),
					resource.TestCheckResourceAttr(resourceName, "roots.0.policy_types.#", acctest.Ct0),
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

func testAccOrganization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tforganizations.ResourceOrganization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganization_serviceAccessPrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_serviceAccessPrincipals1("config.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "config.amazonaws.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_serviceAccessPrincipals2("config.amazonaws.com", "ds.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "config.amazonaws.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "ds.amazonaws.com"),
				),
			},
			{
				Config: testAccOrganizationConfig_serviceAccessPrincipals1("fms.amazonaws.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "aws_service_access_principals.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "aws_service_access_principals.*", "fms.amazonaws.com"),
				),
			},
		},
	})
}

func testAccOrganization_EnabledPolicyTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(string(awstypes.PolicyTypeServiceControlPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct0),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(string(awstypes.PolicyTypeAiservicesOptOutPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", string(awstypes.PolicyTypeAiservicesOptOutPolicy)),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(string(awstypes.PolicyTypeServiceControlPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", string(awstypes.PolicyTypeServiceControlPolicy)),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(string(awstypes.PolicyTypeBackupPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", string(awstypes.PolicyTypeBackupPolicy)),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(string(awstypes.PolicyTypeTagPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.0", string(awstypes.PolicyTypeTagPolicy)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct0),
				),
			},
			{
				Config: testAccOrganizationConfig_enabledPolicyTypes1(string(awstypes.PolicyTypeTagPolicy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "enabled_policy_types.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccOrganization_FeatureSet(t *testing.T) {
	ctx := acctest.Context(t)
	var organization awstypes.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_featureSet(string(awstypes.OrganizationFeatureSetConsolidatedBilling)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &organization),
					resource.TestCheckResourceAttr(resourceName, "feature_set", string(awstypes.OrganizationFeatureSetConsolidatedBilling)),
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

func testAccOrganization_FeatureSetForcesNew(t *testing.T) {
	ctx := acctest.Context(t)
	var beforeValue, afterValue awstypes.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_featureSet(string(awstypes.OrganizationFeatureSetAll)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &beforeValue),
					resource.TestCheckResourceAttr(resourceName, "feature_set", string(awstypes.OrganizationFeatureSetAll)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccOrganizationConfig_featureSet(string(awstypes.OrganizationFeatureSetConsolidatedBilling)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &afterValue),
					resource.TestCheckResourceAttr(resourceName, "feature_set", string(awstypes.OrganizationFeatureSetConsolidatedBilling)),
					testAccOrganizationRecreated(&beforeValue, &afterValue),
				),
			},
		},
	})
}

func testAccOrganization_FeatureSetUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var beforeValue, afterValue awstypes.Organization
	resourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOrganizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationConfig_featureSet(string(awstypes.OrganizationFeatureSetConsolidatedBilling)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &beforeValue),
					resource.TestCheckResourceAttr(resourceName, "feature_set", string(awstypes.OrganizationFeatureSetConsolidatedBilling)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccOrganizationConfig_featureSet(string(awstypes.OrganizationFeatureSetAll)),
				ExpectNonEmptyPlan: true, // See note below on this perpetual difference
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationExists(ctx, resourceName, &afterValue),
					// The below check cannot be performed here because the user must confirm the change
					// via Console. Until then, the FeatureSet will not actually be toggled to ALL
					// and will continue to show as CONSOLIDATED_BILLING when calling DescribeOrganization
					// resource.TestCheckResourceAttr(resourceName, "feature_set", string(awstypes.OrganizationFeatureSetAll)),
					testAccOrganizationNotRecreated(&beforeValue, &afterValue),
				),
			},
		},
	})
}

func testAccCheckOrganizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_organization" {
				continue
			}

			_, err := tforganizations.FindOrganization(ctx, conn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return errors.New("Organizations Organization still exists")
		}

		return nil
	}
}

func testAccCheckOrganizationExists(ctx context.Context, n string, v *awstypes.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsClient(ctx)

		output, err := tforganizations.FindOrganization(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccOrganizationRecreated(before, after *awstypes.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(before.Id) == aws.ToString(after.Id) {
			return fmt.Errorf("Organization (%s) not recreated", aws.ToString(before.Id))
		}
		return nil
	}
}

func testAccOrganizationNotRecreated(before, after *awstypes.Organization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(before.Id) != aws.ToString(after.Id) {
			return fmt.Errorf("Organization (%s) recreated", aws.ToString(before.Id))
		}
		return nil
	}
}

const testAccOrganizationConfig_basic = `
resource "aws_organizations_organization" "test" {}
`

func testAccOrganizationConfig_serviceAccessPrincipals1(principal1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = [%[1]q]
}
`, principal1)
}

func testAccOrganizationConfig_serviceAccessPrincipals2(principal1, principal2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = [%[1]q, %[2]q]
}
`, principal1, principal2)
}

func testAccOrganizationConfig_enabledPolicyTypes1(policyType1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  enabled_policy_types = [%[1]q]
}
`, policyType1)
}

func testAccOrganizationConfig_featureSet(featureSet string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  feature_set = %[1]q
}
`, featureSet)
}
