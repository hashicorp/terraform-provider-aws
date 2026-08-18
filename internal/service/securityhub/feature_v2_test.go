// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
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
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/FeatureV2/basic/"),
				ConfigVariables: config.Variables{},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureV2Exists(ctx, t, resourceName, &feature),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("feature_name"), tfknownvalue.StringExact(awstypes.FeatureNameNetworkScanning)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("feature_status"), tfknownvalue.StringExact(awstypes.FeatureStatusEnabled)),
				},
			},
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/FeatureV2/basic/"),
				ConfigVariables:                      config.Variables{},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "feature_name"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "feature_name",
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
