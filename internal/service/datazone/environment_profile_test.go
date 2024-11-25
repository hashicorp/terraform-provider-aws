// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/datazone"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datazone/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneEnvironmentProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var environmentprofile datazone.GetEnvironmentProfileOutput
	epName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_environment_profile.test"
	domainName := "aws_datazone_domain.test"
	callName := "data.aws_caller_identity.test"
	projectName := "aws_datazone_project.test"
	regionName := "data.aws_region.test"
	blueName := "data.aws_datazone_environment_blueprint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentProfileConfig_basic(epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentProfileExists(ctx, resourceName, &environmentprofile),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAWSAccountID, callName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "aws_account_region", regionName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "environment_blueprint_identifier", blueName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, epName),
					resource.TestCheckResourceAttrPair(resourceName, "project_identifier", projectName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccAuthorizerEnvProfImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_parameters"},
			},
		},
	})
}

func TestAccDataZoneEnvironmentProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var environmentprofile datazone.GetEnvironmentProfileOutput
	epName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_environment_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentProfileConfig_basic(epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentProfileExists(ctx, resourceName, &environmentprofile),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceEnvironmentProfile, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataZoneEnvironmentProfile_update(t *testing.T) {
	ctx := acctest.Context(t)

	var environmentprofile datazone.GetEnvironmentProfileOutput
	epName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_environment_profile.test"
	domainName := "aws_datazone_domain.test"
	callName := "data.aws_caller_identity.test"
	projectName := "aws_datazone_project.test"
	regionName := "data.aws_region.test"
	blueName := "data.aws_datazone_environment_blueprint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentProfileConfig_basic(epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentProfileExists(ctx, resourceName, &environmentprofile),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAWSAccountID, callName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "aws_account_region", regionName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.name", "consumerGlueDbName"),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.value", names.AttrValue),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "environment_blueprint_identifier", blueName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, epName),
					resource.TestCheckResourceAttrPair(resourceName, "project_identifier", projectName, names.AttrID),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccAuthorizerEnvProfImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_parameters"},
			},
			{
				Config: testAccEnvironmentProfileConfig_update(epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentProfileExists(ctx, resourceName, &environmentprofile),
					testAccCheckEnvironmentProfileExists(ctx, resourceName, &environmentprofile),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrAWSAccountID, callName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(resourceName, "aws_account_region", regionName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "created_by"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.name", "consumerGlueDbName"),
					resource.TestCheckResourceAttr(resourceName, "user_parameters.0.value", names.AttrValue),
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "environment_blueprint_identifier", blueName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, epName),
					resource.TestCheckResourceAttrPair(resourceName, "project_identifier", projectName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccAuthorizerEnvProfImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_parameters"},
			},
		},
	})
}

func testAccCheckEnvironmentProfileDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_environment_profile" {
				continue
			}

			_, err := tfdatazone.FindEnvironmentProfileByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["domain_identifier"])

			if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsA[*awstypes.AccessDeniedException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironmentProfile, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironmentProfile, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckEnvironmentProfileExists(ctx context.Context, name string, environmentprofile *datazone.GetEnvironmentProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironmentProfile, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironmentProfile, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := tfdatazone.FindEnvironmentProfileByID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["domain_identifier"])

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironmentProfile, rs.Primary.ID, err)
		}

		*environmentprofile = *resp

		return nil
	}
}

func testAccAuthorizerEnvProfImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.ID, rs.Primary.Attributes["domain_identifier"]), nil
	}
}

func testAccEnvironmentProfileConfig_base(domainName, projectName string) string {
	return acctest.ConfigCompose(
		testAccProjectConfig_basic(projectName, domainName),
		`
data "aws_caller_identity" "test" {}
data "aws_region" "test" {}

data "aws_datazone_environment_blueprint" "test" {
  domain_id = aws_datazone_domain.test.id
  name      = "DefaultDataLake"
  managed   = true
}

resource "aws_datazone_environment_blueprint_configuration" "test" {
  domain_id                = aws_datazone_domain.test.id
  environment_blueprint_id = data.aws_datazone_environment_blueprint.test.id
  provisioning_role_arn    = aws_iam_role.domain_execution_role.arn
  enabled_regions          = [data.aws_region.test.name]
}
`)
}

func testAccEnvironmentProfileConfig_basic(rName, domainName, projectName string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentProfileConfig_base(domainName, projectName),
		fmt.Sprintf(`
resource "aws_datazone_environment_profile" "test" {
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.name
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  description                      = "desc"
  name                             = %[1]q
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
  user_parameters {
    name  = "consumerGlueDbName"
    value = "value"
  }
}
`, rName))
}

func testAccEnvironmentProfileConfig_update(rName, domainName, projectName string) string {
	return acctest.ConfigCompose(
		testAccEnvironmentProfileConfig_base(domainName, projectName),
		fmt.Sprintf(`
resource "aws_datazone_environment_profile" "test" {
  aws_account_id                   = data.aws_caller_identity.test.account_id
  aws_account_region               = data.aws_region.test.name
  description                      = "description"
  environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
  name                             = %[1]q
  project_identifier               = aws_datazone_project.test.id
  domain_identifier                = aws_datazone_domain.test.id
  user_parameters {
    name  = "consumerGlueDbName"
    value = "value"
  }
}
`, rName))
}
