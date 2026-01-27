// Copyright IBM Corp. 2014, 2026
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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfverifiedpermissions "github.com/hashicorp/terraform-provider-aws/internal/service/verifiedpermissions"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVerifiedPermissionsPolicyStore_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	resourceName := "aws_verifiedpermissions_policy_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_basic("OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, resourceName, &policystore),
					resource.TestCheckResourceAttr(resourceName, "validation_settings.0.mode", "OFF"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Terraform acceptance test"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "0"),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "verifiedpermissions", regexache.MustCompile(`policy-store/.+$`)),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_basic("OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, resourceName, &policystore),
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
func TestAccVerifiedPermissionsPolicyStore_deletionProtection(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	resourceName := "aws_verifiedpermissions_policy_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_deletion_protection("DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, resourceName, &policystore),
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "DISABLED"),
				),
			},
			{
				Config: testAccPolicyStoreConfig_deletion_protection("ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "ENABLED"),
				),
			},
			{
				Config: testAccPolicyStoreConfig_deletion_protection("DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrDeletionProtection, "DISABLED"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_basic("OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, resourceName, &policystore),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfverifiedpermissions.ResourcePolicyStore, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVerifiedPermissionsPolicyStore_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var policystore verifiedpermissions.GetPolicyStoreOutput
	resourceName := "aws_verifiedpermissions_policy_store.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VerifiedPermissionsEndpointID)
			testAccPolicyStoresPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VerifiedPermissionsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyStoreDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyStoreConfig_tags1("OFF", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, resourceName, &policystore),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPolicyStoreConfig_tags2("OFF", acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, resourceName, &policystore),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccPolicyStoreConfig_tags1("OFF", acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyStoreExists(ctx, t, resourceName, &policystore),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckPolicyStoreDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_verifiedpermissions_policy_store" {
				continue
			}

			_, err := tfverifiedpermissions.FindPolicyStoreByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckPolicyStoreExists(ctx context.Context, t *testing.T, name string, policystore *verifiedpermissions.GetPolicyStoreOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStore, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStore, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)
		resp, err := tfverifiedpermissions.FindPolicyStoreByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.VerifiedPermissions, create.ErrActionCheckingExistence, tfverifiedpermissions.ResNamePolicyStore, rs.Primary.ID, err)
		}

		*policystore = *resp

		return nil
	}
}

func testAccPolicyStoresPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).VerifiedPermissionsClient(ctx)

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

func testAccPolicyStoreConfig_deletion_protection(deletionProtection string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description         = "Terraform acceptance test"
  deletion_protection = %[1]q
  validation_settings {
    mode = "OFF"
  }
}`, deletionProtection)
}

func testAccPolicyStoreConfig_tags1(mode, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = "Terraform acceptance test"
  validation_settings {
    mode = %[1]q
  }
  tags = {
    %[2]q = %[3]q
  }
}`, mode, tagKey1, tagValue1)
}

func testAccPolicyStoreConfig_tags2(mode, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_verifiedpermissions_policy_store" "test" {
  description = "Terraform acceptance test"
  validation_settings {
    mode = %[1]q
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}`, mode, tagKey1, tagValue1, tagKey2, tagValue2)
}
