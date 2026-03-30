// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecr_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRLifecyclePolicyDocumentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecr_lifecycle_policy_document.test"

	acctest.Test(ctx, t, resource.TestCase{
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

func TestAccECRLifecyclePolicyDocumentDataSource_storageClassTransition(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecr_lifecycle_policy_document.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyDocumentDataSourceConfig_storageClassTransition,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrJSON),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrJSON, regexache.MustCompile(`"targetStorageClass":\s*"archive"`)),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrJSON, regexache.MustCompile(`"storageClass":\s*"archive"`)),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrJSON, regexache.MustCompile(`"sinceImageTransitioned"`)),
				),
			},
		},
	})
}

const testAccLifecyclePolicyDocumentDataSourceConfig_basic = `
data "aws_ecr_lifecycle_policy_document" "test" {
  rule {
    priority    = 1
    description = "This is a test"

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
    description = "This is tag pattern list test"

    selection {
      tag_status       = "tagged"
      tag_pattern_list = ["*test*1*2*3", "test*1*2*3*"]
      count_type       = "imageCountMoreThan"
      count_number     = 100
    }
  }

  rule {
    priority    = 4
    description = "Archive images not pulled for more than 30 days"

    selection {
      tag_status   = "any"
      count_type   = "sinceImagePulled"
      count_unit   = "days"
      count_number = 30
    }

    action {
      type                 = "transition"
      target_storage_class = "archive"
    }
  }

  rule {
    priority    = 5
    description = "Delete images archived for more than 365 days"

    selection {
      tag_status    = "any"
      storage_class = "archive"
      count_type    = "sinceImageTransitioned"
      count_unit    = "days"
      count_number  = 365
    }

    action {
      type = "expire"
    }
  }
}
`

const testAccLifecyclePolicyDocumentDataSourceConfig_storageClassTransition = `
data "aws_ecr_lifecycle_policy_document" "test" {
  rule {
    priority    = 1
    description = "Expire untagged images after 7 days"

    selection {
      tag_status   = "untagged"
      count_type   = "sinceImagePushed"
      count_unit   = "days"
      count_number = 7
    }

    action {
      type = "expire"
    }
  }

  rule {
    priority    = 2
    description = "Archive unused images after 90 days"

    selection {
      tag_status   = "any"
      count_type   = "sinceImagePulled"
      count_unit   = "days"
      count_number = 90
    }

    action {
      type                 = "transition"
      target_storage_class = "archive"
    }
  }

  rule {
    priority    = 3
    description = "Delete archive images after 90 days"

    selection {
      tag_status    = "any"
      count_type    = "sinceImageTransitioned"
      count_unit    = "days"
      storage_class = "archive"
      count_number  = 90
    }

    action {
      type = "expire"
    }
  }
}
`
