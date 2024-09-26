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

func TestAccOpenSearchServerlessLifecyclePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var lifecyclepolicy types.LifecyclePolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckLifecyclePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName, &lifecyclepolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrSet(resourceName, "policy_version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccLifecyclePolicyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessLifecyclePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var lifecyclepolicy types.LifecyclePolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckLifecyclePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName, &lifecyclepolicy),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfopensearchserverless.ResourceLifecyclePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOpenSearchServerlessLifecyclePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var lifecyclepolicy types.LifecyclePolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheckLifecyclePolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_update(rName, names.AttrDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName, &lifecyclepolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
				),
			},
			{
				Config: testAccLifecyclePolicyConfig_update(rName, "description updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, resourceName, &lifecyclepolicy),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "retention"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description updated"),
				),
			},
		},
	})
}

func testAccCheckLifecyclePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_opensearchserverless_lifecycle_policy" {
				continue
			}

			_, err := tfopensearchserverless.FindLifecyclePolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameLifecyclePolicy, rs.Primary.ID, err)
			}

			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingDestroyed, tfopensearchserverless.ResNameLifecyclePolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckLifecyclePolicyExists(ctx context.Context, name string, lifecyclepolicy *types.LifecyclePolicyDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameLifecyclePolicy, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameLifecyclePolicy, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)
		resp, err := tfopensearchserverless.FindLifecyclePolicyByNameAndType(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrType])

		if err != nil {
			return create.Error(names.OpenSearchServerless, create.ErrActionCheckingExistence, tfopensearchserverless.ResNameLifecyclePolicy, rs.Primary.ID, err)
		}

		*lifecyclepolicy = *resp

		return nil
	}
}

func testAccLifecyclePolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrType]), nil
	}
}

func testAccPreCheckLifecyclePolicy(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpenSearchServerlessClient(ctx)

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
