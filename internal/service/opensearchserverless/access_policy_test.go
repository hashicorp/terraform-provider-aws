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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfopensearchserverless "github.com/hashicorp/terraform-provider-aws/internal/service/opensearchserverless"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessAccessPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accesspolicy types.AccessPolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_access_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckAccessPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, resourceName, &accesspolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "data"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAccessPolicyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessAccessPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var accesspolicy types.AccessPolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_access_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckAccessPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_update(rName, names.AttrDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, resourceName, &accesspolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "data"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccAccessPolicyConfig_update(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, resourceName, &accesspolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "data"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessAccessPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var accesspolicy types.AccessPolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_access_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckAccessPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, resourceName, &accesspolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfopensearchserverless.ResourceAccessPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_access_policy" {
				continue
			}

			_, err := tfopensearchserverless.FindAccessPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameAccessPolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAccessPolicyExists(ctx context.Context, name string, accesspolicy *types.AccessPolicyDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameAccessPolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameAccessPolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)
		resp, err := tfopensearchserverless.FindAccessPolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameAccessPolicy, rs.Primary.ID, err)
		}

		*accesspolicy = *resp

		return nil
	}
}

func testAccAccessPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrType]), nil
	}
}

func testAccPreCheckAccessPolicy(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

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
