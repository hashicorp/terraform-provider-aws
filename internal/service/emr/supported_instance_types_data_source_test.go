// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRSupportedInstanceTypesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_emr_supported_instance_types.test"
	releaseLabel := "emr-6.15.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EMREndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccSupportedInstanceTypesDataSourceConfig_basic(releaseLabel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "release_label", releaseLabel),
					// Verify a known supported type is included in the output
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "supported_instance_types.*", map[string]string{
						names.AttrType: "m5.xlarge",
					}),
				),
			},
		},
	})
}

func testAccSupportedInstanceTypesDataSourceConfig_basic(releaseLabel string) string {
	return fmt.Sprintf(`
data "aws_emr_supported_instance_types" "test" {
  release_label = %[1]q
}
`, releaseLabel)
}
