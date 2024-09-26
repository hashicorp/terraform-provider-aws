// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeguruprofiler_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/codeguruprofiler/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeGuruProfilerProfilingGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var profilinggroup awstypes.ProfilingGroupDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeguruprofiler_profiling_group.test"
	resourceName := "aws_codeguruprofiler_profiling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruProfilerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProfilingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProfilingGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, dataSourceName, &profilinggroup),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "compute_platform", resourceName, "compute_platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, "agent_orchestration_config.0.profiling_enabled", resourceName, "agent_orchestration_config.0.profiling_enabled"),
				),
			},
		},
	})
}

func testAccProfilingGroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codeguruprofiler_profiling_group" "test" {
  name             = %[1]q
  compute_platform = "Default"

  agent_orchestration_config {
    profiling_enabled = true
  }
}

data "aws_codeguruprofiler_profiling_group" "test" {
  name = aws_codeguruprofiler_profiling_group.test.name
}
`, rName)
}
