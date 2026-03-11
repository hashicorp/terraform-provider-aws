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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessLifecyclePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var lifecyclepolicy types.LifecyclePolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_lifecycle_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckLifecyclePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, t, resourceName, &lifecyclepolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrSet(resourceName, "policy_version"),
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

func TestAccOpenSearchServerlessLifecyclePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var lifecyclepolicy types.LifecyclePolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_lifecycle_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckLifecyclePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, t, resourceName, &lifecyclepolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearchserverless.ResourceLifecyclePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessLifecyclePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var lifecyclepolicy types.LifecyclePolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_lifecycle_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckLifecyclePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_update(rName, names.AttrDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, t, resourceName, &lifecyclepolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccLifecyclePolicyConfig_update(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, t, resourceName, &lifecyclepolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
		},
	})
}

func testAccCheckLifecyclePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_lifecycle_policy" {
				continue
			}

			_, err := tfopensearchserverless.FindLifecyclePolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Serverless Lifecycle Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLifecyclePolicyExists(ctx context.Context, t *testing.T, n string, v *types.LifecyclePolicyDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		output, err := tfopensearchserverless.FindLifecyclePolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckLifecyclePolicy(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListLifecyclePoliciesInput{
		Type: types.LifecyclePolicyTypeRetention,
	}
	_, err := conn.ListLifecyclePolicies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccLifecyclePolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_lifecycle_policy" "test" {
  name = %[1]q
  type = "retention"
  policy = jsonencode({
    "Rules" : [
      {
        "ResourceType" : "index",
        "Resource" : ["index/%[1]s/*"],
        "MinIndexRetention" : "81d"
      },
      {
        "ResourceType" : "index",
        "Resource" : ["index/sales/%[1]s*"],
        "NoMinIndexRetention" : true
      }
    ]
  })
}
`, rName)
}

func testAccLifecyclePolicyConfig_update(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_lifecycle_policy" "test" {
  name        = %[1]q
  type        = "retention"
  description = %[2]q
  policy = jsonencode({
    "Rules" : [
      {
        "ResourceType" : "index",
        "Resource" : ["index/%[1]s/*"],
        "MinIndexRetention" : "81d"
      },
      {
        "ResourceType" : "index",
        "Resource" : ["index/holiday-sales/%[1]s*"],
        "NoMinIndexRetention" : true
      }
    ]
  })
}
`, rName, description)
}
