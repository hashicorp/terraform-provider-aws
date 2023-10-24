// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccessGrantsInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_grants_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessGrantsInstanceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsInstanceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantsInstanceExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_instance_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "access_grants_instance_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
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

func testAccCheckAccessGrantsInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_access_grants_instance" {
				continue
			}

			_, err := tfs3control.FindAccessGrantsInstance(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Grants Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessGrantsInstanceExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, err := tfs3control.FindAccessGrantsInstance(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccAccessGrantsInstanceConfig_basic() string {
	return `
resource "aws_s3control_access_grants_instance" "test" {}
`
}
