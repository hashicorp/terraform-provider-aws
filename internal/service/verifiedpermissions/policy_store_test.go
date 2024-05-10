// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccVerifiedPermissionsPolicyStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	resourceName := "aws_verifiedpermissions_policy_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_basic("OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, resourceName, &policystore),
					resource.TestCheckResourceAttr(resourceName, "validation_settings.0.mode", "OFF"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Terraform acceptance test"),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "verifiedpermissions", regexache.MustCompile(`policy-store/+.`)),
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

func TestAccVerifiedPermissionsPolicyStore_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	resourceName := "aws_verifiedpermissions_policy_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_basic("OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, resourceName, &policystore),
					resource.TestCheckResourceAttr(resourceName, "validation_settings.0.mode", "OFF"),
				),
			},
			{
				Config: testAccPolicyStoreConfig_basic("STRICT"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "validation_settings.0.mode", "STRICT"),
				),
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStore_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	resourceName := "aws_verifiedpermissions_policy_store.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_basic("OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, resourceName, &policystore),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfverifiedpermissions.ResourcePolicyStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyStoreDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_policy_store" {
				continue
			}

			_, err := tfverifiedpermissions.FindPolicyStoreByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingDestroyed, tfverifiedpermissions.ResNamePolicyStore, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPolicyStoreExists(ctx context.Context, name string, policystore *verifiedpermissions.GetPolicyStoreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStore, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStore, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)
		resp, err := tfverifiedpermissions.FindPolicyStoreByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStore, rs.Primary.ID, err)
		}

		*policystore = *resp

		return nil
	}
}

func testAccPolicyStoresPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).VerifiedPermissionsClient(ctx)

	input := &verifiedpermissions.ListPolicyStoresInput{}
	_, err := conn.ListPolicyStores(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPolicyStoreConfig_basic(mode string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = "Terraform acceptance test"
  validation_settings {
    mode = %[1]q
  }
}`, mode)
}
