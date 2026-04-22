// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfverifiedpermissions "github.com/hashicorp/terraform-provider-aws/internal/service/verifiedpermissions"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsSchema_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schema verifiedpermissions.GetSchemaOutput
	resourceName := "aws_verifiedpermissions_schema.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic("NAMESPACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, t, resourceName, &schema),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "NAMESPACE"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"definition.value"}, // JSON is semantically correct but can be set in state in a different order
			},
		},
	})
}

func TestAccVerifiedPermissionsSchema_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schema verifiedpermissions.GetSchemaOutput
	resourceName := "aws_verifiedpermissions_schema.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic("NAMESPACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, t, resourceName, &schema),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "NAMESPACE"),
				),
			},
			{
				Config: testAccSchemaConfig_basic("CHANGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, t, resourceName, &schema),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "CHANGED"),
				),
			},
		},
	})
}

func TestAccVerifiedPermissionsSchema_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schema verifiedpermissions.GetSchemaOutput
	resourceName := "aws_verifiedpermissions_schema.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic("NAMESPACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, t, resourceName, &schema),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfverifiedpermissions.ResourceSchema, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccVerifiedPermissionsSchema_upgrade_V6_0_0(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schema verifiedpermissions.GetSchemaOutput
	resourceName := "aws_verifiedpermissions_schema.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		CheckDestroy: testAccCheckSchemaDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.95.0",
					},
				},
				Config: testAccSchemaConfig_basic("NAMESPACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, t, resourceName, &schema),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "NAMESPACE"),
					resource.TestCheckResourceAttrSet(resourceName, "definition.value"),
				),
			},
			{
				Config:                   testAccSchemaConfig_basic("NAMESPACE"),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, t, resourceName, &schema),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "NAMESPACE"),
					resource.TestCheckResourceAttr(resourceName, "definition.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "definition.0.value"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckSchemaDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_schema" {
				continue
			}

			_, err := tfverifiedpermissions.FindSchemaByPolicyStoreID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingDestroyed, tfverifiedpermissions.ResNamePolicyStoreSchema, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSchemaExists(ctx context.Context, t *testing.T, name string, schema *verifiedpermissions.GetSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStoreSchema, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStoreSchema, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)
		resp, err := tfverifiedpermissions.FindSchemaByPolicyStoreID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStoreSchema, rs.Primary.ID, err)
		}

		*schema = *resp

		return nil
	}
}

func testAccSchemaConfig_basic(namespace string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = "Terraform acceptance test"
  validation_settings {
    mode = "STRICT"
  }
}

resource "aws_verifiedpermissions_schema" "test" {
  policy_store_id = aws_verifiedpermissions_policy_store.test.policy_store_id

  definition {
    value = "{\"%[1]s\":{\"actions\":{},\"entityTypes\":{}}}"
  }
}`, namespace)
}
