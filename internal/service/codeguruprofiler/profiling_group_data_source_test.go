// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeguruprofiler_test

//import (
//	"fmt"
//	"strings"
//	"testing"
//
//	"github.com/YakDriver/regexache"
//
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler"
//	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler/types"
//	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
//	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
//	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//	"github.com/hashicorp/terraform-plugin-testing/terraform"
//	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	"github.com/hashicorp/terraform-provider-aws/internal/create"
//	tfcodeguruprofiler "github.com/hashicorp/terraform-provider-aws/internal/service/codeguruprofiler"
//	"github.com/hashicorp/terraform-provider-aws/names"
//)
//
//
//
//func TestAccCodeGuruProfilerProfilingGroupDataSource_basic(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var profilinggroup codeguruprofiler.DescribeProfilingGroupResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	dataSourceName := "data.aws_codeguruprofiler_profiling_group.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, codeguruprofiler.EndpointsID)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruProfilerEndpointID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckProfilingGroupDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccProfilingGroupDataSourceConfig_basic(rName),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckProfilingGroupExists(ctx, dataSourceName, &profilinggroup),
//					resource.TestCheckResourceAttr(dataSourceName, "auto_minor_version_upgrade", "false"),
//					resource.TestCheckResourceAttrSet(dataSourceName, "maintenance_window_start_time.0.day_of_week"),
//					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "user.*", map[string]string{
//						"console_access": "false",
//						"groups.#":       "0",
//						"username":       "Test",
//						"password":       "TestTest1234",
//					}),
//					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "codeguruprofiler", regexache.MustCompile(`profilinggroup:+.`)),
//				),
//			},
//		},
//	})
//}
//
//func testAccProfilingGroupDataSourceConfig_basic(rName, version string) string {
//	return fmt.Sprintf(`
//data "aws_security_group" "test" {
//  name = %[1]q
//}
//
//data "aws_codeguruprofiler_profiling_group" "test" {
//  profiling_group_name             = %[1]q
//  engine_type             = "ActiveCodeGuruProfiler"
//  engine_version          = %[2]q
//  host_instance_type      = "codeguruprofiler.t2.micro"
//  security_groups         = [aws_security_group.test.id]
//  authentication_strategy = "simple"
//  storage_type            = "efs"
//
//  logs {
//    general = true
//  }
//
//  user {
//    username = "Test"
//    password = "TestTest1234"
//  }
//}
//`, rName, version)
//}
