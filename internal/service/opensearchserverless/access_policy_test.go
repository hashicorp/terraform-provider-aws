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

func TestAccOpenSearchServerlessAccessPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accesspolicy types.AccessPolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_access_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckAccessPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, t, resourceName, &accesspolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "data"),
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

func TestAccOpenSearchServerlessAccessPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var accesspolicy types.AccessPolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_access_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckAccessPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_update(rName, names.AttrDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, t, resourceName, &accesspolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "data"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccAccessPolicyConfig_update(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, t, resourceName, &accesspolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "data"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
			{
				Config: testAccAccessPolicyConfig_updatePolicy(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, t, resourceName, &accesspolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "data"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
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

func TestAccOpenSearchServerlessAccessPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var accesspolicy types.AccessPolicyDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_access_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckAccessPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, t, resourceName, &accesspolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfopensearchserverless.ResourceAccessPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_access_policy" {
				continue
			}

			_, err := tfopensearchserverless.FindAccessPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("OpenSearch Serverless Access Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessPolicyExists(ctx context.Context, t *testing.T, n string, v *types.AccessPolicyDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

		output, err := tfopensearchserverless.FindAccessPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheckAccessPolicy(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).OpenSearchServerlessClient(ctx)

	input := &opensearchserverless.ListAccessPoliciesInput{
		Type: types.AccessPolicyTypeData,
	}
	_, err := conn.ListAccessPolicies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAccessPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_opensearchserverless_access_policy" "test" {
  name = %[1]q
  type = "data"
  policy = jsonencode([
    {
      "Rules" : [
        {
          "ResourceType" : "index",
          "Resource" : [
            "index/books/*"
          ],
          "Permission" : [
            "aoss:CreateIndex",
            "aoss:ReadDocument",
            "aoss:UpdateIndex",
            "aoss:DeleteIndex",
            "aoss:WriteDocument"
          ]
        }
      ],
      "Principal" : [
        "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/admin"
      ]
    }
  ])
}
`, rName)
}

func testAccAccessPolicyConfig_update(rName, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_opensearchserverless_access_policy" "test" {
  name        = %[1]q
  type        = "data"
  description = %[2]q
  policy = jsonencode([
    {
      "Rules" : [
        {
          "ResourceType" : "index",
          "Resource" : [
            "index/books/*"
          ],
          "Permission" : [
            "aoss:CreateIndex",
            "aoss:ReadDocument",
            "aoss:UpdateIndex",
            "aoss:DeleteIndex",
            "aoss:WriteDocument"
          ]
        }
      ],
      "Principal" : [
        "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/admin"
      ]
    }
  ])
}
`, rName, description)
}

func testAccAccessPolicyConfig_updatePolicy(rName, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_opensearchserverless_access_policy" "test" {
  name        = %[1]q
  type        = "data"
  description = %[2]q
  policy = jsonencode([
    {
      "Rules" : [
        {
          "ResourceType" : "index",
          "Resource" : [
            "index/books/*"
          ],
          "Permission" : [
            "aoss:CreateIndex",
            "aoss:ReadDocument",
            "aoss:UpdateIndex",
            "aoss:DeleteIndex",
            "aoss:WriteDocument"
          ]
        }
      ],
      "Principal" : [
        "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/admin",
        "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/admin2"
      ]
    }
  ])
}
`, rName, description)
}
