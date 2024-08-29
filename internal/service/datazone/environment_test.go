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
	"github.com/aws/aws-sdk-go-v2/service/datazone/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfdatazone "github.com/hashicorp/terraform-provider-aws/internal/service/datazone"
)

func TestAccDataZoneEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var environment datazone.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	epName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_datazone_environment.test"
	envProfileName := "aws_datazone_environment_profile.test"
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
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
					resource.TestCheckResourceAttrPair(resourceName, "enviroment_account_identifier", callName, names.AttrAccountID), // fix
					resource.TestCheckResourceAttrPair(resourceName, "environment_account_region", regionName, "help"),        // fix
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),                                             // fix
					resource.TestCheckResourceAttrSet(regionName, "created_by"),
					//custom parameters
					//deployment parameters
					resource.TestCheckResourceAttrPair(resourceName, "domain_identifier", domainName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "environment_blueprint_identifier", blueName, names.AttrID),     // check this
					resource.TestCheckResourceAttrPair(resourceName, "environment_profile_identifier", envProfileName, names.AttrID), // check this
					resource.TestCheckResourceAttr(resourceName, "glossary_terms.0", "glossary_term"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					// last deployment
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "project_identifier", projectName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "provider_environment"),
					// provisioned resources
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
				),
			},
			{ResourceName: resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user"},
			},
		},
	})
}

func TestAccDataZoneEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var environment datazone.GetEnvironmentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	epName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	pName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datazone_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataZoneEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataZoneServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(rName, epName, dName, pName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnvironmentExists(ctx, resourceName, &environment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatazone.ResourceEnvironment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datazone_environment" {
				continue
			}

			_, err := tfdatazone.FindEnvironmentByID(ctx, conn, rs.Primary.Attributes["domain_identfier"], rs.Primary.Attributes[names.AttrID])
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironment, rs.Primary.ID, err)
			}

			return create.Error(names.DataZone, create.ErrActionCheckingDestroyed, tfdatazone.ResNameEnvironment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
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
		resp, err := tfdatazone.FindEnvironmentByID(ctx, conn, rs.Primary.Attributes["domain_identfier"], rs.Primary.Attributes[names.AttrID])

		if err != nil {
			return create.Error(names.DataZone, create.ErrActionCheckingExistence, tfdatazone.ResNameEnvironment, rs.Primary.ID, err)
		}

		*environment = *resp

		return nil
	}
}

func testAccEnviromentPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataZoneClient(ctx)

	input := &datazone.ListEnvironmentsInput{}
	_, err := conn.ListEnvironments(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckEnvironmentNotRecreated(before, after *datazone.GetEnvironmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if bf, after := aws.ToString(before.Id), aws.ToString(after.Id); bf != after {
			return create.Error(names.DataZone, create.ErrActionCheckingNotRecreated, tfdatazone.ResNameEnvironment, aws.ToString(before.Id), errors.New("recreated"))
		}

		return nil
	}
}

func testAccEnvironmentConfig_basic(rName, epName, dName, pName string) string {
	return acctest.ConfigCompose(testAccEnvironmentProfileConfig_basic_base(epName, dName, pName), fmt.Sprintf(`


resource "aws_datazone_environment" "test" {
	description = "desc"
	environment_account_identifier = data.aws_caller_identity.test.account_id
	environment_account_region = data.aws_region.test.name
	environment_blueprint_identifier = data.aws_datazone_environment_blueprint.test.id
	environment_profile_identifier = aws_datazone_environment_profile.test.id
	glossary_terms = ["glossary_term"]
	name = %[1]q
	project_identifier               = aws_datazone_project.test.id
	domain_identifier                = aws_datazone_domain.test.id
}
`, rName))
}
