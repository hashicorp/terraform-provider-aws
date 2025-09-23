// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/datazone"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAssociateEnvironmentRole_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var plan datazone.AssociateEnvironmentRoleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_associate_environment_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccAssociateEnvironmentRolePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssociateEnvironmentRoleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociateEnvironmentRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociateEnvironmentRoleCheckExists(ctx, resourceName, &plan),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrDomainIdentifier),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEnvironmentIdentifier),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEnvironmentRoleArn),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"domain_identifier",
					"environment_identifier",
					"environment_role_arn",
				},
			},
		},
	})
}

func TestAccAssociateEnvironmentRole_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var domain datazone.AssociateEnvironmentRoleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_associate_environment_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAssociateEnvironmentRoleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssociateEnvironmentRoleCheckExists(ctx, resourceName, &domain),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceDomain, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAssociateEnvironmentRoleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_associate_environment_role" {
				continue
			}
		}

		return nil
	}
}

func testAccCheckAssociateEnvironmentRoleCheckExists(ctx context.Context, n string, v *datazone.AssociateEnvironmentRoleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		return nil
	}
}

func testAccAssociateEnvironmentRolePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

	input := &datazone.ListDomainsInput{}
	_, err := conn.ListDomains(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAssociateEnvironmentRoleConfig_base(rName string) string {
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
  name      = "CustomAwsService"
  managed   = true
}

resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id

  enabled_regions          = [data.aws_region.test.region]
}

resource "aws_datazone_environment_profile" "test" {
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.region
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  description                      = %[1]q
  name                             = %[1]q
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
}

resource "aws_datazone_environment" "test" {
  name                 = %[1]q
  description          = %[1]q
  account_identifier   = data.aws_caller_identity.test.account_id
  account_region       = data.aws_region.test.region

  profile_identifier   = aws_datazone_environment_profile.test.id
  project_identifier   = aws_datazone_project.test.id
  domain_identifier    = aws_datazone_domain.test.id
}
`, rName)
}

func testAccAssociateEnvironmentRoleConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAssociateEnvironmentRoleConfig_base(rName),
		`
resource "aws_datazone_associate_environment_role" "test" {
  domain_identifier      = aws_datazone_domain.test.id
  environment_identifier = aws_datazone_environment.test.id
  environment_role_arn = aws_iam_role.test.arn
}
`,
	)
}
