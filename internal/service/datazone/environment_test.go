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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEnvironment_basic(t *testing.T) {
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrSet(resourceName, "account_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "account_region"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "profile_identifier"),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.#", "3"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "project_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_environment"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccEnvironmentImportStateFunc(resourceName),
				ImportStateVerifyIgnore: []string{"user_parameters"},
			},
		},
	})
}

func testAccEnvironment_disappears(t *testing.T) {
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceEnvironment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccEnvironment_update(t *testing.T) {
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
				Config: testAccEnvironmentConfig_update(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrSet(resourceName, "account_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "account_region"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "blueprint_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "profile_identifier"),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.#", "3"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "project_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "provider_environment"),
				),
			},
			{
				Config: testAccEnvironmentConfig_update(rName, rNameUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rNameUpdate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rNameUpdate),
				),
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
  glossary_terms      = ["2N8w6XJCwZf"]
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
  enabled_regions          = [data.aws_region.test.name]

  regional_parameters = {
    (data.aws_region.test.name) = {
      "S3Location" = "s3://${aws_s3_bucket.test.bucket}"
    }
  }
}

resource "aws_datazone_environment_profile" "test" {
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.name
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  description                      = %[1]q
  name                             = %[1]q
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
  user_parameters {
    name  = "consumerGlueDbName"
    value = "value"
  }
}
`, rName)
}

func testAccEnvironmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment" "test" {
  name                 = %[1]q
  description          = %[1]q
  account_identifier   = data.aws_caller_identity.test.account_id
  account_region       = data.aws_region.test.name
  blueprint_identifier = aws_datazone_environment_blueprint_configuration.test.environment_blueprint_id
  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id

  user_parameters {
    name  = "consumerGlueDbName"
    value = "%[1]s-consumer"
  }

  user_parameters {
    name  = "producerGlueDbName"
    value = "%[1]s-producer"
  }

  user_parameters {
    name  = "workgroupName"
    value = "%[1]s-workgroup"
  }

  depends_on = [
    aws_lakeformation_data_lake_settings.test,
  ]
}
`, rName))
}

func testAccEnvironmentConfig_update(rName, rNameUpdated string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_datazone_environment" "test" {
  name                 = %[2]q
  description          = %[2]q
  account_identifier   = data.aws_caller_identity.test.account_id
  account_region       = data.aws_region.test.name
  blueprint_identifier = aws_datazone_environment_blueprint_configuration.test.environment_blueprint_id
  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id

  user_parameters {
    name  = "consumerGlueDbName"
    value = "%[1]s-consumer"
  }

  user_parameters {
    name  = "producerGlueDbName"
    value = "%[1]s-producer"
  }

  user_parameters {
    name  = "workgroupName"
    value = "%[1]s-workgroup"
  }

  depends_on = [
    aws_lakeformation_data_lake_settings.test,
  ]
}
`, rName, rNameUpdated))
}
