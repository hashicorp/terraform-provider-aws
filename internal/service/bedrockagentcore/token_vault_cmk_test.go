// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTokenVaultCMK_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetTokenVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_token_vault_cmk.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccTokenVaultCMKConfig_customerManaged(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTokenVaultExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kms_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key_type":          tfknownvalue.StringExact(awstypes.KeyTypeCustomerManagedKey),
							names.AttrKMSKeyARN: knownvalue.NotNull(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_vault_id"), knownvalue.StringExact("default")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "token_vault_id"),
				ImportStateVerifyIdentifierAttribute: "token_vault_id",
			},
			{
				Config: testAccTokenVaultCMKConfig_serviceManaged(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTokenVaultExists(ctx, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("kms_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"key_type":          tfknownvalue.StringExact(awstypes.KeyTypeServiceManagedKey),
							names.AttrKMSKeyARN: knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("token_vault_id"), knownvalue.StringExact("default")),
				},
			},
		},
	})
}

func testAccCheckTokenVaultExists(ctx context.Context, n string, v *bedrockagentcorecontrol.GetTokenVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindTokenVaultByID(ctx, conn, rs.Primary.Attributes["token_vault_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccTokenVaultCMKConfig_customerManaged(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_token_vault_cmk" "test" {
  kms_configuration {
    key_type    = "CustomerManagedKey"
    kms_key_arn = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccTokenVaultCMKConfig_serviceManaged(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_token_vault_cmk" "test" {
  token_vault_id = "default"

  kms_configuration {
    key_type = "ServiceManagedKey"
  }
}
`, rName)
}
