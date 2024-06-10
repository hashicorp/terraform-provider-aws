// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappconfig "github.com/hashicorp/terraform-provider-aws/internal/service/appconfig"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppConfigDeployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment.test"
	appResourceName := "aws_appconfig_application.test"
	confProfResourceName := "aws_appconfig_configuration_profile.test"
	depStrategyResourceName := "aws_appconfig_deployment_strategy.test"
	envResourceName := "aws_appconfig_environment.test"
	confVersionResourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appconfig", regexache.MustCompile(`application/[0-9a-z]{4,7}/environment/[0-9a-z]{4,7}/deployment/1`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, appResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_profile_id", confProfResourceName, "configuration_profile_id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_version", confVersionResourceName, "version_number"),
					resource.TestCheckResourceAttr(resourceName, "deployment_number", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_strategy_id", depStrategyResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", envResourceName, "environment_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccAppConfigDeployment_kms(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment.test"
	appResourceName := "aws_appconfig_application.test"
	confProfResourceName := "aws_appconfig_configuration_profile.test"
	depStrategyResourceName := "aws_appconfig_deployment_strategy.test"
	envResourceName := "aws_appconfig_environment.test"
	confVersionResourceName := "aws_appconfig_hosted_configuration_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "appconfig", regexache.MustCompile(`application/[0-9a-z]{4,7}/environment/[0-9a-z]{4,7}/deployment/1`)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrApplicationID, appResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_profile_id", confProfResourceName, "configuration_profile_id"),
					resource.TestCheckResourceAttrPair(resourceName, "configuration_version", confVersionResourceName, "version_number"),
					resource.TestCheckResourceAttr(resourceName, "deployment_number", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_strategy_id", depStrategyResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", envResourceName, "environment_id"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
		},
	})
}

func TestAccAppConfigDeployment_predefinedStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appconfig_deployment.test"
	strategy := "AppConfig.Linear50PercentEvery30Seconds"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_predefinedStrategy(rName, strategy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "deployment_strategy_id", strategy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Since AppConfig Deployments can vary in completion times
				// depending on the predefined deployment strategy,
				// a waiter is not implemented for the resource;
				// thus, we cannot guarantee the "state" value during import.
				ImportStateVerifyIgnore: []string{names.AttrState},
			},
		},
	})
}

func TestAccAppConfigDeployment_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource1Name := "aws_appconfig_deployment.test.0"
	resource2Name := "aws_appconfig_deployment.test.1"
	resource3Name := "aws_appconfig_deployment.test.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_multiple(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resource1Name),
					testAccCheckDeploymentExists(ctx, resource2Name),
					testAccCheckDeploymentExists(ctx, resource3Name),
				),
			},
		},
	})
}

func testAccCheckDeploymentExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		appID, envID, deploymentNum, err := tfappconfig.DeploymentParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppConfigClient(ctx)

		input := &appconfig.GetDeploymentInput{
			ApplicationId:    aws.String(appID),
			DeploymentNumber: aws.Int32(deploymentNum),
			EnvironmentId:    aws.String(envID),
		}

		output, err := conn.GetDeployment(ctx, input)

		if err != nil {
			return fmt.Errorf("error getting Appconfig Deployment (%s): %w", rs.Primary.ID, err)
		}

		if output == nil {
			return fmt.Errorf("AppConfig Deployment (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDeploymentConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q
}

resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
}

resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  location_uri   = "hosted"
}

resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}

resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = %[1]q
}
`, rName)
}

func testAccDeploymentConfig_baseKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_appconfig_application" "test" {
  name = %[1]q
}

resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
}

resource "aws_appconfig_configuration_profile" "test" {
  application_id     = aws_appconfig_application.test.id
  name               = %[1]q
  location_uri       = "hosted"
  kms_key_identifier = aws_kms_key.test.arn
}

resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}

resource "aws_appconfig_hosted_configuration_version" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = %[1]q
}
`, rName)
}

func testAccDeploymentConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  description              = %[1]q
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  environment_id           = aws_appconfig_environment.test.environment_id
}
`, rName))
}

func testAccDeploymentConfig_kms(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_baseKMS(rName), fmt.Sprintf(`
resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  description              = %[1]q
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  environment_id           = aws_appconfig_environment.test.environment_id
  kms_key_identifier       = aws_kms_key.test.arn
}
`, rName))
}

func testAccDeploymentConfig_predefinedStrategy(rName, strategy string) string {
	return acctest.ConfigCompose(testAccDeploymentConfig_base(rName), fmt.Sprintf(`
resource "aws_appconfig_deployment" "test"{
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test.version_number
  description              = %[1]q
  deployment_strategy_id   = %[2]q
  environment_id           = aws_appconfig_environment.test.environment_id
}
`, rName, strategy))
}

func testAccDeploymentConfig_multiple(rName string, n int) string {
	return fmt.Sprintf(`
resource "aws_appconfig_application" "test" {
  name = %[1]q
}

resource "aws_appconfig_environment" "test" {
  name           = %[1]q
  application_id = aws_appconfig_application.test.id
}

resource "aws_appconfig_configuration_profile" "test" {
  count = %[2]d

  application_id = aws_appconfig_application.test.id
  name           = "%[1]s-${count.index}"
  location_uri   = "hosted"
}

resource "aws_appconfig_deployment_strategy" "test" {
  name                           = %[1]q
  deployment_duration_in_minutes = 3
  growth_factor                  = 10
  replicate_to                   = "NONE"
}

resource "aws_appconfig_hosted_configuration_version" "test" {
  count = %[2]d

  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test[count.index].configuration_profile_id
  content_type             = "application/json"

  content = jsonencode({
    foo = "bar"
  })

  description = "%[1]s-${count.index}"
}

resource "aws_appconfig_deployment" "test" {
  count = %[2]d

  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test[count.index].configuration_profile_id
  configuration_version    = aws_appconfig_hosted_configuration_version.test[count.index].version_number
  description              = "%[1]s-${count.index}"
  deployment_strategy_id   = aws_appconfig_deployment_strategy.test.id
  environment_id           = aws_appconfig_environment.test.environment_id
}
`, rName, n)
}
