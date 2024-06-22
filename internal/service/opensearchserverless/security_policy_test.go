// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessSecurityPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var securitypolicy types.SecurityPolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, resourceName, &securitypolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "encryption"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccSecurityPolicyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var securitypolicy types.SecurityPolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_update(rName, names.AttrDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, resourceName, &securitypolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "encryption"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccSecurityPolicyConfig_update(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, resourceName, &securitypolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "encryption"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var securitypolicy types.SecurityPolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, resourceName, &securitypolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfopensearchserverless.ResourceSecurityPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecurityPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_security_policy" {
				continue
			}

			_, err := tfopensearchserverless.FindSecurityPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameSecurityPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSecurityPolicyExists(ctx context.Context, name string, securitypolicy *types.SecurityPolicyDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)
		resp, err := tfopensearchserverless.FindSecurityPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameSecurityPolicy, rs.Primary.ID, err)
		}

		*securitypolicy = *resp

		return nil
	}
}

func testAccSecurityPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrType]), nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListSecurityPoliciesInput{
		Type: types.SecurityPolicyTypeEncryption,
	}
	_, err := conn.ListSecurityPolicies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccSecurityPolicyConfig_basic(rName string) string {
	collection := fmt.Sprintf("collection/%s", rName)
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name        = %[1]q
  type        = "encryption"
  description = %[1]q
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
				%[2]q
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}
`, rName, collection)
}

func testAccSecurityPolicyConfig_update(rName, description string) string {
	collection := fmt.Sprintf("collection/%s", rName)
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name        = %[1]q
  type        = "encryption"
  description = %[3]q
  policy = jsonencode({
    "Rules" = [
      {
        "Resource" = [
				%[2]q
        ],
        "ResourceType" = "collection"
      }
    ],
    "AWSOwnedKey" = true
  })
}
`, rName, collection, description)
}
