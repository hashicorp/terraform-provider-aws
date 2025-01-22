// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfverifiedpermissions "github.com/hashicorp/terraform-provider-aws/internal/service/verifiedpermissions"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsPolicyTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policytemplate verifiedpermissions.GetPolicyTemplateOutput
	resourceName := "aws_verifiedpermissions_policy_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyTemplateConfig_basic("permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource) unless { resource.IsPrivate };", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyTemplateExists(ctx, resourceName, &policytemplate),
					resource.TestCheckResourceAttr(resourceName, "statement", "permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource) unless { resource.IsPrivate };"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
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

func TestAccVerifiedPermissionsPolicyTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policytemplate verifiedpermissions.GetPolicyTemplateOutput
	resourceName := "aws_verifiedpermissions_policy_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyTemplateConfig_basic("permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource) unless { resource.IsPrivate };", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyTemplateExists(ctx, resourceName, &policytemplate),
					resource.TestCheckResourceAttr(resourceName, "statement", "permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource) unless { resource.IsPrivate };"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
				),
			},
			{
				Config: testAccPolicyTemplateConfig_basic("permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource);", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyTemplateExists(ctx, resourceName, &policytemplate),
					resource.TestCheckResourceAttr(resourceName, "statement", "permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource);"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test"),
				),
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policytemplate verifiedpermissions.GetPolicyTemplateOutput
	resourceName := "aws_verifiedpermissions_policy_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyTemplateConfig_basic(`permit (principal in ?principal, action in PhotoFlash::Action::"FullPhotoAccess", resource == ?resource) unless { resource.IsPrivate };`, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyTemplateExists(ctx, resourceName, &policytemplate),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfverifiedpermissions.ResourcePolicyTemplate, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_policy_template" {
				continue
			}
			policyStoreID, policyTemplateID, err := tfverifiedpermissions.PolicyTemplateParseID(rs.Primary.ID)
			if err != nil {
				return create.Error(names.VerifiedPermissions, create.ErrActionCheckingDestroyed, tfverifiedpermissions.ResNamePolicyTemplate, rs.Primary.ID, err)
			}

			_, err = tfverifiedpermissions.FindPolicyTemplateByID(ctx, conn, policyStoreID, policyTemplateID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingDestroyed, tfverifiedpermissions.ResNamePolicyTemplate, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPolicyTemplateExists(ctx context.Context, name string, policytemplate *verifiedpermissions.GetPolicyTemplateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyTemplate, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyTemplate, name, errors.New("not set"))
		}

		policyStoreID, policyTemplateID, err := tfverifiedpermissions.PolicyTemplateParseID(rs.Primary.ID)
		if err != nil {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyTemplate, rs.Primary.ID, err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)
		resp, err := tfverifiedpermissions.FindPolicyTemplateByID(ctx, conn, policyStoreID, policyTemplateID)

		if err != nil {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyTemplate, rs.Primary.ID, err)
		}

		*policytemplate = *resp

		return nil
	}
}

func testAccPolicyTemplateConfig_basic(statement, description string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  validation_settings {
    mode = "OFF"
  }
}

resource "aws_verifiedpermissions_policy_template" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id

  statement   = %[1]q
  description = %[2]q
}
`, statement, description)
}
