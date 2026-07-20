// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMPatchBaselinesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_patch_baselines.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselinesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "baseline_identities.*", map[string]string{
						"baseline_name":    "AWS-UbuntuDefaultPatchBaseline",
						"default_baseline": acctest.CtTrue,
						"operating_system": "UBUNTU",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "baseline_identities.*", map[string]string{
						"baseline_name":    "AWS-WindowsPredefinedPatchBaseline-OS",
						"default_baseline": acctest.CtFalse,
						"operating_system": "WINDOWS",
					}),
				),
			},
		},
	})
}

func TestAccSSMPatchBaselinesDataSource_defaultBaselines(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_patch_baselines.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselinesDataSourceConfig_defaultBaselines(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "baseline_identities.*", map[string]string{
						"baseline_name":    "AWS-UbuntuDefaultPatchBaseline",
						"default_baseline": acctest.CtTrue,
						"operating_system": "UBUNTU",
					}),
				),
			},
		},
	})
}

func TestAccSSMPatchBaselinesDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ssm_patch_baselines.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSMEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccPatchBaselinesDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "baseline_identities.*", map[string]string{
						"baseline_name":    "AWS-DefaultPatchBaseline",
						"default_baseline": acctest.CtTrue,
						"operating_system": "WINDOWS",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "baseline_identities.*", map[string]string{
						"baseline_name":    "AWS-WindowsPredefinedPatchBaseline-OS",
						"default_baseline": acctest.CtFalse,
						"operating_system": "WINDOWS",
					}),
				),
			},
		},
	})
}

func testAccPatchBaselinesDataSourceConfig_basic() string {
	return `
data "aws_ssm_patch_baselines" "test" {}
`
}

func testAccPatchBaselinesDataSourceConfig_defaultBaselines() string {
	return `
data "aws_ssm_patch_baselines" "test" {
  default_baselines = true
}
`
}

func testAccPatchBaselinesDataSourceConfig_filter() string {
	return `
data "aws_ssm_patch_baselines" "test" {
  filter {
    key    = "OWNER"
    values = ["AWS"]
  }
  filter {
    key    = "OPERATING_SYSTEM"
    values = ["WINDOWS"]
  }
}
`
}
