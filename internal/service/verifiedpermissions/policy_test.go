// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	interflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfverifiedpermissions "github.com/hashicorp/terraform-provider-aws/internal/service/verifiedpermissions"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy verifiedpermissions.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	policyStatement := "permit (principal, action == Action::\"view\", resource in Album:: \"test_album\");"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyStatement),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatement),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
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

func TestAccVerifiedPermissionsPolicy_templateLinked(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy verifiedpermissions.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_templateLinked(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttrSet(resourceName, "definition.0.template_linked.0.policy_template_id"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.template_linked.0.principal.0.entity_id", "TestUsers"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.template_linked.0.principal.0.entity_type", "User"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.template_linked.0.resource.0.entity_id", "test_album"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.template_linked.0.resource.0.entity_type", "Album"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
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

func TestAccVerifiedPermissionsPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy verifiedpermissions.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	policyStatement := "permit (principal, action == Action::\"view\", resource in Album:: \"test_album\");"
	policyStatementActionUpdated := "permit (principal, action == Action::\"write\", resource in Album:: \"test_album\");"
	policyStatementEffectUpdated := "forbid (principal, action == Action::\"view\", resource in Album:: \"test_album\");"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyStatement),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatement),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
				),
			},
			{
				Config: testAccPolicyConfig_basic(rName, policyStatementActionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatementActionUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
				),
			},
			{
				Config: testAccPolicyConfig_basic(rName, policyStatementEffectUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatementEffectUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
				),
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policy verifiedpermissions.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	policyStatement := "permit (principal, action == Action::\"view\", resource in Album:: \"test_album\");"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyStatement),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, resourceName, &policy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfverifiedpermissions.ResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_policy" {
				continue
			}

			rID, err := interflex.ExpandResourceId(rs.Primary.ID, tfverifiedpermissions.ResourcePolicyIDPartsCount, false)
			if err != nil {
				return err
			}

			_, err = tfverifiedpermissions.FindPolicyByID(ctx, conn, rID[0], rID[1])

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return nil
			}

			if err != nil {
				return create.Error(names.VerifiedPermissions, create.ErrActionCheckingDestroyed, tfverifiedpermissions.ResNamePolicy, rs.Primary.ID, err)
			}

			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingDestroyed, tfverifiedpermissions.ResNamePolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPolicyExists(ctx context.Context, name string, policy *verifiedpermissions.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)
		rID, err := interflex.ExpandResourceId(rs.Primary.ID, tfverifiedpermissions.ResourcePolicyIDPartsCount, false)
		if err != nil {
			return err
		}

		resp, err := tfverifiedpermissions.FindPolicyByID(ctx, conn, rID[0], rID[1])

		if err != nil {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicy, rs.Primary.ID, err)
		}

		*policy = *resp

		return nil
	}
}

func testAccPolicyConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = %[1]q
  validation_settings {
    mode = "OFF"
  }
}

resource "aws_verifiedpermissions_schema" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.policy_store_id

  definition {
    value = "{\"CHANGEDD\":{\"actions\":{},\"entityTypes\":{}}}"
  }
}

`, rName)
}

func testAccPolicyConfig_basic(rName, policyStatement string) string {
	return acctest.ConfigCompose(
		testAccPolicyConfig_base(rName),
		fmt.Sprintf(`
resource "aws_verifiedpermissions_policy" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id

  definition {
    static {
      description = %[1]q
      statement   = %[2]q
    }
  }
}
`, rName, policyStatement))
}

func testAccPolicyConfig_templateLinked(rName string) string {
	return acctest.ConfigCompose(
		testAccPolicyConfig_base(rName),
		fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_template" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id

  statement   = "permit (principal in ?principal, action in PhotoFlash::Action::\"FullPhotoAccess\", resource == ?resource) unless { resource.IsPrivate };"
  description = %[1]q
}

resource "aws_verifiedpermissions_policy" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.id

  definition {
    template_linked {
      policy_template_id = aws_verifiedpermissions_policy_template.test.policy_template_id

      principal {
        entity_id   = "TestUsers"
        entity_type = "User"
      }

      resource {
        entity_id   = "test_album"
        entity_type = "Album"
      }
    }
  }
}
`, rName))
}
