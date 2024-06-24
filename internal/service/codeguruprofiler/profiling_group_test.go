// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeguruprofiler_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/codeguruprofiler"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codeguruprofiler/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcodeguruprofiler "github.com/hashicorp/terraform-provider-aws/internal/service/codeguruprofiler"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCodeGuruProfilerProfilingGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var profilinggroup awstypes.ProfilingGroupDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccProfilingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Default"),
					resource.TestCheckResourceAttr(resourceName, "agent_orchestration_config.0.profiling_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCodeGuruProfilerProfilingGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var profilinggroup awstypes.ProfilingGroupDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccProfilingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcodeguruprofiler.ResourceProfilingGroup, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCodeGuruProfilerProfilingGroup_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var profilinggroup awstypes.ProfilingGroupDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccProfilingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Default"),
					resource.TestCheckResourceAttr(resourceName, "agent_orchestration_config.0.profiling_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccProfilingGroupConfig_update(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Default"),
					resource.TestCheckResourceAttr(resourceName, "agent_orchestration_config.0.profiling_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccCodeGuruProfilerProfilingGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var profilinggroup awstypes.ProfilingGroupDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
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
				Config: testAccProfilingGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccProfilingGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccProfilingGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProfilingGroupExists(ctx, resourceName, &profilinggroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckProfilingGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruProfilerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codeguruprofiler_profiling_group" {
				continue
			}

			_, err := tfcodeguruprofiler.FindProfilingGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingDestroyed, tfcodeguruprofiler.ResNameProfilingGroup, rs.Primary.ID, err)
			}

			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingDestroyed, tfcodeguruprofiler.ResNameProfilingGroup, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckProfilingGroupExists(ctx context.Context, name string, profilinggroup *awstypes.ProfilingGroupDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingExistence, tfcodeguruprofiler.ResNameProfilingGroup, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingExistence, tfcodeguruprofiler.ResNameProfilingGroup, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruProfilerClient(ctx)
		resp, err := tfcodeguruprofiler.FindProfilingGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.CodeGuruProfiler, create.ErrActionCheckingExistence, tfcodeguruprofiler.ResNameProfilingGroup, rs.Primary.ID, err)
		}

		*profilinggroup = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruProfilerClient(ctx)

	input := &codeguruprofiler.ListProfilingGroupsInput{}
	_, err := conn.ListProfilingGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccProfilingGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codeguruprofiler_profiling_group" "test" {
  name             = %[1]q
  compute_platform = "Default"

  agent_orchestration_config {
    profiling_enabled = true
  }
}
`, rName)
}

func testAccProfilingGroupConfig_update(rName string, profilingEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_codeguruprofiler_profiling_group" "test" {
  name             = %[1]q
  compute_platform = "Default"

  agent_orchestration_config {
    profiling_enabled = %[2]t
  }
}
`, rName, profilingEnabled)
}

func testAccProfilingGroupConfig_tags1(rName, key1, value1 string) string {
	return fmt.Sprintf(`
resource "aws_codeguruprofiler_profiling_group" "test" {
  name             = %[1]q
  compute_platform = "Default"

  agent_orchestration_config {
    profiling_enabled = true
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1)
}
func testAccProfilingGroupConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_codeguruprofiler_profiling_group" "test" {
  name             = %[1]q
  compute_platform = "Default"

  agent_orchestration_config {
    profiling_enabled = true
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2)
}
