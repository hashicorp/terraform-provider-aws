// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeguruprofiler_test

//import (
//	"context"
//	"errors"
//	"fmt"
//	"testing"
//
//	"github.com/YakDriver/regexache"
//
//	"github.com/aws/aws-sdk-go-v2/aws"
//	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler"
//	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler/types"
//	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
//	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//	"github.com/hashicorp/terraform-plugin-testing/terraform"
//	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
//	"github.com/hashicorp/terraform-provider-aws/internal/conns"
//	"github.com/hashicorp/terraform-provider-aws/internal/create"
//	"github.com/hashicorp/terraform-provider-aws/internal/errs"
//	"github.com/hashicorp/terraform-provider-aws/names"
//	tfcodeguruprofiler "github.com/hashicorp/terraform-provider-aws/internal/service/codeguruprofiler"
//)
//
//func TestAccCodeGuruProfilerProfilingGroup_basic(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var profilinggroup codeguruprofiler.DescribeProfilingGroupResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_codeguruprofiler_profiling_group.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, names.CodeGuruProfilerEndpointID)
//			testAccPreCheck(ctx, t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruProfilerEndpointID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckProfilingGroupDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccProfilingGroupConfig_basic(rName),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
//					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
//					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
//					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
//						"console_access": "false",
//						"groups.#":       "0",
//						"username":       "Test",
//						"password":       "TestTest1234",
//					}),
//					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codeguruprofiler", regexache.MustCompile(`profilinggroup:+.`)),
//				),
//			},
//			{
//				ResourceName:            resourceName,
//				ImportState:             true,
//				ImportStateVerify:       true,
//				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
//			},
//		},
//	})
//}
//
//func TestAccCodeGuruProfilerProfilingGroup_disappears(t *testing.T) {
//	ctx := acctest.Context(t)
//	if testing.Short() {
//		t.Skip("skipping long-running test in short mode")
//	}
//
//	var profilinggroup codeguruprofiler.DescribeProfilingGroupResponse
//	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
//	resourceName := "aws_codeguruprofiler_profiling_group.test"
//
//	resource.ParallelTest(t, resource.TestCase{
//		PreCheck: func() {
//			acctest.PreCheck(ctx, t)
//			acctest.PreCheckPartitionHasService(t, names.CodeGuruProfilerEndpointID)
//			testAccPreCheck(t)
//		},
//		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruProfilerEndpointID),
//		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
//		CheckDestroy:             testAccCheckProfilingGroupDestroy(ctx),
//		Steps: []resource.TestStep{
//			{
//				Config: testAccProfilingGroupConfig_basic(rName, testAccProfilingGroupVersionNewer),
//				Check: resource.ComposeTestCheckFunc(
//					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
//					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcodeguruprofiler.ResourceProfilingGroup, resourceName),
//				),
//				ExpectNonEmptyPlan: true,
//			},
//		},
//	})
//}
//
//func testAccCheckProfilingGroupDestroy(ctx context.Context) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruProfilerClient(ctx)
//
//		for _, rs := range s.RootModule().Resources {
//			if rs.Type != "aws_codeguruprofiler_profiling_group" {
//				continue
//			}
//
//			input := &codeguruprofiler.DescribeProfilingGroupInput{
//				ProfilingGroupId: aws.String(rs.Primary.ID),
//			}
//			_, err := conn.DescribeProfilingGroup(ctx, &codeguruprofiler.DescribeProfilingGroupInput{
//				ProfilingGroupId: aws.String(rs.Primary.ID),
//			})
//			if errs.IsA[*types.ResourceNotFoundException](err){
//				return nil
//			}
//			if err != nil {
//			        return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingDestroyed, tfcodeguruprofiler.ResNameProfilingGroup, rs.Primary.ID, err)
//			}
//
//			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingDestroyed, tfcodeguruprofiler.ResNameProfilingGroup, rs.Primary.ID, errors.New("not destroyed"))
//		}
//
//		return nil
//	}
//}
//
//func testAccCheckProfilingGroupExists(ctx context.Context, name string, profilinggroup *codeguruprofiler.DescribeProfilingGroupResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		rs, ok := s.RootModule().Resources[name]
//		if !ok {
//			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingExistence, tfcodeguruprofiler.ResNameProfilingGroup, name, errors.New("not found"))
//		}
//
//		if rs.Primary.ID == "" {
//			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingExistence, tfcodeguruprofiler.ResNameProfilingGroup, name, errors.New("not set"))
//		}
//
//		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruProfilerClient(ctx)
//		resp, err := conn.DescribeProfilingGroup(ctx, &codeguruprofiler.DescribeProfilingGroupInput{
//			ProfilingGroupId: aws.String(rs.Primary.ID),
//		})
//
//		if err != nil {
//			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingExistence, tfcodeguruprofiler.ResNameProfilingGroup, rs.Primary.ID, err)
//		}
//
//		*profilinggroup = *resp
//
//		return nil
//	}
//}
//
//func testAccPreCheck(ctx context.Context, t *testing.T) {
//	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruProfilerClient(ctx)
//
//	input := &codeguruprofiler.ListProfilingGroupsInput{}
//	_, err := conn.ListProfilingGroups(ctx, input)
//
//	if acctest.PreCheckSkipError(err) {
//		t.Skipf("skipping acceptance testing: %s", err)
//	}
//	if err != nil {
//		t.Fatalf("unexpected PreCheck error: %s", err)
//	}
//}
//
//func testAccCheckProfilingGroupNotRecreated(before, after *codeguruprofiler.DescribeProfilingGroupResponse) resource.TestCheckFunc {
//	return func(s *terraform.State) error {
//		if before, after := aws.ToString(before.ProfilingGroupId), aws.ToString(after.ProfilingGroupId); before != after {
//			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingNotRecreated, tfcodeguruprofiler.ResNameProfilingGroup, aws.ToString(before.ProfilingGroupId), errors.New("recreated"))
//		}
//
//		return nil
//	}
//}
//
//func testAccProfilingGroupConfig_basic(rName, version string) string {
//	return fmt.Sprintf(`
//resource "aws_security_group" "test" {
//  name = %[1]q
//}
//
//resource "aws_codeguruprofiler_profiling_group" "test" {
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
