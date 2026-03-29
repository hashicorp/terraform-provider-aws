// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessSecurityPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var securitypolicy types.SecurityPolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, t, resourceName, &securitypolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "encryption"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: acctest.AttrsImportStateIdFunc(resourceName, "/", names.AttrName, names.AttrType),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var securitypolicy types.SecurityPolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_update(rName, names.AttrDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, t, resourceName, &securitypolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "encryption"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccSecurityPolicyConfig_update(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, t, resourceName, &securitypolicy),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, t, resourceName, &securitypolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearchserverless.ResourceSecurityPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityPolicy_string(t *testing.T) {
	ctx := acctest.Context(t)
	var securitypolicy types.SecurityPolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, t, resourceName, &securitypolicy),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, names.AttrPolicy, testAccSecurityPolicyConfig_String_ExpectedJSON(rName)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				// verify no planned changes
				Config: testAccSecurityPolicyConfig_string(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ResourceName:            resourceName,
				ImportStateIdFunc:       acctest.AttrsImportStateIdFunc(resourceName, "/", names.AttrName, names.AttrType),
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy}, // JSON is semantically correct but can be set in state in a different order
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityPolicy_stringUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var securitypolicy types.SecurityPolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityPolicyConfig_string(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, t, resourceName, &securitypolicy),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, names.AttrPolicy, testAccSecurityPolicyConfig_String_ExpectedJSON(rName)),
				),
			},
			{
				Config: testAccSecurityPolicyConfig_stringUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityPolicyExists(ctx, t, resourceName, &securitypolicy),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, names.AttrPolicy, testAccSecurityPolicyConfig_String_ExpectedJSON(rName)),
				),
			},
		},
	})
}

func testAccCheckSecurityPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_security_policy" {
				continue
			}

			_, err := tfopensearchserverless.FindSecurityPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Serverless Security Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSecurityPolicyExists(ctx context.Context, t *testing.T, n string, v *types.SecurityPolicyDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		output, err := tfopensearchserverless.FindSecurityPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

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

func testAccSecurityPolicyConfig_String_ExpectedJSON(rName string) string {
	collection := fmt.Sprintf("collection/%s", rName)
	return fmt.Sprintf(`{"Rules":[{"Resource":["%s"],"ResourceType":"collection"}],"AWSOwnedKey":true}`, collection)
}

func testAccSecurityPolicyConfig_string(rName string) string {
	collection := fmt.Sprintf("collection/%s", rName)
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name        = %[1]q
  type        = "encryption"
  description = %[1]q
  policy      = "{\"Rules\":[{\"Resource\":[\"%[2]s\"],\"ResourceType\":\"collection\"}],\"AWSOwnedKey\":true}"
}
`, rName, collection)
}

// testAccSecurityPolicyConfig_stringUpdate uses the same policy with additional whitespace
// to verify normalization prevents persistent differences
func testAccSecurityPolicyConfig_stringUpdate(rName string) string {
	collection := fmt.Sprintf("collection/%s", rName)
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_policy" "test" {
  name        = %[1]q
  type        = "encryption"
  description = "%[1]s-updated"
  policy      = "{\"Rules\":[{\"Resource\":[\"%[2]s\"],\"ResourceType\":\"collection\"}],\"AWSOwnedKey\": true }"
}
`, rName, collection)
}
