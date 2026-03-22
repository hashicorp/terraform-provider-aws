// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/connect"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

func testAccApprovedOrigin_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_approved_origin.test"
	origin := "https://example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "connect"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApprovedOriginDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApprovedOriginConfig_basic(rName, origin),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApprovedOriginExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", origin),
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

func testAccCheckApprovedOriginDestroy(ctx context.Context) resource.TestCheckDestroyFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_approved_origin" {
				continue
			}

			instanceID, origin, err := tfconnect.ApprovedOriginParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			paginator := connect.NewListApprovedOriginsPaginator(conn, &connect.ListApprovedOriginsInput{
				InstanceId: &instanceID,
			})
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if err != nil {
					return nil
				}
				for _, o := range page.Origins {
					if o == origin {
						return fmt.Errorf("Connect Approved Origin %s still exists", rs.Primary.ID)
					}
				}
			}
		}
		return nil
	}
}

func testAccCheckApprovedOriginExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectClient(ctx)

		instanceID, origin, err := tfconnect.ApprovedOriginParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		paginator := connect.NewListApprovedOriginsPaginator(conn, &connect.ListApprovedOriginsInput{
			InstanceId: &instanceID,
		})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return err
			}
			for _, o := range page.Origins {
				if o == origin {
					return nil
				}
			}
		}

		return fmt.Errorf("Connect Approved Origin %s not found", rs.Primary.ID)
	}
}

func testAccApprovedOriginConfig_basic(rName, origin string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}

resource "aws_connect_approved_origin" "test" {
  instance_id = aws_connect_instance.test.id
  origin      = %[2]q
}
`, rName, origin)
}
