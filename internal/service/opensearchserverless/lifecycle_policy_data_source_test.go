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

func TestAccOpenSearchServerlessLifecyclePolicyDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var lifecyclepolicy types.LifecyclePolicyDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_opensearchserverless_lifecycle_policy.test"
	resourceName := "aws_opensearchserverless_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyDataSourceConfig_basic(rName, "retention"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecyclePolicyExists(ctx, dataSourceName, &lifecyclepolicy),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPolicy, resourceName, names.AttrPolicy),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy_version", resourceName, "policy_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_date"),
				),
			},
		},
	})
}

func testAccLifecyclePolicyDataSourceConfig_basic(rName, policyType string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_lifecycle_policy" "test" {
  name        = %[1]q
  type        = %[2]q
  description = %[1]q
  policy = jsonencode({
    "Rules" : [
      {
        "ResourceType" : "index",
        "Resource" : ["index/%[1]sy/*"],
        "MinIndexRetention" : "81d"
      },
      {
        "ResourceType" : "index",
        "Resource" : ["index/local-sales/%[1]s*"],
        "NoMinIndexRetention" : true
      }
    ]
  })
}

data "aws_opensearchserverless_lifecycle_policy" "test" {
  name = aws_opensearchserverless_lifecycle_policy.test.name
  type = aws_opensearchserverless_lifecycle_policy.test.type
}
`, rName, policyType)
}
