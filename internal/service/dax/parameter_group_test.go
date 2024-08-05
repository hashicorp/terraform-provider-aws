// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dax_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dax"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dax/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdax "github.com/hashicorp/terraform-provider-aws/internal/service/dax"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDAXParameterGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dax_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccParameterGroupConfig_parameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccDAXParameterGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dax_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DAXServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckParameterGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckParameterGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdax.ResourceParameterGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckParameterGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dax_parameter_group" {
				continue
			}

			_, err := conn.DescribeParameterGroups(ctx, &dax.DescribeParameterGroupsInput{
				ParameterGroupNames: []string{rs.Primary.ID},
			})
			if err != nil {
				if errs.IsA[*awstypes.ParameterGroupNotFoundFault](err) {
					return nil
				}
				return err
			}
		}
		return nil
	}
}

func testAccCheckParameterGroupExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DAXClient(ctx)

		_, err := conn.DescribeParameterGroups(ctx, &dax.DescribeParameterGroupsInput{
			ParameterGroupNames: []string{rs.Primary.ID},
		})

		return err
	}
}

func testAccParameterGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dax_parameter_group" "test" {
  name = "%s"
}
`, rName)
}

func testAccParameterGroupConfig_parameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_dax_parameter_group" "test" {
  name = "%s"

  parameters {
    name  = "query-ttl-millis"
    value = "100000"
  }

  parameters {
    name  = "record-ttl-millis"
    value = "100000"
  }
}
`, rName)
}
