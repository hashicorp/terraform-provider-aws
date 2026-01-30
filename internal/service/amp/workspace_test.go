// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAMPWorkspace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDescription
	resourceName := "aws_prometheus_workspace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, ""),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "aps", "workspace/{id}"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceName, names.AttrID, aws.ToString(v.WorkspaceId))(s)
					},
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyARN, ""),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "prometheus_endpoint"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfamp.ResourceWorkspace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAMPWorkspace_kms(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.WorkspaceDescription
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_kms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyARN, "aws_kms_key.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccAMPWorkspace_alias(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4 types.WorkspaceDescription
	rName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.AMPEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_alias(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v1),
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
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v2),
					testAccCheckWorkspaceNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, rName2),
				),
			},
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v3),
					testAccCheckWorkspaceRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrAlias, ""),
				),
			},
			{
				Config: testAccWorkspaceConfig_alias(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v4),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AMPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_loggingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
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
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_configuration.0.log_group_arn"),
				),
			},
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "0"),
				),
			},
			{
				Config: testAccWorkspaceConfig_loggingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_configuration.0.log_group_arn"),
				),
			},
		},
	})
}

func testAccCheckWorkspaceExists(ctx context.Context, t *testing.T, n string, v *types.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		output, err := tfamp.FindWorkspaceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckWorkspaceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AMPClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_prometheus_workspace" {
				continue
			}

			_, err := tfamp.FindWorkspaceByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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
  enable_key_rotation     = true
}
`, rName)
}
