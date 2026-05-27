// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountImportStep(n string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      n,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"close_on_deletion",
			"create_govcloud",
			"govcloud_id",
		},
	}
}

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_basic(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "organizations", regexache.MustCompile(`account/`+organizationIDRegexPattern+`/\d{12}$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, email),
					resource.TestCheckResourceAttrSet(resourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccAccount_CloseOnDeletion(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_closeOnDeletion(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "organizations", regexache.MustCompile(`account/`+organizationIDRegexPattern+`/\d{12}$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, email),
					resource.TestCheckResourceAttr(resourceName, "govcloud_id", ""),
					resource.TestCheckResourceAttrSet(resourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccAccount_ParentID(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	var v awstypes.Account
	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"
	parentIdResourceName1 := "aws_organizations_organizational_unit.test1"
	parentIdResourceName2 := "aws_organizations_organizational_unit.test2"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_parentId1(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName1, names.AttrID),
				),
			},
			testAccAccountImportStep(resourceName),
			{
				Config: testAccAccountConfig_parentId2(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "parent_id", parentIdResourceName2, names.AttrID),
				),
			},
		},
	})
}

func testAccAccount_AccountUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	var v awstypes.Account
	rInt := acctest.RandInt(t)

	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf_acctest_+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"
	newName := fmt.Sprintf("tf_acctest_renamed_%d", rInt)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_closeOnDeletion(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "organizations", regexache.MustCompile(`account/`+organizationIDRegexPattern+`/\d{12}$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrEmail, email),
					resource.TestCheckResourceAttrSet(resourceName, "joined_method"),
					acctest.CheckResourceAttrRFC3339(resourceName, "joined_timestamp"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, name),
					resource.TestCheckResourceAttrSet(resourceName, "parent_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			testAccAccountImportStep(resourceName),
			{
				Config: testAccAccountConfig_closeOnDeletion(newName, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, newName),
				),
			},
		},
	})
}

func testAccAccount_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	var v awstypes.Account
	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)
	resourceName := "aws_organizations_account.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_tags1(name, email, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			testAccAccountImportStep(resourceName),
			{
				Config: testAccAccountConfig_tags2(name, email, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccAccountConfig_tags1(name, email, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccAccount_govCloud(t *testing.T) {
	ctx := acctest.Context(t)
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_govCloud(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "govcloud_id"),
				),
			},
			testAccAccountImportStep(resourceName),
		},
	})
}

func testAccOrganizationsAccount_identitySerial(t *testing.T) {
	t.Helper()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:    testAccOrganizationsAccount_Identity_basic,
		"ExistingResource": testAccOrganizationsAccount_Identity_ExistingResource_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccOrganizationsAccount_Identity_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OrganizationsServiceID),
		CheckDestroy:             testAccCheckAccountDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				Config: testAccAccountConfig_closeOnDeletion(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrID:        knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 2: Import command
			{
				Config:            testAccAccountConfig_closeOnDeletion(name, email),
				ImportStateKind:   resource.ImportCommandWithID,
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"close_on_deletion",
					"create_govcloud",
					"govcloud_id",
				},
			},

			// Step 3: Import block with Import ID
			{
				Config:          testAccAccountConfig_closeOnDeletion(name, email),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithID,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					},
				},
				ExpectNonEmptyPlan: true,
			},

			// Step 4: Import block with Resource Identity
			{
				Config:          testAccAccountConfig_closeOnDeletion(name, email),
				ResourceName:    resourceName,
				ImportState:     true,
				ImportStateKind: resource.ImportBlockWithResourceIdentity,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrID), knownvalue.NotNull()),
					},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccOrganizationsAccount_Identity_ExistingResource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v awstypes.Account
	resourceName := "aws_organizations_account.test"
	orgsEmailDomain := acctest.SkipIfEnvVarNotSet(t, "TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")
	rInt := acctest.RandInt(t)
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.OrganizationsServiceID),
		CheckDestroy: testAccCheckAccountDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create pre-Identity
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccAccountConfig_closeOnDeletion(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},

			// Step 2: v6.0 Identity set on refresh
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccAccountConfig_closeOnDeletion(name, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrID:        knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrID)),
				},
			},

			// Step 3: Current version
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccAccountConfig_closeOnDeletion(name, email),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrID:        knownvalue.NotNull(),
					}),
					statecheck.ExpectIdentityValueMatchesState(resourceName, tfjsonpath.New(names.AttrID)),
				},
			},
		},
	})
}

func testAccCheckAccountDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_account" {
				continue
			}

			_, err := tforganizations.FindAccountByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AWS Organizations Account %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccountExists(ctx context.Context, t *testing.T, n string, v *awstypes.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OrganizationsClient(ctx)

		output, err := tforganizations.FindAccountByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAccountConfig_basic(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name  = %[1]q
  email = %[2]q
}
`, name, email)
}

func testAccAccountConfig_closeOnDeletion(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountConfig_parentId1(name, email string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  parent_id         = aws_organizations_organizational_unit.test1.id
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountConfig_parentId2(name, email string) string {
	return fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test1" {
  name      = "test1"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_organizational_unit" "test2" {
  name      = "test2"
  parent_id = data.aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  parent_id         = aws_organizations_organizational_unit.test2.id
  close_on_deletion = true
}
`, name, email)
}

func testAccAccountConfig_tags1(name, email, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  close_on_deletion = true

  tags = {
    %[3]q = %[4]q
  }
}
`, name, email, tagKey1, tagValue1)
}

func testAccAccountConfig_tags2(name, email, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name              = %[1]q
  email             = %[2]q
  close_on_deletion = true

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, name, email, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAccountConfig_govCloud(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name            = %[1]q
  email           = %[2]q
  create_govcloud = true
}
`, name, email)
}
