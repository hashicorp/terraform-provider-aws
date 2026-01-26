// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	policyStatement := "permit (principal, action == Action::\"view\", resource in Album:: \"test_album\");"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyStatement),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_templateLinked(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	policyStatement := "permit (principal, action == Action::\"view\", resource in Album:: \"test_album\");"
	policyStatementActionUpdated := "permit (principal, action == Action::\"write\", resource in Album:: \"test_album\");"
	policyStatementEffectUpdated := "forbid (principal, action == Action::\"view\", resource in Album:: \"test_album\");"
	policyStatementResourceUpdated := "forbid (principal, action == Action::\"view\", resource in Album:: \"test_album_updated\");"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyStatement),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatement),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
				),
			},
			{
				Config: testAccPolicyConfig_basic(rName, policyStatementActionUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatementActionUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
				),
			},
			{
				Config: testAccPolicyConfig_basic(rName, policyStatementEffectUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatementEffectUpdated),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
				),
			},
			{
				Config: testAccPolicyConfig_basic(rName, policyStatementResourceUpdated),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.description", rName),
					resource.TestCheckResourceAttr(resourceName, "definition.0.static.0.statement", policyStatementResourceUpdated),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_verifiedpermissions_policy.test"

	policyStatement := "permit (principal, action == Action::\"view\", resource in Album:: \"test_album\");"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyStatement),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &policy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfverifiedpermissions.ResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)

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

func testAccCheckPolicyExists(ctx context.Context, t *testing.T, name string, policy *verifiedpermissions.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicy, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)
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
