// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfamp "github.com/hashicorp/terraform-provider-aws/internal/service/amp"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccAMPWorkspace_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v prometheusservice.WorkspaceDescription
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, prometheusservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "alias", ""),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "prometheus_endpoint"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var v prometheusservice.WorkspaceDescription
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, prometheusservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
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

func TestAccAMPWorkspace_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v prometheusservice.WorkspaceDescription
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspaceConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkspaceConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAMPWorkspace_alias(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3, v4 prometheusservice.WorkspaceDescription
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, prometheusservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_alias(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "alias", rName1),
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
					testAccCheckVPNConnectionNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "alias", rName2),
				),
			},
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v3),
					testAccCheckVPNConnectionRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "alias", ""),
				),
			},
			{
				Config: testAccWorkspaceConfig_alias(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v4),
					testAccCheckVPNConnectionNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "alias", rName1),
				),
			},
		},
	})
}

func TestAccAMPWorkspace_loggingConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v prometheusservice.WorkspaceDescription
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_prometheus_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, prometheusservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkspaceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspaceConfig_loggingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
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
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_configuration.0.log_group_arn"),
				),
			},
			{
				Config: testAccWorkspaceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "0"),
				),
			},
			{
				Config: testAccWorkspaceConfig_loggingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkspaceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "logging_configuration.0.log_group_arn"),
				),
			},
		},
	})
}

func testAccCheckWorkspaceExists(ctx context.Context, n string, v *prometheusservice.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Prometheus Workspace ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPConn(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).AMPConn(ctx)

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

func testAccCheckVPNConnectionRecreated(before, after *prometheusservice.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.WorkspaceId), aws.StringValue(after.WorkspaceId); before == after {
			return fmt.Errorf("Expected Prometheus Workspace IDs to change, %s", before)
		}

		return nil
	}
}

func testAccCheckVPNConnectionNotRecreated(before, after *prometheusservice.WorkspaceDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.WorkspaceId), aws.StringValue(after.WorkspaceId); before != after {
			return fmt.Errorf("Expected Prometheus Workspace IDs not to change, but got before: %s, after: %s", before, after)
		}

		return nil
	}
}

func testAccWorkspaceConfig_basic() string {
	return `
resource "aws_prometheus_workspace" "test" {}
`
}

func testAccWorkspaceConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccWorkspaceConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_prometheus_workspace" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
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
