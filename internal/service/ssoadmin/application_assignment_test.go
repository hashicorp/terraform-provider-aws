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

func TestAccSSOAdminApplicationAssignment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment.test"
	applicationResourceName := "aws_ssoadmin_application.test"
	userResourceName := "aws_identitystore_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_arn", applicationResourceName, "application_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", userResourceName, "user_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "USER"),
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

func TestAccSSOAdminApplicationAssignment_group(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment.test"
	applicationResourceName := "aws_ssoadmin_application.test"
	groupResourceName := "aws_identitystore_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfig_group(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "application_arn", applicationResourceName, "application_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "principal_id", groupResourceName, "group_id"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "GROUP"),
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

func TestAccSSOAdminApplicationAssignment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceApplicationAssignment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccSSOAdminApplicationAssignment_disappears_Application(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ssoadmin_application_assignment.test"
	applicationResourceName := "aws_ssoadmin_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSOAdminEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSOAdminServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationAssignmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationAssignmentExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfssoadmin.ResourceApplication, applicationResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationAssignmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ssoadmin_application_assignment" {
				continue
			}

			_, err := tfssoadmin.FindApplicationAssignmentByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplicationAssignment, rs.Primary.ID, err)
			}

			return create.Error(names.SSOAdmin, create.ErrActionCheckingDestroyed, tfssoadmin.ResNameApplicationAssignment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckApplicationAssignmentExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SSOAdminClient(ctx)

		_, err := tfssoadmin.FindApplicationAssignmentByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.SSOAdmin, create.ErrActionCheckingExistence, tfssoadmin.ResNameApplicationAssignment, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccApplicationAssignmentConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_application" "test" {
  name                     = %[1]q
  application_provider_arn = %[2]q
  instance_arn             = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName, testAccApplicationProviderARN)
}

func testAccApplicationAssignmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationAssignmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]

  display_name = "Acceptance Test"
  user_name    = %[1]q

  name {
    family_name = "Doe"
    given_name  = "John"
  }
}

resource "aws_ssoadmin_application_assignment" "test" {
  application_arn = aws_ssoadmin_application.test.application_arn
  principal_id    = aws_identitystore_user.test.user_id
  principal_type  = "USER"
}
`, rName))
}

func testAccApplicationAssignmentConfig_group(rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationAssignmentConfigBase(rName),
		fmt.Sprintf(`
resource "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  display_name      = %[1]q
}

resource "aws_ssoadmin_application_assignment" "test" {
  application_arn = aws_ssoadmin_application.test.application_arn
  principal_id    = aws_identitystore_group.test.group_id
  principal_type  = "GROUP"
}
`, rName))
}
