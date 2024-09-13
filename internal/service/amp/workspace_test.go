// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPWorkspace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDescription
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "prometheus_endpoint"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccAMPWorkspace_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDescription
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfamp.ResourceWorkspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAMPWorkspace_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDescription
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_kms(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyARN),
				),
			},
		},
	})
}

func TestAccAMPWorkspace_alias(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4 types.WorkspaceDescription
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_alias(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_alias(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v2),
					testAccCheckWorkspaceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName2),
				),
			},
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v3),
					testAccCheckWorkspaceRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, ""),
				),
			},
			{
				Config: testAccWorkspaceConfig_alias(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v4),
					testAccCheckWorkspaceNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName1),
				),
			},
		},
	})
}

func TestAccAMPWorkspace_loggingConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_loggingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "logging_configuration.0.log_group_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_loggingConfiguration(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "logging_configuration.0.log_group_arn"),
				),
			},
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccWorkspaceConfig_loggingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "logging_configuration.0.log_group_arn"),
				),
			},
		},
	})
}

func testAccCheckWorkspaceExists(ctx context.Context, n string, v *types.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		output, err := tfamp.FindWorkspaceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkspaceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_workspace" {
				continue
			}

			_, err := tfamp.FindWorkspaceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Prometheus Workspace %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWorkspaceRecreated(i, j *types.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(i.WorkspaceId), aws.ToString(j.WorkspaceId); before == after {
			return fmt.Errorf("Prometheus Workspace (%s) not recreated", before)
		}

		return nil
	}
}

func testAccCheckWorkspaceNotRecreated(i, j *types.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(i.WorkspaceId), aws.ToString(j.WorkspaceId); before != after {
			return fmt.Errorf("Prometheus Workspace (%s) recreated", before)
		}

		return nil
	}
}

func testAccWorkspaceConfig_basic() string {
	return `
resource "aws_prometheus_workspace" "test" {}
`
}

func testAccWorkspaceConfig_alias(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias = %[1]q
}
`, rName)
}

func testAccWorkspaceConfig_loggingConfiguration(rName string, idx int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}

resource "aws_prometheus_workspace" "test" {
  logging_configuration {
    log_group_arn = "${aws_cloudwatch_log_group.test[%[2]d].arn}:*"
  }
}
`, rName, idx)
}

func testAccWorkspaceConfig_kms(rName string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  alias       = %[1]q
  kms_key_arn = aws_kms_key.test.arn
}

resource "aws_kms_key" "test" {
  description             = "Test"
  deletion_window_in_days = 7
}
`, rName)
}
