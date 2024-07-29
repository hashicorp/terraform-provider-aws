// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datazone"
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var environmentprofile datazone.GetEnvironmentProfileOutput
	epName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	envProfName := "aws_datazone_environment_profile.test"
	domainName := "aws_datazone_domain.test"
	callName := "data.aws_caller_identity.test"
	projectName := "aws_datazone_project.test"
	regionName := "data.aws_region.test"
	blueName := "data.aws_datazone_environment_blueprint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentProfileConfig_basic(epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentProfileExists(ctx, envProfName, &environmentprofile),
					resource.TestCheckResourceAttrPair(envProfName, names.AttrAWSAccountID, callName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(envProfName, "aws_account_region", regionName, names.AttrName),
					resource.TestCheckResourceAttrSet(envProfName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(envProfName, "created_by"),
					resource.TestCheckResourceAttr(envProfName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttr(envProfName, "user_parameters.0.name", "consumerGlueDbName"),
					resource.TestCheckResourceAttr(envProfName, "user_parameters.0.value", names.AttrValue),
					resource.TestCheckResourceAttrPair(envProfName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(envProfName, "environment_blueprint_identifier", blueName, names.AttrID),
					resource.TestCheckResourceAttrSet(envProfName, names.AttrID),
					resource.TestCheckResourceAttr(envProfName, names.AttrName, epName),
					resource.TestCheckResourceAttrPair(envProfName, "project_identifier", projectName, names.AttrID),
				),
			},
			{
				ResourceName:            envProfName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccAuthorizerEnvProfImportStateIdFunc(envProfName),
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
			{
				Config: testAccEnvironmentProfileConfig_update(epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentProfileExists(ctx, envProfName, &environmentprofile),
					testAccCheckEnvironmentProfileExists(ctx, envProfName, &environmentprofile),
					resource.TestCheckResourceAttrPair(envProfName, names.AttrAWSAccountID, callName, names.AttrAccountID),
					resource.TestCheckResourceAttrPair(envProfName, "aws_account_region", regionName, names.AttrName),
					resource.TestCheckResourceAttrSet(envProfName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(envProfName, "created_by"),
					resource.TestCheckResourceAttr(envProfName, names.AttrDescription, names.AttrDescription),
					resource.TestCheckResourceAttr(envProfName, "user_parameters.0.name", "consumerGlueDbName"),
					resource.TestCheckResourceAttr(envProfName, "user_parameters.0.value", names.AttrValue),
					resource.TestCheckResourceAttrPair(envProfName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(envProfName, "environment_blueprint_identifier", blueName, names.AttrID),
					resource.TestCheckResourceAttrSet(envProfName, names.AttrID),
					resource.TestCheckResourceAttr(envProfName, names.AttrName, epName),
					resource.TestCheckResourceAttrPair(envProfName, "project_identifier", projectName, names.AttrID),
					resource.TestCheckResourceAttrSet(envProfName, "updated_at"),
				),
			},
			{
				ResourceName:            envProfName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccAuthorizerEnvProfImportStateIdFunc(envProfName),
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func TestAccDataZoneEnvironmentProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var environmentprofile datazone.GetEnvironmentProfileOutput
	epName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	envProfName := "aws_datazone_environment_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentProfileConfig_basic(epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentProfileExists(ctx, envProfName, &environmentprofile),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceEnvironmentProfile, envProfName),
				),
				ExpectNonEmptyPlan: true,
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

			t := rs.Primary.Attributes["domain_identifier"]

			_, err := conn.GetEnvironmentProfile(ctx, &datazone.GetEnvironmentProfileInput{
				DomainIdentifier: &t,
				Identifier:       aws.String(rs.Primary.ID),
			})

			if errs.IsA[*types.ResourceNotFoundException](err) || errs.IsA[*types.AccessDeniedException](err) {
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

		t := rs.Primary.Attributes["domain_identifier"]
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)
		resp, err := conn.GetEnvironmentProfile(ctx, &datazone.GetEnvironmentProfileInput{
			DomainIdentifier: &t,
			Identifier:       &rs.Primary.ID,
		})

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironmentProfile, rs.Primary.ID, err)
		}

		*environmentprofile = *resp

		return nil
	}
}

/*
func envProfTestAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

	input := &datazone.ListEnvironmentProfilesInput{}
	_, err := conn.ListEnvironmentProfiles(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
*/

func testAccAuthorizerEnvProfImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}


		return strings.Join([]string{rs.Primary.Attributes["domain_identifier"], rs.Primary.ID, rs.Primary.Attributes["environment_blueprint_identifier"], rs.Primary.Attributes["project_identifier"]}, ":"), nil
	}
}

func testAccEnvironmentProfileConfig_base(domainName, projectName string) string {
	return acctest.ConfigCompose(testAccProjectConfig_basic(projectName, domainName), (`
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
	`))
}

func testAccEnvironmentProfileConfig_basic(rName, domainName, projectName string) string {
	return acctest.ConfigCompose(testAccEnvironmentProfileConfig_base(domainName, projectName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccEnvironmentProfileConfig_base(domainName, projectName), fmt.Sprintf(`
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
