// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	awstypes "github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfworkspaces "github.com/hashicorp/terraform-provider-aws/internal/service/workspaces"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPool_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.WorkspacesPool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workspaces_pool.test"
	resourceBundleName := "data.aws_workspaces_bundle.standard"
	resourceDirectory := "aws_workspaces_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpaces),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &pool),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_settings.*", map[string]string{
						names.AttrStatus: string(awstypes.ApplicationSettingsStatusEnumDisabled),
					}),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "workspaces", regexache.MustCompile(`workspacespool/wspool-[0-9a-z]+`)),
					resource.TestCheckResourceAttrPair(resourceName, "bundle_id", resourceBundleName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "capacity.0.desired_user_sessions", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "capacity.*", map[string]string{
						"desired_user_sessions": "1",
					}),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", resourceDirectory, "directory_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "running_mode", string(awstypes.RunningModeAutoStop)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrState),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "timeout_settings.*", map[string]string{
						"disconnect_timeout_in_seconds":      "0",
						"idle_disconnect_timeout_in_seconds": "0",
						"max_user_duration_in_seconds":       "0",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "user"},
			},
		},
	})
}

func testAccPool_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pool awstypes.WorkspacesPool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workspaces_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WorkSpaces),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &pool),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfworkspaces.ResourcePool, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckPoolDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_workspaces_pool" {
				continue
			}

			_, err := tfworkspaces.FindPoolByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.WorkSpaces, create.ErrActionCheckingDestroyed, tfworkspaces.ResNamePool, rs.Primary.ID, err)
			}

			return create.Error(names.WorkSpaces, create.ErrActionCheckingDestroyed, tfworkspaces.ResNamePool, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckPoolExists(ctx context.Context, name string, pool *awstypes.WorkspacesPool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNamePool, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNamePool, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

		resp, err := tfworkspaces.FindPoolByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.WorkSpaces, create.ErrActionCheckingExistence, tfworkspaces.ResNamePool, rs.Primary.ID, err)
		}

		*pool = *resp

		return nil
	}
}

func testAccPreCheckPool(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkSpacesClient(ctx)

	input := &workspaces.DescribeWorkspacesPoolsInput{}

	_, err := conn.DescribeWorkspacesPools(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPool_ApplicationSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.WorkspacesPool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workspaces_pool.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_ApplicationSettings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "application_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "application_settings.0.settings_group", "test"),
				),
			},
		},
	})
}

func testAccPool_TimeoutSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.WorkspacesPool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workspaces_pool.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_TimeoutSettings(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "timeout_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "timeout_settings.0.disconnect_timeout_in_seconds", "2000"),
					resource.TestCheckResourceAttr(resourceName, "timeout_settings.0.idle_disconnect_timeout_in_seconds", "2000"),
					resource.TestCheckResourceAttr(resourceName, "timeout_settings.0.max_user_duration_in_seconds", "2000"),
				),
			},
		},
	})
}

func testAccPool_TimeoutSettings_MaxUserDurationInSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var pool awstypes.WorkspacesPool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_workspaces_pool.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, strings.ToLower(workspaces.ServiceID))
			testAccPreCheckPool(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, strings.ToLower(workspaces.ServiceID)),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolConfig_TimeoutSettings_MaxUserDurationInSeconds(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPoolExists(ctx, resourceName, &pool),
					resource.TestCheckResourceAttr(resourceName, "timeout_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "timeout_settings.0.max_user_duration_in_seconds", "2000"),
				),
			},
		},
	})
}

func testAccPoolConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		//lintignore:AWSAT003
		fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

data "aws_workspaces_bundle" "standard" {
  owner = "AMAZON"
  name  = "Standard with Windows 10 (Server 2022 based) (WSP)"
}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "primary" {
  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[0]
  cidr_block           = "10.0.1.0/24"

  tags = {
    Name = "%[1]s-primary"
  }
}

resource "aws_subnet" "secondary" {
  vpc_id               = aws_vpc.test.id
  availability_zone_id = local.workspaces_az_ids[1]
  cidr_block           = "10.0.2.0/24"

  tags = {
    Name = "%[1]s-secondary"
  }
}

resource "aws_workspaces_directory" "test" {
  subnet_ids                      = [aws_subnet.primary.id, aws_subnet.secondary.id]
  workspace_type                  = "POOLS"
  workspace_directory_name        = %[1]q
  workspace_directory_description = %[1]q
  user_identity_type              = "CUSTOMER_MANAGED"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccPoolConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccPoolConfig_base(rName),
		fmt.Sprintf(`
resource "aws_workspaces_pool" "test" {
  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = %[1]q
  directory_id = aws_workspaces_directory.test.directory_id
  name         = %[1]q
	running_mode = "AUTO_STOP"
}
`, rName))
}

func testAccPoolConfig_ApplicationSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccPoolConfig_base(rName),
		fmt.Sprintf(`
resource "aws_workspaces_pool" "test" {
  application_settings {
    status         = "ENABLED"
    settings_group = "test"
  }
  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = %[1]q
  directory_id = aws_workspaces_directory.test.directory_id
  name         = %[1]q
	running_mode = "AUTO_STOP"
}
`, rName))
}

func testAccPoolConfig_TimeoutSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccPoolConfig_base(rName),
		fmt.Sprintf(`
resource "aws_workspaces_pool" "test" {
  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = %[1]q
  directory_id = aws_workspaces_directory.test.directory_id
  name         = %[1]q
	running_mode = "AUTO_STOP"
  timeout_settings {
    disconnect_timeout_in_seconds      = 2000
    idle_disconnect_timeout_in_seconds = 2000
    max_user_duration_in_seconds       = 2000
  }
}
`, rName))
}

func testAccPoolConfig_TimeoutSettings_MaxUserDurationInSeconds(rName string) string {
	return acctest.ConfigCompose(
		testAccPoolConfig_base(rName),
		fmt.Sprintf(`
resource "aws_workspaces_pool" "test" {
  bundle_id = data.aws_workspaces_bundle.standard.id
  capacity {
    desired_user_sessions = 1
  }
  description  = %[1]q
  directory_id = aws_workspaces_directory.test.directory_id
  name         = %[1]q
	running_mode = "AUTO_STOP"
  timeout_settings {
    max_user_duration_in_seconds = 2000
  }
}
`, rName))
}
