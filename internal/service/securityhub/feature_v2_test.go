// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccFeatureV2_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.FeatureDetail
	resourceName := "aws_securityhub_feature_v2.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureV2Config_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureV2Exists(ctx, t, resourceName, &feature),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("feature_name"), knownvalue.StringExact(string(awstypes.FeatureNameNetworkScanning))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("feature_status"), knownvalue.StringExact(string(awstypes.FeatureStatusEnabled))),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "feature_name"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "feature_name",
			},
		},
	})
}

func testAccFeatureV2_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var feature awstypes.FeatureDetail
	resourceName := "aws_securityhub_feature_v2.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureV2Config_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureV2Exists(ctx, t, resourceName, &feature),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsecurityhub.ResourceFeatureV2, resourceName),
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

func testAccCheckFeatureV2Exists(ctx context.Context, t *testing.T, n string, v *awstypes.FeatureDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindFeatureV2ByName(ctx, conn, awstypes.FeatureName(rs.Primary.Attributes["feature_name"]))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFeatureV2Destroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_securityhub_feature_v2" {
				continue
			}

			_, err := tfsecurityhub.FindFeatureV2ByName(ctx, conn, awstypes.FeatureName(rs.Primary.Attributes["feature_name"]))

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Security Hub V2 Feature %s still exists", rs.Primary.Attributes["feature_name"])
		}

		return nil
	}
}

const testAccFeatureV2Config_basic = `
resource "aws_securityhub_account_v2" "test" {}

resource "aws_securityhub_feature_v2" "test" {
  feature_name = "NETWORK_SCANNING"

  depends_on = [aws_securityhub_account_v2.test]
}
`
