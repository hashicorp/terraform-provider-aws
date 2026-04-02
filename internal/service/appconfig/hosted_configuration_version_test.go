// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigHostedConfigurationVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedConfigurationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedConfigurationVersionExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "appconfig", regexache.MustCompile(`application/[0-9a-z]{4,7}/configurationprofile/[0-9a-z]{4,7}/hostedconfigurationversion/[0-9]+`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, "aws_appconfig_application.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_profile_id", "aws_appconfig_configuration_profile.test", "configuration_profile_id"),
					resource.TestCheckResourceAttr(resourceName, names.AttrContent, "{\"foo\":\"bar\"}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrContentType, "application/json"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "version_number", "1"),
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

func TestAccAppConfigHostedConfigurationVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_appconfig_hosted_configuration_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHostedConfigurationVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHostedConfigurationVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostedConfigurationVersionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappconfig.ResourceHostedConfigurationVersion(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHostedConfigurationVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppConfigClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appconfig_hosted_configuration_version" {
				continue
			}

			_, err := tfappconfig.FindHostedConfigurationVersionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["configuration_profile_id"], flex.StringValueToInt32Value(rs.Primary.Attributes["version_number"]))

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppConfig Hosted Configuration Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHostedConfigurationVersionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppConfigClient(ctx)

		_, err := tfappconfig.FindHostedConfigurationVersionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrApplicationID], rs.Primary.Attributes["configuration_profile_id"], flex.StringValueToInt32Value(rs.Primary.Attributes["version_number"]))

		return err
	}
}

func testAccHostedConfigurationVersionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccConfigurationProfileConfig_name(rName),
		fmt.Sprintf(`
resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = %q
}
`, rName))
}
