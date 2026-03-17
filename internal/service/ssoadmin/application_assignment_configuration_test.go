// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfssoadmin "github.com/hashicorp/terraform-provider-aws/internal/service/ssoadmin"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSOAdminApplicationAssignmentConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment_configuration.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_arn", applicationResourceName, names.AttrARN),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment_configuration.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfssoadmin.ResourceApplication, applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminApplicationAssignmentConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment_configuration.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, t, resourceName),
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
					testAccCheckApplicationAssignmentConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "assignment_required", acctest.CtFalse),
				),
			},
			{
				Config: testAccApplicationAssignmentConfigurationConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentConfigurationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "assignment_required", acctest.CtTrue),
				),
			},
		},
	})
}

func testAccCheckApplicationAssignmentConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

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

func testAccCheckApplicationAssignmentConfigurationExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignmentConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignmentConfiguration, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).SSOAdminClient(ctx)

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
  application_arn     = aws_ssoadmin_application.test.arn
  assignment_required = %[3]t
}
`, rName, testAccApplicationProviderARN, assignmentRequired)
}
