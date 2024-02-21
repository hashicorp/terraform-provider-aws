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

func TestAccVerifiedPermissionsSchema_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var schema verifiedpermissions.GetSchemaOutput
	resourceName := "aws_verifiedpermissions_schema.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic("NAMESPACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &schema),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic("NAMESPACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &schema),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "NAMESPACE"),
				),
			},
			{
				Config: testAccSchemaConfig_basic("CHANGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &schema),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSchemaDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSchemaConfig_basic("NAMESPACE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSchemaExists(ctx, resourceName, &schema),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfverifiedpermissions.ResourceSchema, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSchemaDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_schema" {
				continue
			}

			_, err := tfverifiedpermissions.FindSchemaByPolicyStoreID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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

func testAccCheckSchemaExists(ctx context.Context, name string, schema *verifiedpermissions.GetSchemaOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStoreSchema, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStoreSchema, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)
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
