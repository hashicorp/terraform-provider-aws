// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneEnvironmentBlueprintConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.#", acctest.Ct0),
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
	domainName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_basic(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceEnvironmentBlueprintConfiguration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataZoneEnvironmentBlueprintConfiguration_enabled_regions(t *testing.T) {
	ctx := acctest.Context(t)

	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_enabled_regions(domainName, names.USEast1RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.0", names.USEast1RegionID),
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
				Config: testAccEnvironmentBlueprintConfigurationConfig_enabled_regions(domainName, names.APSoutheast2RegionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "enabled_regions.0", names.APSoutheast2RegionID),
				),
			},
		},
	})
}

func testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_environment_blueprint_configuration" {
				continue
			}

			_, err := conn.GetEnvironmentBlueprintConfiguration(ctx, &datazone.GetEnvironmentBlueprintConfigurationInput{
				DomainIdentifier:               aws.String(rs.Primary.Attributes["domain_id"]),
				EnvironmentBlueprintIdentifier: aws.String(rs.Primary.Attributes["environment_blueprint_id"]),
			})
			if tfdatazone.IsResourceMissing(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironmentBlueprintConfiguration, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironmentBlueprintConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func TestAccDataZoneEnvironmentBlueprintConfiguration_manage_access_role_arn(t *testing.T) {
	ctx := acctest.Context(t)

	var environmentblueprintconfiguration datazone.GetEnvironmentBlueprintConfigurationOutput
	domainName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_manage_access_role_arn(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
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
	domainName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_provisioning_role_arn(domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
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
	domainName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment_blueprint_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentBlueprintConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, names.USWest2RegionID, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "regional_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.%%", names.USWest2RegionID), acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.key1", names.USWest2RegionID), acctest.CtValue1),
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
				Config: testAccEnvironmentBlueprintConfigurationConfig_regional_parameters(domainName, names.USWest2RegionID, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentBlueprintConfigurationExists(ctx, resourceName, &environmentblueprintconfiguration),
					resource.TestCheckResourceAttrSet(resourceName, "environment_blueprint_id"),
					resource.TestCheckResourceAttr(resourceName, "regional_parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.%%", names.USWest2RegionID), acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("regional_parameters.%s.key2", names.USWest2RegionID), acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckEnvironmentBlueprintConfigurationExists(ctx context.Context, name string, environmentblueprintconfiguration *datazone.GetEnvironmentBlueprintConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironmentBlueprintConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironmentBlueprintConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := conn.GetEnvironmentBlueprintConfiguration(ctx, &datazone.GetEnvironmentBlueprintConfigurationInput{
			DomainIdentifier:               aws.String(rs.Primary.Attributes["domain_id"]),
			EnvironmentBlueprintIdentifier: aws.String(rs.Primary.Attributes["environment_blueprint_id"]),
		})

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironmentBlueprintConfiguration, rs.Primary.ID, err)
		}

		*environmentblueprintconfiguration = *resp

		return nil
	}
}

func testAccEnvironmentBlueprintConfigurationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
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
