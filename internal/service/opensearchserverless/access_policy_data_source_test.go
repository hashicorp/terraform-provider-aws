// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessAccessPolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var accesspolicy types.AccessPolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_opensearchserverless_access_policy.test"
	resourceName := "aws_opensearchserverless_access_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessPolicyDataSourceConfig_basic(rName, "data"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessPolicyExists(ctx, dataSourceName, &accesspolicy),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy_version", resourceName, "policy_version"),
				),
			},
		},
	})
}

func testAccAccessPolicyDataSourceConfig_basic(rName, policyType string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_opensearchserverless_access_policy" "test" {
  name        = %[1]q
  type        = %[2]q
  description = %[1]q
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

data "aws_opensearchserverless_access_policy" "test" {
  name = aws_opensearchserverless_access_policy.test.name
  type = aws_opensearchserverless_access_policy.test.type
}
`, rName, policyType)
}
