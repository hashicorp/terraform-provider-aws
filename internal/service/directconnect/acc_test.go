// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func testAccCheckVirtualInterfaceExists(ctx context.Context, name string, vif *awstypes.VirtualInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualInterfaces(ctx, &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		for _, v := range resp.VirtualInterfaces {
			if aws.ToString(v.VirtualInterfaceId) == rs.Primary.ID {
				*vif = v

				return nil
			}
		}

		return fmt.Errorf("Direct Connect virtual interface (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckVirtualInterfaceDestroy(ctx context.Context, s *terraform.State, t string) error { // nosemgrep:ci.semgrep.acctest.naming.destroy-check-signature
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectClient(ctx)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != t {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualInterfaces(ctx, &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		})
		if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "does not exist") {
			continue
		}
		if err != nil {
			return err
		}

		for _, v := range resp.VirtualInterfaces {
			if aws.ToString(v.VirtualInterfaceId) == rs.Primary.ID && v.VirtualInterfaceState != awstypes.VirtualInterfaceStateDeleted {
				return fmt.Errorf("[DESTROY ERROR] Direct Connect virtual interface (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}
