// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/datazone"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Tests need to be serialized due to `aws_lakeformation_data_lake_settings` dependency
func TestAccDataZoneEnvironment_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:            testAccDataZoneEnvironment_basic,
		acctest.CtDisappears:       testAccDataZoneEnvironment_disappears,
		"updateNameAndDescription": testAccDataZoneEnvironment_updateNameAndDescription,
		"accountIDAndRegion":       testAccDataZoneEnvironment_accountIDAndRegion,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccDataZoneEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var environment datazone.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "account_identifier"),
					resource.TestCheckResourceAttr(resourceName, "account_region", acctest.Region()),
					resource.TestCheckResourceAttrPair(resourceName, "blueprint_identifier", "aws_datazone_environment_blueprint_configuration.test", "environment_blueprint_id"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", "aws_datazone_domain.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "profile_identifier", "aws_datazone_environment_profile.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "project_identifier", "aws_datazone_project.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "provider_environment", "Amazon DataZone"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("glossary_terms"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("provisioned_resources"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("user_parameters"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEnvironmentImportStateFunc(resourceName),
			},
		},
	})
}

func testAccDataZoneEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var environment datazone.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceEnvironment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDataZoneEnvironment_updateNameAndDescription(t *testing.T) {
	ctx := acctest.Context(t)

	var environment datazone.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdate := fmt.Sprintf("%s-update", rName)
	resourceName := "aws_datazone_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_updateNameAndDescription(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName+"-description"),
				),
			},
			{
				Config: testAccEnvironmentConfig_updateNameAndDescription(rName, rNameUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rNameUpdate+"-description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEnvironmentImportStateFunc(resourceName),
			},
		},
	})
}

func testAccDataZoneEnvironment_accountIDAndRegion(t *testing.T) {
	ctx := acctest.Context(t)

	var environment datazone.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_accountIDAndRegion(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, "account_identifier"),
					resource.TestCheckResourceAttr(resourceName, "account_region", acctest.Region()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccEnvironmentImportStateFunc(resourceName),
			},
		},
	})
}

func testAccEnvironmentImportStateFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return strings.Join([]string{rs.Primary.Attributes["domain_identifier"], rs.Primary.ID}, ","), nil
	}
}

func testAccCheckEnvironmentExists(ctx context.Context, name string, environment *datazone.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := tfdatazone.FindEnvironmentByID(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.ID)

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironment, rs.Primary.ID, err)
		}

		*environment = *resp

		return nil
	}
}

func testAccCheckEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_environment" {
				continue
			}

			_, err := tfdatazone.FindEnvironmentByID(ctx, conn, rs.Primary.Attributes["domain_identifier"], rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironment, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccEnvironmentConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "datazone.amazonaws.com"
        }
      },
      {
        Action = ["sts:AssumeRole", "sts:TagSession"]
        Effect = "Allow"
        Principal = {
          Service = "cloudformation.amazonaws.com"
        }
      },
    ]
  })

  inline_policy {
    name = %[1]q
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "datazone:*",
            "ram:*",
            "sso:*",
            "kms:*",
            "glue:*",
            "lakeformation:*",
            "s3:*",
            "cloudformation:*",
            "athena:*",
            "iam:*",
            "logs:*",
          ]
          Effect   = "Allow"
          Resource = "*"
        },
      ]
    })
  }
}

data "aws_caller_identity" "test" {}
data "aws_region" "test" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.test.arn
}

resource "aws_lakeformation_data_lake_settings" "test" {
  admins = [
    data.aws_iam_session_context.current.issuer_arn,
    aws_iam_role.test.arn,
  ]
}

resource "aws_datazone_domain" "test" {
  name                  = %[1]q
  domain_execution_role = aws_iam_role.test.arn

  skip_deletion_check = true
}

resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_datazone_project" "test" {
  domain_identifier   = aws_datazone_domain.test.id
  name                = %[1]q
  description         = %[1]q
  skip_deletion_check = true
}

data "aws_datazone_environment_blueprint" "test" {
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  provisioning_role_arn    = aws_iam_role.test.arn
  manage_access_role_arn   = aws_iam_role.test.arn
  enabled_regions          = [data.aws_region.test.region]

  regional_parameters = {
    (data.aws_region.test.region) = {
      "S3Location" = "s3://${aws_s3_bucket.test.bucket}"
    }
  }
}
`, rName)
}

func testAccEnvironmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment" "test" {
  name                 = %[1]q
  blueprint_identifier = aws_datazone_environment_blueprint_configuration.test.environment_blueprint_id
  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id

  depends_on = [
    aws_lakeformation_data_lake_settings.test,
  ]
}

resource "aws_datazone_environment_profile" "test" {
  name                             = %[1]q
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.region
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
}
`, rName))
}

func testAccEnvironmentConfig_updateNameAndDescription(rName, rNameUpdated string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment" "test" {
  name                 = %[2]q
  description          = "%[2]s-description"
  blueprint_identifier = aws_datazone_environment_blueprint_configuration.test.environment_blueprint_id
  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id

  depends_on = [
    aws_lakeformation_data_lake_settings.test,
  ]
}

resource "aws_datazone_environment_profile" "test" {
  name                             = %[1]q
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.region
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
}
`, rName, rNameUpdated))
}

func testAccEnvironmentConfig_accountIDAndRegion(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment" "test" {
  name                 = %[1]q
  blueprint_identifier = aws_datazone_environment_blueprint_configuration.test.environment_blueprint_id
  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id

  account_identifier   = data.aws_caller_identity.test.account_id
  account_region       = data.aws_region.test.region

  depends_on = [
    aws_lakeformation_data_lake_settings.test,
  ]
}

resource "aws_datazone_environment_profile" "test" {
  name                             = %[1]q
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.region
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
}
`, rName))
}
