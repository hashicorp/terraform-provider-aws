// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneEnvironmentBlueprintConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintConfigurationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "environment_blueprint_id",
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprintConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfdatazone.ResourceEnvironmentBlueprintConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprintConfiguration_enabled_regions(t *testing.T) {
	ctx := acctest.Context(t)
	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_enabled_regions(domainName, endpoints.UsEast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.0", endpoints.UsEast1RegionID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintConfigurationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "environment_blueprint_id",
			},
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_enabled_regions(domainName, endpoints.ApSoutheast2RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.0", endpoints.ApSoutheast2RegionID),
				),
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprintConfiguration_manage_access_role_arn(t *testing.T) {
	ctx := acctest.Context(t)
	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_manage_access_role_arn(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "manage_access_role_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintConfigurationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "environment_blueprint_id",
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprintConfiguration_provisioning_role_arn(t *testing.T) {
	ctx := acctest.Context(t)
	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_provisioning_role_arn(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintConfigurationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "environment_blueprint_id",
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprintConfiguration_regional_parameters(t *testing.T) {
	ctx := acctest.Context(t)
	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, endpoints.UsWest2RegionID, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "regional_parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.%%", endpoints.UsWest2RegionID), "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.key1", endpoints.UsWest2RegionID), acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccEnvironmentBlueprintConfigurationImportStateIdFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "environment_blueprint_id",
			},
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, endpoints.UsWest2RegionID, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "regional_parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.%%", endpoints.UsWest2RegionID), "1"),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.key2", endpoints.UsWest2RegionID), acctest.CtValue2),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/43316.
func TestAccDataZoneEnvironmentBlueprintConfiguration_upgradeFromV5_100_0(t *testing.T) {
	ctx := acctest.Context(t)
	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.DataZoneServiceID),
		CheckDestroy: testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, endpoints.UsWest2RegionID, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config:      testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, endpoints.UsWest2RegionID, acctest.CtKey1, acctest.CtValue1),
				ExpectError: regexache.MustCompile(`Incorrect attribute value type`),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, endpoints.UsWest2RegionID, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, t, resourceName, &environmentblueprintconfiguration),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_environment_blueprint_configuration" {
				continue
			}

			_, err := tfdatazone.FindEnvironmentBlueprintConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes["domain_id"], rs.Primary.Attributes["environment_blueprint_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataZone Environment Blueprint Configuration %s/%s still exists", rs.Primary.Attributes["domain_id"], rs.Primary.Attributes["environment_blueprint_id"])
		}

		return nil
	}
}

func testAccCheckEnvironmentBlueprintConfigurationExists(ctx context.Context, t *testing.T, n string, v *datazone.GetEnvironmentBlueprintConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DataZoneClient(ctx)

		output, err := tfdatazone.FindEnvironmentBlueprintConfigurationByTwoPartKey(ctx, conn, rs.Primary.Attributes["domain_id"], rs.Primary.Attributes["environment_blueprint_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEnvironmentBlueprintConfigurationImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["domain_id"], rs.Primary.Attributes["environment_blueprint_id"]), nil
	}
}

func testAccEnvironmentBlueprintConfigurationConfig_basic(domainName string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentBlueprintDataSourceConfig_basic(domainName),
		`
resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  enabled_regions          = []
}
`,
	)
}

func testAccEnvironmentBlueprintConfigurationConfig_enabled_regions(domainName, enabledRegion string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentBlueprintDataSourceConfig_basic(domainName),
		fmt.Sprintf(`
resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  enabled_regions          = [%[1]q]
}
`, enabledRegion),
	)
}

func testAccEnvironmentBlueprintConfigurationConfig_manage_access_role_arn(domainName string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentBlueprintDataSourceConfig_basic(domainName),
		`
resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  manage_access_role_arn   = aws_iam_role.domain_execution_role.arn
  enabled_regions          = []
}
`,
	)
}

func testAccEnvironmentBlueprintConfigurationConfig_provisioning_role_arn(domainName string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentBlueprintDataSourceConfig_basic(domainName),
		`
resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  provisioning_role_arn    = aws_iam_role.domain_execution_role.arn
  enabled_regions          = []
}
`,
	)
}

func testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, region, key, value string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentBlueprintDataSourceConfig_basic(domainName),
		fmt.Sprintf(`
resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  enabled_regions          = []
  regional_parameters = {
    %[1]q = {
      %[2]q = %[3]q
    }
  }
}
`, region, key, value),
	)
}
