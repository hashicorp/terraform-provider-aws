// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountAccessEntitlementsDataSource_byPrincipal(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_entitlements.test"
	entitlementName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementsDataSourceConfig_byPrincipal(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey("entitlement_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrCreatedAt), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						entitlementName, tfjsonpath.New("principal_id"),
						dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey("principal_id"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						entitlementName, tfjsonpath.New(names.AttrRoleARN),
						dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrRoleARN),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccAccountAccessEntitlementsDataSource_byRole(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_entitlements.test"
	entitlementName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementsDataSourceConfig_byRole(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey("entitlement_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrCreatedAt), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						entitlementName, tfjsonpath.New(names.AttrRoleARN),
						dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrRoleARN),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						entitlementName, tfjsonpath.New(names.AttrAccountID),
						dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrAccountID),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccAccountAccessEntitlementsDataSource_byAccount(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_accountaccess_entitlements.test"
	entitlementName := "aws_accountaccess_entitlement.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEntitlementDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEntitlementsDataSourceConfig_byAccount(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey("entitlement_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrCreatedAt), knownvalue.NotNull()),
					statecheck.CompareValuePairs(
						entitlementName, tfjsonpath.New(names.AttrAccountID),
						dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrAccountID),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						entitlementName, tfjsonpath.New(names.AttrRoleARN),
						dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey(names.AttrRoleARN),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						entitlementName, tfjsonpath.New("principal_id"),
						dataSourceName, tfjsonpath.New("entitlements").AtSliceIndex(0).AtMapKey("principal_id"),
						compare.ValuesSame(),
					),
				},
			},
		},
	})
}

func testAccEntitlementsDataSourceConfig_byPrincipal(rName string) string {
	return acctest.ConfigCompose(testAccEntitlementConfig_user(rName), `
data "aws_accountaccess_entitlements" "test" {
  application_arn = aws_accountaccess_entitlement.test.application_arn
  principal_id    = aws_accountaccess_entitlement.test.principal_id
  principal_type  = aws_accountaccess_entitlement.test.principal_type
}
`)
}

func testAccEntitlementsDataSourceConfig_byRole(rName string) string {
	return acctest.ConfigCompose(testAccEntitlementConfig_user(rName), `
data "aws_accountaccess_entitlements" "test" {
  application_arn = aws_accountaccess_entitlement.test.application_arn
  role_arn        = aws_accountaccess_entitlement.test.role_arn
}
`)
}

func testAccEntitlementsDataSourceConfig_byAccount(rName string) string {
	return acctest.ConfigCompose(testAccEntitlementConfig_user(rName), `
data "aws_accountaccess_entitlements" "test" {
  application_arn = aws_accountaccess_entitlement.test.application_arn
  account_id      = aws_accountaccess_entitlement.test.account_id
}
`)
}
