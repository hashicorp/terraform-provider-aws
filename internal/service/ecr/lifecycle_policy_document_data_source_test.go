// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRLifecyclePolicyDocumentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecr_lifecycle_policy_document.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyDocumentDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrJSON),
				),
			},
		},
	})
}

const testAccLifecyclePolicyDocumentDataSourceConfig_basic = `
data "aws_ecr_lifecycle_policy_document" "test" {
  rule {
    priority    = 1
    description = "This is a test."

    selection {
      tag_status      = "tagged"
      tag_prefix_list = ["prod"]
      count_type      = "imageCountMoreThan"
      count_number    = 100
    }
  }

  rule {
    priority    = 2
    description = "Another one"

    selection {
      tag_status      = "any"
      tag_prefix_list = ["ay", "bee"]
      count_type      = "sinceImagePushed"
      count_number    = 40
    }
  }

  rule {
    priority    = 3
    description = "This is tag pattern list test."

    selection {
      tag_status       = "tagged"
      tag_pattern_list = ["*test*1*2*3", "test*1*2*3*"]
      count_type       = "imageCountMoreThan"
      count_number     = 100
    }
  }
}
`
