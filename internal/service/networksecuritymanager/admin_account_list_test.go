// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Only the current account can be the administrator account in these tests,
// so a single resource is listed. The include_resource query runs in the same
// test case, so the administrator account is created and removed only once.
// A Query step reuses the query file of the preceding non-query step, so a
// no-op apply of the list_include_resource configuration comes first.
func testAccAdminAccount_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_networksecuritymanager_admin_account.test"
	identity := tfstatecheck.Identity()

	acctest.Test(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		CheckDestroy:             testAccCheckAdminAccountDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/AdminAccount/list_basic/"),
				ConfigStateChecks: []statecheck.StateCheck{
					identity.GetIdentity(resourceName),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("admin_account_id"), tfknownvalue.AccountID()),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AdminAccount/list_basic/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), tfknownvalue.AccountID()),
					tfquerycheck.ExpectNoResourceObject(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks())),
				},
			},

			// Step 3: Switch to the include_resource query; the resource is unchanged
			{
				ConfigDirectory: config.StaticDirectory("testdata/AdminAccount/list_include_resource/"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},

			// Step 4: Query with include_resource
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/AdminAccount/list_include_resource/"),
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc(resourceName, identity.Checks()),
					querycheck.ExpectResourceDisplayName(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), tfknownvalue.AccountID()),
					querycheck.ExpectResourceKnownValues(resourceName, tfqueryfilter.ByResourceIdentityFunc(identity.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("admin_account_id"), tfknownvalue.AccountID()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("admin_scope"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.AdminAccountStatusOnboarded)),
					}),
				},
			},
		},
	})
}
