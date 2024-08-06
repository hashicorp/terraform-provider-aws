// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminApplicationAssignmentConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment_configuration.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_arn", applicationResourceName, "application_arn"),
					resource.TestCheckResourceAttr(resourceName, "assignment_required", acctest.CtTrue),
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

func TestAccSSOAdminApplicationAssignmentConfiguration_disappears_Application(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment_configuration.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceApplication, applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminApplicationAssignmentConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "assignment_required", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "assignment_required", acctest.CtFalse),
				),
			},
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "assignment_required", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckApplicationAssignmentConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_application_assignment_configuration" {
				continue
			}

			_, err := tfssoadmin.FindApplicationAssignmentConfigurationByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplicationAssignmentConfiguration, rs.Primary.ID, err)
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplicationAssignmentConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationAssignmentConfigurationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignmentConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignmentConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err := tfssoadmin.FindApplicationAssignmentConfigurationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignmentConfiguration, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccApplicationAssignmentConfigurationConfig_basic(rName string, assignmentRequired bool) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_application_assignment_configuration" "test" {
  application_arn     = aws_ssoadmin_application.test.application_arn
  assignment_required = %[3]t
}
`, rName, testAccApplicationProviderARN, assignmentRequired)
}
