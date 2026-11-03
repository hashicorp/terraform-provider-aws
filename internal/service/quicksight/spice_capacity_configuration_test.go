// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccSPICECapacityConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_quicksight_spice_capacity_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSPICECapacityConfigurationConfig_basic,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("purchase_mode"), knownvalue.StringExact(string(awstypes.PurchaseModeManual))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAWSAccountID),
				ImportStateVerifyIdentifierAttribute: names.AttrAWSAccountID,
				// The purchase mode cannot be read back from the QuickSight API.
				ImportStateVerifyIgnore: []string{"purchase_mode"},
			},
		},
	})
}

func testAccSPICECapacityConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_quicksight_spice_capacity_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccSPICECapacityConfigurationConfig_purchaseMode(string(awstypes.PurchaseModeAutoPurchase)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("purchase_mode"), knownvalue.StringExact(string(awstypes.PurchaseModeAutoPurchase))),
				},
			},
			{
				Config: testAccSPICECapacityConfigurationConfig_purchaseMode(string(awstypes.PurchaseModeManual)),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("purchase_mode"), knownvalue.StringExact(string(awstypes.PurchaseModeManual))),
				},
			},
		},
	})
}

const testAccSPICECapacityConfigurationConfig_basic = `
resource "aws_quicksight_spice_capacity_configuration" "test" {}
`

func testAccSPICECapacityConfigurationConfig_purchaseMode(purchaseMode string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_spice_capacity_configuration" "test" {
  purchase_mode = %[1]q
}
`, purchaseMode)
}
