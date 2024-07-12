// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMPatchBaselineDataSource_existingBaseline(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_patch_baseline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineDataSourceConfig_existing(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "approved_patches.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "approved_patches_compliance_level", "UNSPECIFIED"),
					resource.TestCheckResourceAttr(dataSourceName, "approved_patches_enable_non_security", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "approval_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "default_baseline", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDescription, "Default Patch Baseline for CentOS Provided by AWS."),
					resource.TestCheckResourceAttr(dataSourceName, "global_filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "AWS-CentOSDefaultPatchBaseline"),
					resource.TestCheckResourceAttr(dataSourceName, "rejected_patches.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "rejected_patches_action", "ALLOW_AS_DEPENDENCY"),
					resource.TestCheckResourceAttr(dataSourceName, "source.#", acctest.Ct0),
					acctest.CheckResourceAttrJMES(dataSourceName, names.AttrJSON, "ApprovedPatches|length(@)", acctest.Ct0),
					acctest.CheckResourceAttrJMESPair(dataSourceName, names.AttrJSON, "Name", dataSourceName, names.AttrName),
					acctest.CheckResourceAttrJMESPair(dataSourceName, names.AttrJSON, "Description", dataSourceName, names.AttrDescription),
					acctest.CheckResourceAttrJMESPair(dataSourceName, names.AttrJSON, "ApprovedPatchesEnableNonSecurity", dataSourceName, "approved_patches_enable_non_security"),
					acctest.CheckResourceAttrJMESPair(dataSourceName, names.AttrJSON, "OperatingSystem", dataSourceName, "operating_system"),
				),
			},
		},
	})
}

func TestAccSSMPatchBaselineDataSource_newBaseline(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_patch_baseline.test"
	resourceName := "aws_ssm_patch_baseline.test"
	rName := sdkacctest.RandomWithPrefix("tf-bl-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPatchBaselineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselineDataSourceConfig_new(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "approved_patches", resourceName, "approved_patches"),
					resource.TestCheckResourceAttrPair(dataSourceName, "approved_patches_compliance_level", resourceName, "approved_patches_compliance_level"),
					resource.TestCheckResourceAttrPair(dataSourceName, "approved_patches_enable_non_security", resourceName, "approved_patches_enable_non_security"),
					resource.TestCheckResourceAttrPair(dataSourceName, "approval_rule", resourceName, "approval_rule"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "global_filter.#", resourceName, "global_filter.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "operating_system", resourceName, "operating_system"),
					resource.TestCheckResourceAttrPair(dataSourceName, "rejected_patches", resourceName, "rejected_patches"),
					resource.TestCheckResourceAttrPair(dataSourceName, "rejected_patches_action", resourceName, "rejected_patches_action"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrSource, resourceName, names.AttrSource),
				),
			},
		},
	})
}

// Test against one of the default baselines created by AWS
func testAccPatchBaselineDataSourceConfig_existing() string {
	return `
data "aws_ssm_patch_baseline" "test" {
  owner            = "AWS"
  name_prefix      = "AWS-"
  operating_system = "CENTOS"
}
`
}

// Create a new baseline and pull it back
func testAccPatchBaselineDataSourceConfig_new(name string) string {
	return fmt.Sprintf(`
resource "aws_ssm_patch_baseline" "test" {
  name             = "%s"
  operating_system = "AMAZON_LINUX_2"
  description      = "Test"

  approval_rule {
    approve_after_days = 5
    patch_filter {
      key    = "CLASSIFICATION"
      values = ["*"]
    }
  }
}

data "aws_ssm_patch_baseline" "test" {
  owner            = "Self"
  name_prefix      = aws_ssm_patch_baseline.test.name
  operating_system = "AMAZON_LINUX_2"
}
`, name)
}
